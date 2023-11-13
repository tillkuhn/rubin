package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

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
	timeoutAfter = 9 * time.Second
)

func main() {
	fmt.Printf("Welcome to %s %s built %s by %s (%s)\n\n", appName, version, date, builtBy, commit)
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	p, err := polly.NewClientFromEnv()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt) // https://henvic.dev/posts/signal-notify-context/
	defer stop()
	errChan := make(chan error, 1)
	go func() {
		errChan <- p.Consume(ctx, polly.ConsumeRequest{Topic: "ci.event", Handler: polly.DumpMessage})
	}()

	select {
	case err := <-errChan:
		fmt.Printf("Got error from Kafka Consumer: %v\n", err)
	case <-time.After(timeoutAfter):
		fmt.Printf("CLI timeout after %v\n", timeoutAfter)
	case <-ctx.Done():
		stop() // should be called asap according to docs
		fmt.Println("Received interrupt signal, shutdown polly consumer")
		p.CloseWait() // make sure wait-group is zero
	}
	return nil
}
