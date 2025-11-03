package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/tillkuhn/rubin/internal/usage"

	"github.com/pkg/errors"

	"github.com/joho/godotenv"

	"github.com/tillkuhn/rubin/pkg/polly"
)

const (
	envconfigPrefix = "kafka"
	appName         = "polly"
)

var (
	version      = "latest"
	date         = "now"
	commit       = ""
	builtBy      = "go"
	timeoutAfter = 30 * time.Second
)

func main() {
	fmt.Printf("Welcome to %s %s built %s by %s (%s)\n\n", appName, version, date, builtBy, commit)
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	log.Logger = log.With().Str("app", appName).Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
	mLogger := log.With().Str("logger", "main").Logger()

	var topic = flag.String("topic", "", "Kafka topic for message consumption ")
	var help = flag.Bool("help", false, "Display this help")
	var ce = flag.Bool("ce", false, "except CloudEvents format for event payload")
	var callback = flag.String("callback", "", "Callback command with optional arguments to pass message payload via STDIN")
	var envFile = flag.String("env-file", "", "location of environment variable file e.g. /tmp/.env")
	flag.Parse() // call after all flags are defined and before flags are accessed by the program
	if *help {
		usage.ShowHelp(envconfigPrefix, &polly.Options{})
		return nil
	}

	if *envFile != "" {
		mLogger.Info().Msgf("Loading environment from custom location %s", *envFile)
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
	ctx = log.Logger.WithContext(ctx)
	defer func() {
		stop()
		p.WaitForClose(ctx)
		// _ = logger.Sync()
	}()

	handler := polly.DumpMessage // default simple message-as-is dump handler
	if *callback != "" {
		handler = PassToCallbackHandler(*callback)
	} else if *ce {
		handler = DumpCloudEvent
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- p.Poll(ctx, kafka.ReaderConfig{Topic: *topic}, handler)
	}()

	select {
	case err = <-errChan:
		mLogger.Info().Msgf("CLI: Got error from Kafka Consumer: %v\n", err)
		return err
	case <-time.After(timeoutAfter):
		mLogger.Info().Msgf("CLI: Deadline exceeded after %v\n", timeoutAfter)
	case <-ctx.Done():
		mLogger.Info().Msgf("CLI: Context Notified on '%v', waiting for polly subsytem shutdown\n", ctx.Err())
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

// PassToCallbackHandler wraps handlerCmd and returns a function that can be used as polly.HandleMessageFunc
func PassToCallbackHandler(handlerCmd string) func(ctx context.Context, message kafka.Message) {
	log.Info().Msgf("Registering callback for externalCommand: %s", handlerCmd)
	return func(ctx context.Context, message kafka.Message) {
		payload := string(message.Value)
		// Split command and arguments
		parts := strings.Fields(handlerCmd)
		if len(parts) == 0 {
			log.Error().Msg("Handler command is empty")
			return
		}
		//
		cmd := exec.CommandContext(ctx, parts[0], parts[1:]...) // #nosec G204
		cmd.Stdin = bytes.NewBufferString(payload)

		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error().Err(err).Msgf("Failed to execute handler command: %s, output: %s", handlerCmd, string(output))
			return
		}
		log.Info().Msgf("Handler command executed successfully: %s, output: %s", handlerCmd, string(output))
	}
}
