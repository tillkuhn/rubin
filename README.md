# 💿 rubin - a simple record producer for kafka topics

[![GoTemplate](https://img.shields.io/badge/go/template-black?logo=go)](https://github.com/SchwarzIT/go-template)
![ci-build](https://github.com/tillkuhn/rubin/workflows/main/badge.svg)

## Introduction

A thin wrapper around [Confluent's  REST Proxy API (v3)](https://docs.confluent.io/platform/current/kafka-rest/api.html#records-v3) that makes it easy to send messages and events to a Kafka Topic.

## Installation and usage

### Use as library in external application


```
go get github.com/tillkuhn/rubin
```
```
client := rubin.New(&rubin.Options{
	RestEndpoint: "https://localhost:443",
	ClusterID:    "abc-r2d2",
	ApiKey: "1234567890",
	ApiPassword: "**********",
})
topic := "toppig"
if _, err := client.Produce(context.Background(), topic, "", "hey"); err != nil {
	logger.Errorf("Cannot produce record to %s: %v", topic, err)

```
### Use as standalone CLI

```
$ printenv | grep -e ^RUBIN
RUBIN_REST_ENDPOINT=https://localhost:443
RUBIN_API_KEY=1234567890
RUBIN_CLUSTER_ID=abc-r2d2
RUBIN_API_PASSWORD=********
```
```
$ rubin -topic public.hello -message "Hello Franz!"
3:17PM	INFO	rubin/client.go:23	Rubin  Client configured	{"endpoint": "https://localhost:443", "useSecret": true}
3:17PM	INFO	rubin/client.go:42	Push record to https://localhost:443/kafka/v3/clusters/abc-r2d2/topics/public.hello/records
3:17PM	INFO	rubin/client.go:67	Record committed	{"status": 200, "offset": 23, "topic": "public.hello"}
```

### Use as docker image

todo


## API stability

The package API for yaml v0 and therefore not yet considered stable as described in [gopkg.in](https://gopkg.in)

## Development

The project uses `make` to make your life easier. If you're not familiar with Makefiles you can take a look at [this quickstart guide](https://makefiletutorial.com).

Whenever you need help regarding the available actions, just use the following command.

```bash
make help
```

## Setup

To get your setup up and running the only thing you have to do is

```bash
make all
```

This will initialize a git repo, download the dependencies in the latest versions and install all needed tools.
If needed code generation will be triggered in this target as well.

## Test & lint

Run linting

```bash
make lint
```

Run tests

```bash
make test
```


## Why the funky name?

Initially I thought of technical names like `kafka-record-prodcer` or `topic-pusher`, but all of them turned out to be pretty boring. [Rick Rubin](https://en.wikipedia.org/wiki/Rick_Rubin) was simply the first name that showed up when I googled for "famous record producers", so I named the tool in his honour, and also in honour of the great Albums he produced in the past decades.   

## Credits

This project was bootstrapped with [go/template](https://github.com/SchwarzIT/go-template), an efficient tool to "provides a blueprint for production-ready Go project layouts."

## Contribution
If you want to contribute to go/template please have a look at the [contribution guidelines](./CONTRIBUTING.md).

