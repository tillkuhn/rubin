package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"

	"github.com/kelseyhightower/envconfig"
	"github.com/tillkuhn/rubin/pkg/rubin"
	"go.uber.org/automaxprocs/maxprocs"
)

const (
	envconfigPrefix = "kafka"
	appName         = "rubin"
)

// useful variables to pass with ldflags during build, for example
// e.g. go run -ldflags="-w -s -X 'main.version=$(shell git describe --tags --abbrev=0)' -X 'main.commit=$(shell git rev-parse --short HEAD)'"
// Default: '-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}} -X main.builtBy=goreleaser'
// see also https://goreleaser.com/cookbooks/using-main.version/

var (
	version   = "latest"
	date      = "now"
	commit    = ""
	builtBy   = "go"
	errClient = errors.New("client error") // used to wrap fine-grained errors
)

// arrayFlags based on https://stackoverflow.com/a/28323276/4292075
// How to get a list of values into a flag in Golang?
//
//	go run your_file.go --list1 value1 --list1 value2
type arrayFlags []string

func (af *arrayFlags) String() string {
	return "my string representation"
}

func (af *arrayFlags) Set(value string) error {
	*af = append(*af, value)
	return nil
}

func main() {
	// Disable automaxprocs log see https://github.com/uber-go/automaxprocs/issues/18
	nopLog := func(string, ...interface{}) {}
	_, _ = maxprocs.Set(maxprocs.Logger(nopLog))

	fmt.Printf("Welcome to %s %s built %s by %s (%s)\n\n", appName, version, date, builtBy, commit)
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// Parse cli args, 	skip if !flag.Parsed() check
	ce := flag.Bool("ce", false, "Use cloudevents format (default: STRING or JSON)")
	help := flag.Bool("help", false, "Display help")
	key := flag.String("key", "", "Key for the message (optional, default is generated uuid)")
	record := flag.String("record", "", "Record to send to the topic")
	source := flag.String("source", "rubin/cli", "Identifies the context in which an event happened (CE)")
	subject := flag.String("subject", "", "Describes the subject of the event in the context of the event producer")
	topic := flag.String("topic", "", "Kafka topic name to push records")
	eType := flag.String("type", "com.github.cloudevents.Event", "Describes the type of event related to the originating occurrence")
	verbosity := flag.String("v", "info", "Verbosity")
	var headers arrayFlags
	flag.Var(&headers, "header", "Header formatted as key=value, can be used multiple times")

	// nice: we can also use flags for maps https://www.emmanuelgautier.com/blog/string-map-command-argument-go
	flag.Parse()
	if *help {
		showHelp()
		return nil
	}

	// Validate mandatory args
	if strings.TrimSpace(*topic) == "" {
		return errors.Wrap(errClient, "kafka topic must not be empty")
	}
	// Validate mandatory args
	if strings.TrimSpace(*record) == "" {
		return errors.Wrap(errClient, "message record must not be empty")
	}
	headerMap := make(map[string]string)
	minParts := 2 // golangci treats 2 as a magic number
	for _, h := range headers {
		parts := strings.Split(h, "=")
		if len(parts) < minParts {
			continue
		}
		headerMap[parts[0]] = parts[1]
	}
	// fmt.Printf("%v map %v", headers, headerMap)
	var payloadData interface{}
	if *ce {
		event, err := rubin.NewCloudEvent(*source, *eType, *subject, *record)
		if err != nil {
			return err
		}
		payloadData = event
	} else {
		payloadData = *record
	}

	options, err := rubin.NewOptionsFromEnvconfig()
	if err != nil {
		return err
	}

	// overwrite selected options based on CLI args
	options.LogLevel = *verbosity
	client := rubin.NewClient(options)
	if _, err := client.Produce(context.Background(), *topic, *key, payloadData, headerMap); err != nil {
		return err
	}
	return nil
}

func showHelp() {
	usagePadding := 4
	tabs := tabwriter.NewWriter(os.Stdout, 1, 0, usagePadding, ' ', 0)
	_ = envconfig.Usagef(envconfigPrefix, &rubin.Options{}, tabs, envconfig.DefaultTableFormat)
	_ = tabs.Flush()
	// use below approach instead of flag.Usage() for customized output: https://stackoverflow.com/a/23726033/4292075
	fmt.Println("\nIn addition,the following CLI arguments")
	flag.PrintDefaults()
	fmt.Println()
}
