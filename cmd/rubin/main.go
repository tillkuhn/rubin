package main

import (
	"context"
	"github.com/user/rubin/pkg/rubin"
	"go.uber.org/automaxprocs/maxprocs"
	"os"

	"github.com/user/rubin/internal/log"
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
	logger := log.NewAtLevel(os.Getenv("LOG_LEVEL"))

	defer func() {
		_ = logger.Sync() // flushed any buffered log entries
	}()

	logger.Infow("Hello world!", "location", "world")

	client := rubin.New(&rubin.Options{
		RestEndpoint: "https://localhost:443",
		ClusterID:    "abc-r2d2",
		ApiKey:       "1234567890",
		ApiPassword:  "**********",
	})
	topic := "toppig"
	if _, err := client.Produce(context.Background(), topic, "", "hey"); err != nil {
		logger.Errorf("Cannot produce record to %s: %v", topic, err)
		return err
	}

	return nil

}
