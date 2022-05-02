# CronitorCLI Modified Version

Please attention, this version isn't compatible to official cronitor.io , but to https://github.com/meoww-bot/cronitor-server self-hosted version.

### before you run 

```
cp .env.example .env
```

Set API_HOST in `.env`

### What I Modified

- Add several additional fields for cron job `queue`() when running `cronitor discover`:
  - Queue (return which hive queue it use if crontab job script file is related to hive, or "")
  - CommandToRun (the whole line of cronjob)
  - RunAs (which user run this cronjob)
  - Host 
- Improve `createTags` by parsing crontab job script file, now support(see `CheckMap` in `cmd/discover.go` for more details):
  - hive
  - oracle
  - snowball (clickhouse)
  - hdfs
  - hadoop
- Set API_HOST via `.env` file


**Command line tools for Cronitor.io**

CronitorCLI is the recommended companion application to the Cronitor monitoring service.  Use it on your workstation and deploy it to your server for powerful features, including:

* Import and sync all of your cron jobs
* Rich integration with Cronitor
* Power tools for your cron jobs

### Installation
CronitorCLI is packaged as a single executable for Linux, MacOS and Windows. There is no installation program, all you need to do is download and decompress the app into a location of your choice for easy system-wide use.

For the latest installation details, see https://cronitor.io/docs/using-cronitor-cli#installation

### Usage

```
CronitorCLI version 28.8

Command line tools for Cronitor.io. See https://cronitor.io/docs/using-cronitor-cli for details.

Usage:
  cronitor [command]

Available Commands:
  activity    View monitor activity
  configure   Save configuration variables to the config file
  discover    Attach monitoring to new cron jobs and watch for schedule updates
  exec        Execute a command with monitoring
  help        Help about any command
  list        Search for and list all cron jobs
  ping        Send a single ping to the selected monitoring endpoint
  select      Select a cron job to run interactively
  shell       Run commands from a cron-like shell
  status      View monitor status
  update      Update to the latest version

Flags:
  -k, --api-key string        Cronitor API Key
  -c, --config string         Config file
  -h, --help                  help for cronitor
  -n, --hostname string       A unique identifier for this host (default: system hostname)
  -l, --log string            Write debug logs to supplied file
  -p, --ping-api-key string   Ping API Key
  -v, --verbose               Verbose output

Use "cronitor [command] --help" for more information about a command.
```
