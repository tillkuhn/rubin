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
// go run -ldflags="-w -s -X 'main.version=$(shell git describe --tags --abbrev=0)' -X 'main.commit=$(shell git rev-parse --short HEAD)'"
var (
	version = "latest"
	date    = "now"
	commit  = ""

	errClient = errors.New("client error") // used to wrap fine-grained errors
)

func main() {
	// Disable automaxprocs log see https://github.com/uber-go/automaxprocs/issues/18
	nopLog := func(string, ...interface{}) {}
	_, _ = maxprocs.Set(maxprocs.Logger(nopLog))

	fmt.Printf("Welcome to %s %s built %s (%s)\n\n", appName, version, date, commit)
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// Parse cli args
	topic := flag.String("topic", "", "Kafka topic name to push records")
	record := flag.String("record", "", "Record to send to the topic")
	key := flag.String("key", "", "Key for the message (optional, default is generated uuid)")
	verbosity := flag.String("v", "info", "Verbosity")
	// if !flag.Parsed() { // avoid, seems to be true when we invoke run() from _test so we can't test args
	help := flag.Bool("help", false, "Display help")
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

	options, err := rubin.NewOptionsFromEnvconfig()
	if err != nil {
		return err
	}

	// overwrite selected options based on CLI args
	options.LogLevel = *verbosity
	client := rubin.New(options)

	if _, err := client.Produce(context.Background(), *topic, *key, *record); err != nil {
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
