package main

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/tillkuhn/rubin/internal/testutil"
)

// Test error handling (does not require server mock)
func TestRunMainWithImmediateTimeout(t *testing.T) {
	// resetEnvAndFlags()
	// setupMock()
	timeoutAfter = 10 * time.Millisecond // speed up timeout
	os.Args = []string{"noop", "-topic", testutil.Topic(200), "-ce"}
	errMain := run()
	// sending a signal to myself does not work, apparently
	// p, err := os.FindProcess(os.Getpid()); err = p.Signal(os.Interrupt)
	assert.NoError(t, errMain) // b/c deadline exceeded is not considered an error
}

func TestPassToCallbackHandler(t *testing.T) {
	// Setup logger to capture output
	var logBuf bytes.Buffer
	log.Logger = zerolog.New(&logBuf)

	handler := PassToCallbackHandler("cat")
	msg := kafka.Message{Value: []byte("test payload")}

	handler(context.Background(), msg)

	logOutput := logBuf.String()
	if !bytes.Contains([]byte(logOutput), []byte("Handler command executed successfully")) {
		t.Errorf("Expected successful handler execution, got log: %s", logOutput)
	}
}
