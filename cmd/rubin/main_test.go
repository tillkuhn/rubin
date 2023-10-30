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
	os.Args = []string{"noop", "-topic=hase", "-record=horst", "arg2"}
	mock := testutil.ServerMock(http.StatusOK)
	_ = os.Setenv("KAFKA_REST_ENDPOINT", mock.URL)
	_ = os.Setenv("KAFKA_CLUSTER_ID", testutil.ClusterID)
	err := run()
	assert.ErrorContains(t, err, " response status code 404") // url ü cluster id ok, but no topic ...
}
