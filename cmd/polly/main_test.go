package main

import (
	"os"
	"testing"
	"time"

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
