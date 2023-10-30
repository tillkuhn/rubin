package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kelseyhightower/envconfig"
	"github.com/tillkuhn/rubin/pkg/rubin"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/tillkuhn/rubin/internal/log"
)

const envconfigPrefix = "kafka"

// useful variables to pass with ldflags during build, for example
// go run -ldflags="-w -s -X 'main.version=$(shell git describe --tags --abbrev=0)' -X 'main.commit=$(shell git rev-parse --short HEAD)'"
var (
	version = "latest"
	date    = "now"
	commit  = ""
)

func main() {
	// Disable automaxprocs log see https://github.com/uber-go/automaxprocs/issues/18
	nopLog := func(string, ...interface{}) {}
	_, _ = maxprocs.Set(maxprocs.Logger(nopLog))
	if err := run(); err != nil {
		// _, _ = fmt.Fprintf(os.Stderr, "an error occurred: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	topic := flag.String("topic", "", "Kafka topic name to push records")
	record := flag.String("record", "", "Record to send to the topic")
	key := flag.String("key", "", "Key for the message (optional, default is generated uuid)")
	verbosity := flag.String("v", "info", "Verbosity")
	// if !flag.Parsed() { // avoid, seems to be true when we invoke run() from _test so we can't test args
	help := flag.Bool("help", false, "Display help")
	flag.Parse()
	if *help {
		// app := path.Base(os.Args[0])
		// fmt.Printf("*%s* is configured via %s.\nThe following environment variables can be used (`%s --help`):",
		//	app, "https://github.com/kelseyhightower/envconfig[envconfig]", app)
		usagePadding := 4
		tabs := tabwriter.NewWriter(os.Stdout, 1, 0, usagePadding, ' ', 0)
		_ = envconfig.Usagef(envconfigPrefix, &rubin.Options{}, tabs, envconfig.DefaultTableFormat)
		_ = tabs.Flush()
		// flag.Usage() // https://stackoverflow.com/a/23726033/4292075
		fmt.Println("\nThis Application also supports the following CLI arguments")
		flag.PrintDefaults()
		return nil
	}
	logger := log.NewAtLevel(*verbosity)
	defer func() {
		_ = logger.Sync() // flushed any buffered log entries
	}()
	logger.Infow("Welcome to rubin", "version", version, "built", date, "commit", commit)

	var options rubin.Options
	if err := envconfig.Process(envconfigPrefix, &options); err != nil {
		logger.Errorf("Cannot process environment config: %v", err)
		return err
	}
	client := rubin.New(&options)

	if _, err := client.Produce(context.Background(), *topic, *key, *record); err != nil {
		logger.Errorf("Cannot produce record to %s: %v", *topic, err)
		return err
	}

	return nil
}
