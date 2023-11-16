package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/pkg/errors"

	"github.com/joho/godotenv"
	"github.com/tillkuhn/rubin/internal/log"

	"github.com/tillkuhn/rubin/pkg/polly"
)

const (
	appName = "polly"
)

var (
	version      = "latest"
	date         = "now"
	commit       = ""
	builtBy      = "go"
	timeoutAfter = 15 * time.Second
)

func main() {
	fmt.Printf("Welcome to %s %s built %s by %s (%s)\n\n", appName, version, date, builtBy, commit)
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	logger := log.New()

	var topic = flag.String("topic", "", "Kafka topic for message consumption ")
	var ce = flag.Bool("ce", false, "except CloudEvents format for event payload")

	var envFile = flag.String("env-file", "", "location of environment variable file e.g. /tmp/.env")
	flag.Parse() // call after all flags are defined and before flags are accessed by the program
	if *envFile != "" {
		logger.Infof("Loading environment from custom location %s", *envFile)
		err := godotenv.Load(*envFile)
		if err != nil {
			return errors.Wrap(err, "Error Loading environment vars from "+*envFile)
		}
	}
	p, err := polly.NewClientFromEnv()
	if err != nil {
		return err
	}

	// Nice: we no longer have to manage signal chanel manually https://henvic.dev/posts/signal-notify-context/
	// also a good intro on different contexts: https://www.sohamkamani.com/golang/context/
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer func() {
		stop()
		p.WaitForClose()
	}()

	handler := polly.DumpMessage
	if *ce {
		handler = DumpCloudEvent
	}
	errChan := make(chan error, 1)
	go func() {
		errChan <- p.Poll(ctx, kafka.ReaderConfig{Topic: *topic}, handler)
	}()

	select {
	case err := <-errChan:
		logger.Infof("CLI: Got error from Kafka Consumer: %v\n", err)
		return err
	case <-time.After(timeoutAfter):
		logger.Infof("CLI: Deadline exceededafter %v\n", timeoutAfter)
	case <-ctx.Done():
		logger.Infof("CLI: Context Notified on '%v', waiting for polly subsytem shutdown\n", ctx.Err())
	}
	return nil
}

func DumpCloudEvent(_ context.Context, message kafka.Message) {
	ce, err := polly.AsCloudEvent(message)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%d/%d type %s\npayload: %v\n", message.Partition, message.Offset, ce.Type(), ce)
}
