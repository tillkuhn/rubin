package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/tillkuhn/rubin/internal/testutil"
	"net/http"
	"os"
	"testing"
)

func TestRunMainMessageProducer(t *testing.T) {
	os.Clearenv()
	mock := testutil.ServerMock(http.StatusOK)
	_ = os.Setenv("KAFKA_REST_ENDPOINT", mock.URL)
	_ = os.Setenv("KAFKA_CLUSTER_ID", testutil.ClusterID)
	err := run()
	assert.ErrorContains(t, err, " response status code 404") // url ü cluster id ok, but no topic ...
}
