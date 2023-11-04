# 💿 rubin - a simple record producer for kafka topics

[![GoTemplate](https://img.shields.io/badge/go/template-black?logo=go)](https://github.com/SchwarzIT/go-template)
![ci-build](https://github.com/tillkuhn/rubin/actions/workflows/main.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tillkuhn/rubin)](https://goreportcard.com/report/github.com/tillkuhn/rubin)
[![Go Reference](https://pkg.go.dev/badge/github.com/tillkuhn/rubin.svg)](https://pkg.go.dev/github.com/tillkuhn/rubin)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=tillkuhn_rubin&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=tillkuhn_rubin)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=tillkuhn_rubin&metric=coverage)](https://sonarcloud.io/component_measures?id=tillkuhn_rubin&metric=coverage&view=list)
[![Latest Release](https://img.shields.io/github/v/release/tillkuhn/rubin?include_prereleases)](https://github.com/tillkuhn/rubin/releases)

## Introduction

*rubin* is basically a thin wrapper around the [Confluent  REST Proxy API (v3)](https://docs.confluent.io/platform/current/kafka-rest/api.html#records-v3) that makes it easy to produce 
event records for an existing Kafka Topic. It's written in Go and uses just plain http communication, a couple of environment variables and CLI switches.

## Installation and usage

### Use as standalone CLI

Grap the most recent release from the [releases page](https://github.com/tillkuhn/rubin/releases)

Configure Environment (see *help* section below for optional args)

```
$ printenv | grep -e ^KAFKA

KAFKA_REST_ENDPOINT=https://localhost:443
KAFKA_PRODUCER_API_KEY=1234567890
KAFKA_CLUSTER_ID=abc-r2d2
KAFKA_PRODUCER_API_SECRET=********
```

Run the standalone binary (simple example with string message)

```
$ rubin -topic public.hello -record "Dragonfly out in the sun you know what I mean"

9:49PM	INFO	rubin/main.go:60	Welcome to rubin  {"version": "v0.0.5", "built": "now", "commit": "7759eb6"}
9:49PM	INFO	rubin/client.go:27	Kafka REST Proxy Client configured  {"endpoint": "https://localhost:443", "useSecret": true}
9:49PM	INFO	rubin/client.go:53	Push record  {"url": "https://localhost.cloud:443/kafka/v3/clusters/abc-r2d2/topics/public.hello/records"}
9:49PM	INFO	rubin/client.go:84	Record committed  {"status": "topic", "public.hello": 200, "offset": 43, "partition": 0}
```

Example with JSON Payload, custom message key and multiple message headers

```
$ rubin -topic public.hello -record '{"msg":"hello"}' -key "123" -header "source=ci" -header "mood=good"
```

Get Help on CLI arguments and environment configuration

```
$ rubin -help

Welcome to rubin v0.1.5 built now by go (b368bf8)

This application is configured via the environment. The following environment
variables can be used:

KEY                          TYPE             DEFAULT    REQUIRED    DESCRIPTION
KAFKA_REST_ENDPOINT          String                      false       Kafka REST Proxy Endpoint
KAFKA_CLUSTER_ID             String                      false       Kafka Cluster ID
KAFKA_PRODUCER_API_KEY       String                      false       Kafka API Key with Producer Privileges
KAFKA_PRODUCER_API_SECRET    String                      false       Kafka API Secret with Producer Privileges
KAFKA_HTTP_TIMEOUT           Duration         10s        false       Timeout for HTTP Client
KAFKA_DUMP_MESSAGES          True or False    false      false       Print http request/response to stdout
KAFKA_LOG_LEVEL              String           info       false       Min LogLevel debug,info,warn,error

In addition,the following CLI arguments
  -ce
    	Use cloudevents format (default: STRING or JSON)
  -header value
    	Header formatted as key=value, can be used multiple times
  -help
    	Display help
  -key string
    	Key for the message (optional, default is generated uuid)
  -record string
    	Record to send to the topic
  -source string
    	Identifies the context in which an event happened (CE) (default "rubin/cli")
  -subject string
    	Describes the subject of the event in the context of the event producer
  -topic string
    	Kafka topic name to push records
  -type string
    	Describes the type of event related to the originating occurrence (default "com.github.cloudevents.Event")
  -v string
    	Verbosity (default "info")
```

### Use as library in external application

```
go get github.com/tillkuhn/rubin
```
```
client := rubin.New(&rubin.Options{
	RestEndpoint:      "https://localhost:443",
	ClusterID:         "abc-r2d2",
	ProducerAPIKey:    "1234567890",
	ProducerAPISecret: "**********",
})
resp, err := cc.Produce(context.Background(), Record{
	Topic:   "public.hello",
	Data:    "Dragonfly out in the sun you know what I mean",
	Key:     "134-5678",
	Headers: map[string]string{"heading": "for tomorrow"},
})
fmt.Printf("Record successfully commited, offset=%d partition=%d\n", resp.Offset, resp.PartitionId)
```

### Use as docker image

Released vaultpal versions are build for multiple architectures and pushed to the public GitHub Container Registry (https://ghcr.io).

```
docker run --rm ghcr.io/tillkuhn/rubin:v0.1.2 -help
```

## Why the funky name?

Initially I thought of technical names like `kafka-record-prodcer` or `topic-pusher`, but all of them turned out to be pretty boring. [Rick Rubin](https://en.wikipedia.org/wiki/Rick_Rubin) was simply the first name that showed up when I googled for "famous record producers", so I named the tool in his honour, and also in honour of the great Albums he produced in the past decades.

<a href="https://de.wikipedia.org/wiki/Rick_Rubin">
 <img src="https://upload.wikimedia.org/wikipedia/commons/archive/4/43/20210617192624%21RickRubinSept09.jpg" >
</a>
  
*"Rick Rubin in September 2006" by jasontheexploder is licensed under [CC BY 3.0](https://creativecommons.org/licenses/by/3.0/)*


## Development

The project uses `make` to make your life easier. If you're not familiar with Makefiles you can take a look at [this quickstart guide](https://makefiletutorial.com).

Whenever you need help regarding the available actions, just use the following command.

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

## API stability

The package API for rubin is still version zero and therefore not yet considered stable as described in [gopkg.in](https://gopkg.in)

## Credits

This project was bootstrapped with [go/template](https://github.com/SchwarzIT/go-template), an efficient tool to "provides a blueprint for production-ready Go project layouts."

## Contribution
If you want to contribute to *rubin* please have a look at the [contribution guidelines](./CONTRIBUTING.md).

