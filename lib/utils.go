package lib

import (
	"math/rand"
	"os/user"
	"time"
)

func randomMinute() int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(59)
}

func CheckEnv() string {
	_, err := user.Lookup("bighead")
	if err != nil {
		return "dev"
	}
	return "production"
}
