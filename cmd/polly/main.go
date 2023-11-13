package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"

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
	timeoutAfter = 120 * time.Second
)

func main() {
	fmt.Printf("Welcome to %s %s built %s by %s (%s)\n\n", appName, version, date, builtBy, commit)
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var topic = flag.String("topic", "", "Kafka topic for message consumption ")
	var envFile = flag.String("env-file", "", "location of environment variable file e.g. /tmp/.env")
	flag.Parse() // call after all flags are defined and before flags are accessed by the program
	if *envFile != "" {
		log.Printf("Loading environment from custom location %s", *envFile)
		err := godotenv.Load(*envFile)
		if err != nil {
			log.Fatalf("Error Loading environment vars from %s: %v", *envFile, err)
		}
	}
	p, err := polly.NewClientFromEnv()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt) // https://henvic.dev/posts/signal-notify-context/
	defer stop()
	errChan := make(chan error, 1)
	go func() {
		errChan <- p.Consume(ctx, polly.ConsumeRequest{Topic: *topic, Handler: polly.DumpMessage})
	}()

	select {
	case err := <-errChan:
		fmt.Printf("CLI: Got error from Kafka Consumer: %v\n", err)
	case <-time.After(timeoutAfter):
		fmt.Printf("CLI: Timeout after %v\n", timeoutAfter)
	case <-ctx.Done():
		stop() // should be called asap according to docs
		fmt.Println("CLI: Received interrupt signal, shutdown polly consumer")
		p.CloseWait() // make sure wait-group is zero
	}
	return nil
}
