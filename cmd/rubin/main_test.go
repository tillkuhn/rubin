package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tillkuhn/rubin/internal/testutil"
)

func TestRunMainMessageProducer(t *testing.T) {
	resetEnvAndFlags()
	setupMock()
	os.Args = []string{"noop", "-topic", testutil.Topic(200), "-record", "Horst Tester", "-header", "id=1", "-source", "open/source", "-ce"}
	err := run()
	assert.NoError(t, err)

	// Test simple string
}

// Test error handling (does not require server mock)
func TestRunMainMessageProducerSimple(t *testing.T) {
	resetEnvAndFlags()
	setupMock()
	os.Args = []string{"noop", "-topic", testutil.Topic(200), "-record", "This is a simple string"}
	err := run()
	assert.NoError(t, err)
}

func TestHelp(t *testing.T) {
	resetEnvAndFlags()
	os.Args = []string{"noop", "-help"}
	err := run()
	assert.NoError(t, err)
}

func TestEmptyRecord(t *testing.T) {
	resetEnvAndFlags()
	os.Args = []string{"noop", "-topic", "hase", "-record", ""}
	err := run()
	assert.ErrorContains(t, err, "record must not be empty")
}

// Helper

func resetEnvAndFlags() {
	os.Clearenv()
	// How to unset flags Visited on command line in GoLang for Tests
	// https://stackoverflow.com/a/57972717/4292075
	// Otherwise you get "tmp/GoLand/___main_test_go.test flag redefined: topic"
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // flags are now reset
}

func setupMock() {
	mock := testutil.ServerMock()
	_ = os.Setenv("KAFKA_REST_ENDPOINT", mock.URL)
	_ = os.Setenv("KAFKA_CLUSTER_ID", testutil.ClusterID)
	_ = os.Setenv("KAFKA_PRODUCER_API_KEY", "hase")
	_ = os.Setenv("KAFKA_PRODUCER_API_SECRET", "friedrich")
}
