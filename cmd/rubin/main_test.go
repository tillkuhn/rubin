package main

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tillkuhn/rubin/internal/testutil"
)

func TestRunMainMessageProducer(t *testing.T) {
	os.Clearenv()
	os.Args = []string{"noop", "-topic", testutil.Topic, "-record", "Horst Tester"}
	mock := testutil.ServerMock(http.StatusOK)
	_ = os.Setenv("KAFKA_REST_ENDPOINT", mock.URL)
	_ = os.Setenv("KAFKA_CLUSTER_ID", testutil.ClusterID)
	_ = os.Setenv("KAFKA_API_KEY", "hase")
	_ = os.Setenv("KAFKA_API_SECRET", "friedrich")
	err := run()
	assert.NoError(t, err)
}
