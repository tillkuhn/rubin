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
	"github.com/tillkuhn/rubin/internal/logging"
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

type cliFlags struct {
	ce        bool
	envFile   string
	handler   string
	help      bool
	timeout   time.Duration
	topic     string
	verbosity string
}

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
	ctx := log.Logger.WithContext(context.Background())

	flags := parseFlags()

	if flags.help {
		usage.ShowHelp(envconfigPrefix, &polly.Options{})
		return nil
	}

	mLogger.Debug().Msgf("Switching to LogLevel=%s", logging.ApplyLogLevel(flags.verbosity))

	if err := initEnv(ctx, flags.envFile); err != nil {
		return err
	}
	p, err := polly.NewClientFromEnv()
	if err != nil {
		return err
	}

	// Nice: From go 1.16 onwards we no longer have to manage signal channel manually https://henvic.dev/posts/signal-notify-context/
	// also a good intro on different contexts: https://www.sohamkamani.com/golang/context/
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer func() {
		stop()
		p.WaitForClose(ctx)
	}()

	handlerFunc := selectHandler(flags.handler, flags.ce)
	timeoutChan := initTimeoutChannel(ctx, flags.timeout)
	errChan := make(chan error, 1)

	go func() {
		errChan <- p.Poll(ctx, kafka.ReaderConfig{Topic: flags.topic}, handlerFunc)
	}()

	select {
	case err = <-errChan:
		mLogger.Info().Msgf("CLI: Got error from Kafka Consumer: %v", err)
		return err
	case <-timeoutChan:
		mLogger.Info().Msgf("CLI: Timeout period exceeded after %v, shutting down consumer", flags.timeout)
	case <-ctx.Done():
		mLogger.Info().Msgf("CLI: Context Notified on '%v', waiting for polly subsystem shutdown\n", ctx.Err())
	}
	return nil
}

// set the timeout channel to nil when the timeout is zero or negative. A nil channel in a select never fires,
// so we can keep the select block simple.
func initTimeoutChannel(ctx context.Context, timeout time.Duration) <-chan time.Time {
	timeoutChan := (<-chan time.Time)(nil)
	if timeout > 0 {
		timeoutChan = time.After(timeout)
		log.Ctx(ctx).Info().Msgf("Timeout is set, consumer will terminate after %v", timeout)
	} else {
		log.Ctx(ctx).Info().Msg("No timeout set, running consumer until interrupted")
	}
	return timeoutChan
}

func initEnv(ctx context.Context, envFile string) error {
	if envFile != "" {
		log.Ctx(ctx).Info().Msgf("Loading environment from custom location %s", envFile)
		err := godotenv.Load(envFile)
		if err != nil {
			return errors.Wrap(err, "Error Loading environment vars from "+envFile)
		}
	}
	return nil
}

func parseFlags() cliFlags {
	var flags cliFlags
	flag.BoolVar(&flags.ce, "ce", false, "expect CloudEvents format for event payload")
	flag.StringVar(&flags.envFile, "env-file", "", "location of environment variable file e.g. /tmp/.env")
	flag.StringVar(&flags.handler, "handler", "", "External command with optional arguments to pass message payload via STDIN, if not set messages will be dumped to STDOUT")
	flag.BoolVar(&flags.help, "help", false, "Display this help")
	flag.DurationVar(&flags.timeout, "timeout", timeoutAfter, "Timeout duration to run the consumer, zero or negative value means no timeout")
	flag.StringVar(&flags.topic, "topic", "", "Kafka topic for message consumption")
	flag.StringVar(&flags.verbosity, "v", "info", "verbosity level, one of 'debug', 'info', 'warn', 'error'")
	flag.Parse() // call after all flags are defined and before flags are accessed by the program

	return flags
}

func selectHandler(handlerCmd string, ce bool) polly.HandleMessageFunc {
	switch {
	case handlerCmd != "":
		return PassToCallbackHandler(handlerCmd)
	case ce:
		return DumpCloudEvent
	default:
		return polly.DumpMessage
	}
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
