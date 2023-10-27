# 💿 rubin - a simple record producer for kafka topics

[![GoTemplate](https://img.shields.io/badge/go/template-black?logo=go)](https://github.com/SchwarzIT/go-template)
![ci-build](https://github.com/tillkuhn/rubin/actions/workflows/main.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tillkuhn/rubin)](https://goreportcard.com/report/github.com/tillkuhn/rubin)
[![Go Reference](https://pkg.go.dev/badge/github.com/tillkuhn/rubin.svg)](https://pkg.go.dev/github.com/tillkuhn/rubin)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=tillkuhn_rubin&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=tillkuhn_rubin)
[![Latest Release](https://img.shields.io/github/v/release/tillkuhn/rubin?include_prereleases)](https://github.com/tillkuhn/rubin/releases)

## Introduction

*rubin* is basically a thin wrapper around [Confluent's  REST Proxy API (v3)](https://docs.confluent.io/platform/current/kafka-rest/api.html#records-v3) that makes it easy to produce 
event records for an existing Kafka Topic. It's written in Go and uses just plain http communication, a couple of environment variables and CLI switches.

## Why the funky name?

Initially I thought of technical names like `kafka-record-prodcer` or `topic-pusher`, but all of them turned out to be pretty boring. [Rick Rubin](https://en.wikipedia.org/wiki/Rick_Rubin) was simply the first name that showed up when I googled for "famous record producers", so I named the tool in his honour, and also in honour of the great Albums he produced in the past decades.


## Installation and usage

### Use as standalone CLI

Grap the most recent release from the [releases page](https://github.com/tillkuhn/rubin/releases)

Configure Environment

```
$ printenv | grep -e ^RUBIN
RUBIN_REST_ENDPOINT=https://localhost:443
RUBIN_API_KEY=1234567890
RUBIN_CLUSTER_ID=abc-r2d2
RUBIN_API_SECRET=********
```

Run the standalone binary

```
$ rubin -topic public.hello -record "Hello Franz!"

9:49PM	INFO	rubin/main.go:60	Welcome to rubin  {"version": "v0.0.5", "built": "now", "commit": "7759eb6"}
9:49PM	INFO	rubin/client.go:27	Kafka REST Proxy Client configured  {"endpoint": "https://localhost:443", "useSecret": true}
9:49PM	INFO	rubin/client.go:53	Push record  {"url": "https://localhost.cloud:443/kafka/v3/clusters/abc-r2d2/topics/public.hello/records"}
9:49PM	INFO	rubin/client.go:84	Record committed  {"status": "topic", "public.hello": 200, "offset": 43, "partition": 0}

```
### Use as library in external application


```
go get github.com/tillkuhn/rubin
```
```
client := rubin.New(&rubin.Options{
	RestEndpoint: "https://localhost:443",
	ClusterID:    "abc-r2d2",
	APIKey: "1234567890",
	APISecret: "**********",
})
resp, err := client.Produce(context.Background(), "toppic", "", "hey")
if err != nil {
	logger.Errorf("Cannot produce record to %s: %v", topic, err)
}
```

### Use as docker image

todo

## Development

The project uses `make` to make your life easier. If you're not familiar with Makefiles you can take a look at [this quickstart guide](https://makefiletutorial.com).

### Target Help 

Whenever you need help regarding the available actions, just use the following command.

```bash
make help
```

### Setup

To get your setup up and running the only thing you have to do is

```bash
make all
```

This will initialize a git repo, download the dependencies in the latest versions and install all needed tools.
If needed code generation will be triggered in this target as well.

### Other targets

```
$ make help
```

```
$ make help
Usage: make <OPTIONS> ... <TARGETS>

Available targets are:

all                  Initializes all tools
build                Builds all binaries
ci                   Executes lint and test and generates reports
clean                Cleans up everything
coverage             Displays coverage per func on cli
docker               Builds docker image
download             Downloads the dependencies
fmt                  Formats all code with go fmt
help                 Shows the help
html-coverage        Displays the coverage results in the browser
lint                 Lints all code with golangci-lint
run                  Run the app
run-help             Run the app and display app helm
test                 Runs all tests  (with colorized output support if gotest is installed)
test-build           Tests whether the code compiles
test-int             Run integration test with tag //go:build integration
tidy                 Cleans up go.mod and go.sum
```

### API stability

The package API for rubin is still version zero and therefore not yet considered stable as described in [gopkg.in](https://gopkg.in)

## Credits

This project was bootstrapped with [go/template](https://github.com/SchwarzIT/go-template), an efficient tool to "provides a blueprint for production-ready Go project layouts."

## Contribution
If you want to contribute to *rubin* please have a look at the [contribution guidelines](./CONTRIBUTING.md).

