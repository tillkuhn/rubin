# 💿 rubin - a simple record producer for kafka topics

[![GoTemplate](https://img.shields.io/badge/go/template-black?logo=go)](https://github.com/SchwarzIT/go-template)
![ci-build](https://github.com/tillkuhn/rubin/actions/workflows/main.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tillkuhn/rubin)](https://goreportcard.com/report/github.com/tillkuhn/rubin)
[![Go Reference](https://pkg.go.dev/badge/github.com/tillkuhn/rubin.svg)](https://pkg.go.dev/github.com/tillkuhn/rubin)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=tillkuhn_rubin&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=tillkuhn_rubin)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=tillkuhn_rubin&metric=coverage)](https://sonarcloud.io/component_measures?id=tillkuhn_rubin&metric=coverage&view=list)
[![Latest Release](https://img.shields.io/github/v/release/tillkuhn/rubin?include_prereleases)](https://github.com/tillkuhn/rubin/releases)

## Introduction

*rubin* is a thin wrapper around the [Confluent  REST Proxy API (v3)](https://docs.confluent.io/platform/current/kafka-rest/api.html#records-v3) that makes it easy to produce 
structured event records into an existing Kafka Topic. 

It's written in Go, can be used as a CLI standalone binary or as a library, and uses just plain http communication.

## Installation and usage

### Use as standalone CLI

Grap the most recent release from the [releases page](https://github.com/tillkuhn/rubin/releases) and make sure your environment is configured properly (see *help* section below for additional optional args)

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

```
```
In addition, the following CLI arguments are supported
  -ce
    	CloudEvents format for event payload (default: STRING or JSON)
  -header value
    	Header formatted as key=value, can be used multiple times
  -help
    	Display this help
  -key string
    	Kafka Message Key (optional, default is generated uuid)
  -record string
    	Request payload to send into the Kafka Topic
  -source string
    	CloudEventy: The context in which an event happened (default "rubin/cli")
  -subject string
    	CloudEventy: The subject of the event in the context of the event producer
  -topic string
    	Name of target Kafka Topic
  -type string
    	CloudEvents: Type of event related to the originating occurrence (default "event.Event")
  -v string
    	Verbosity, one of 'debug', 'info', 'warn', 'error' (default "info")
```

### Use as library in an external Go app

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

### 🐳 Use as docker image

Released vaultpal versions are build for multiple architectures and pushed to the public GitHub Container Registry (https://ghcr.io).

```
docker run --rm ghcr.io/tillkuhn/rubin:v0.1.2 -help
```

## ☁️ Support for *CloudEvents*

In Addition to simple string records and json payloads, *rubin* also supports [CloudEvents](https://cloudevents.io/), a "a specification for describing event data in common formats to provide interoperability across services, platforms and systems", hosted by the [Cloud Native Computing Foundation](https://www.cncf.io/). See the `-ce` switch and the [JSON Spec for CloudEvents](https://github.com/cloudevents/spec/blob/main/cloudevents/formats/json-format.md) for more details.

CloudEvents with CLI
```
$ rubin -topic public.hello -record '{"action":"push"}' \
    -ce -type "events.published" -source "/ci/build/123" -subject "artifact.zip"
```
CloudEvents with library
```
payload := map[string]string{"user": "james.bond", "id": "007"}
event, err = NewCloudEvent("//hr/manager", "user.created", payload)
resp, err := cc.Produce(ctx, Request{ "app.user",  event})
```
Resulting event structure (CLI example code)
```
{
  "specversion": "1.0",
  "id": "5b4eefde-bf4a-4d48-8f47-e2978df8d139",
  "source": "/ci/build/123",
  "type": "events.published",
  "subject": "artifact.zip",
  "datacontenttype": "application/json",
  "time": "2023-11-04T14:11:33Z",
  "data": {
	"action": "push"
  }
 }
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

