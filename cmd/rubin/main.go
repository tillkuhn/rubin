package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kelseyhightower/envconfig"
	"github.com/user/rubin/pkg/rubin"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/user/rubin/internal/log"
)

const AppID = "rubin"

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
	if !flag.Parsed() { // or we get panic if configureEnvironment is called twice
		// debug = flag.Bool("debug", false, "log debug")
		help := flag.Bool("help", false, "Display help")
		flag.Parse()
		if *help {
			// app := path.Base(os.Args[0])
			// fmt.Printf("*%s* is configured via %s.\nThe following environment variables can be used (`%s --help`):",
			//	app, "https://github.com/kelseyhightower/envconfig[envconfig]", app)
			tabs := tabwriter.NewWriter(os.Stdout, 1, 0, 4, ' ', 0)
			_ = envconfig.Usagef("env_config", &rubin.Options{}, tabs, envconfig.DefaultTableFormat)
			_ = tabs.Flush()
			fmt.Println("\nThis Application also supports the following CLI arguments")
			flag.PrintDefaults()
			return nil
		}
	}
	logger := log.NewAtLevel(os.Getenv("LOG_LEVEL"))
	var options rubin.Options
	if err := envconfig.Process(AppID, &options); err != nil {
		logger.Errorf("Cannot process environment config: %v", err)
		return err
	}

	defer func() {
		_ = logger.Sync() // flushed any buffered log entries
	}()

	client := rubin.New(&options)
	if _, err := client.Produce(context.Background(), *topic, "", record); err != nil {
		logger.Errorf("Cannot produce record to %s: %v", *topic, err)
		return err
	}

	return nil
}
