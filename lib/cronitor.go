package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type RuleValue string

type Rule struct {
	RuleType     string    `json:"rule_type"`
	Value        RuleValue `json:"value"`
	TimeUnit     string    `json:"time_unit,omitempty"`
	GraceSeconds uint      `json:"grace_seconds,omitempty"`
}

type Monitor struct {
	Name             string              `json:"name,omitempty"`
	DefaultName      string              `json:"defaultName"`
	Host             string              `json:"host"`
	CommandToRun     string              `json:"commandToRun"`
	RunAs            string              `json:"runAs"`
	Key              string              `json:"key"`
	Rules            []Rule              `json:"rules"`
	Tags             []string            `json:"tags"`
	Queue            string              `json:"queue"`
	Type             string              `json:"type"`
	Code             string              `json:"code,omitempty"`
	Timezone         string              `json:"timezone,omitempty"`
	Note             string              `json:"defaultNote,omitempty"`
	Notifications    map[string][]string `json:"notifications,omitempty"`
	NoStdoutPassthru bool                `json:"-"`
}

type MonitorSummary struct {
	Name        string `json:"name,omitempty"`
	DefaultName string `json:"defaultName"`
	Key         string `json:"key"`
	Code        string `json:"code,omitempty"`
}

type CronitorApi struct {
	IsDev          bool
	IsAutoDiscover bool
	ApiKey         string
	UserAgent      string
	Logger         func(string)
}

func (fi *RuleValue) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		return json.Unmarshal(b, (*string)(fi))
	}

	var i int
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}
	s := strconv.Itoa(i)

	*fi = RuleValue(s)
	return nil
}

func (api CronitorApi) PutMonitors(monitors map[string]*Monitor) (map[string]*Monitor, error) {
	url := api.Url()
	if api.IsAutoDiscover {
		url = url + "?auto-discover=1"
	}

	monitorsArray := make([]Monitor, 0, len(monitors))
	for _, v := range monitors {
		monitorsArray = append(monitorsArray, *v)
	}

	jsonBytes, _ := json.Marshal(monitorsArray)
	jsonString := string(jsonBytes)

	buf := new(bytes.Buffer)
	json.Indent(buf, jsonBytes, "", "  ")
	api.Logger("\nRequest:")
	api.Logger(buf.String() + "\n")

	response, err := api.sendHttpPut(url, jsonString)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Request to %s failed: %s", url, err))
	}

	buf.Truncate(0)
	json.Indent(buf, response, "", "  ")
	api.Logger("\nResponse:")
	api.Logger(buf.String() + "\n")

	responseMonitors := []Monitor{}
	if err = json.Unmarshal(response, &responseMonitors); err != nil {
		return nil, errors.New(fmt.Sprintf("Error from %s: %s", url, response))
	}

	for _, value := range responseMonitors {
		// We only need to update the Monitor struct with a code if this is a new monitor.
		// For updates the monitor code is sent as well as the key and that takes precedence.
		if _, ok := monitors[value.Key]; ok {
			monitors[value.Key].Code = value.Code
		}

	}

	return monitors, nil
}

func (api CronitorApi) GetMonitors() ([]MonitorSummary, error) {
	url := api.Url()
	page := 1
	monitors := []MonitorSummary{}

	for {
		response, err := api.sendHttpGet(fmt.Sprintf("%s?page=%d", url, page))
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Request to %s failed: %s", url, err))
		}

		type ExpectedResponse struct {
			TotalMonitorCount int              `json:"total_monitor_count"`
			PageSize          int              `json:"page_size"`
			Monitors          []MonitorSummary `json:"monitors"`
		}

		responseMonitors := ExpectedResponse{}
		if err = json.Unmarshal(response, &responseMonitors); err != nil {
			return nil, errors.New(fmt.Sprintf("Error from %s: %s", url, err.Error()))
		}

		monitors = append(monitors, responseMonitors.Monitors...)
		if page*responseMonitors.PageSize >= responseMonitors.TotalMonitorCount {
			break
		}

		page += 1
	}

	return monitors, nil
}

func (api CronitorApi) GetRawResponse(url string) ([]byte, error) {
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	request.SetBasicAuth(viper.GetString(api.ApiKey), "")
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("User-Agent", api.UserAgent)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Unexpected %d API response", response.StatusCode))
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return nil, err
	}

	return contents, nil
}

func (api CronitorApi) Url() string {

	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	API_HOST := os.Getenv("API_HOST")

	return API_HOST + "/v3/monitors"
}

func (api CronitorApi) sendHttpPut(url string, body string) ([]byte, error) {
	client := &http.Client{
		Timeout: 120 * time.Second,
	}
	request, err := http.NewRequest("PUT", url, strings.NewReader(body))
	request.SetBasicAuth(viper.GetString(api.ApiKey), "")
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("User-Agent", api.UserAgent)
	request.ContentLength = int64(len(body))
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return nil, err
	}

	return contents, nil
}

func (api CronitorApi) sendHttpGet(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 120 * time.Second,
	}
	request, err := http.NewRequest("GET", url, nil)
	request.SetBasicAuth(viper.GetString(api.ApiKey), "")
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("User-Agent", api.UserAgent)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return nil, err
	}

	return contents, nil
}
