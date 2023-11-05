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
	ce := flag.Bool("ce", false, "CloudEvents format for event payload (default: STRING or JSON)")
	help := flag.Bool("help", false, "Display this help")
	key := flag.String("key", "", "Kafka Message Key (optional, default is generated uuid)")
	record := flag.String("record", "", "ProduceRequest payload to send into the Kafka Topic")
	source := flag.String("source", "rubin/cli", "CloudEventy: The context in which an event happened")
	subject := flag.String("subject", "", "CloudEventy: The subject of the event in the context of the event producer")
	topic := flag.String("topic", "", "Name of target Kafka Topic")
	eType := flag.String("type", "event.Event", "CloudEvents: Type of event related to the originating occurrence")
	verbosity := flag.String("v", "info", "Verbosity, one of 'debug', 'info', 'warn', 'error'")
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

	// overwrite selected options based on CLI args
	client, err := rubin.NewClientFromEnv()
	if err != nil {
		return err
	}
	client.LogLevel(*verbosity)

	if _, err := client.Produce(context.Background(), rubin.ProduceRequest{
		Topic:        *topic,
		Data:         *record,
		Key:          *key,
		Headers:      headerMap,
		AsCloudEvent: *ce,
		Source:       *source,
		Type:         *eType,
		Subject:      *subject,
	}); err != nil {
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
	fmt.Println("\nIn addition, the following CLI arguments are supported")
	flag.PrintDefaults()
	fmt.Println()
}
