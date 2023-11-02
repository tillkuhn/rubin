package rubin

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tillkuhn/rubin/internal/testutil"
)

func TestRunEnv(t *testing.T) {
	os.Clearenv()
	os.Args = []string{"noop", "-topic", testutil.Topic, "-record", "Horst Tester"}
	mock := testutil.ServerMock(http.StatusOK)
	prefix := strings.ToUpper(envconfigDefaultPrefix)
	_ = os.Setenv(prefix+"_REST_ENDPOINT", mock.URL)
	_ = os.Setenv(prefix+"_CLUSTER_ID", testutil.ClusterID)
	_ = os.Setenv(prefix+"_PRODUCER_API_KEY", "hase")
	_ = os.Setenv(prefix+"_PRODUCER_API_SECRET", "friedrich")
	opts, err := NewOptionsFromEnvconfig()
	assert.NoError(t, err)
	assert.Equal(t, mock.URL, opts.RestEndpoint)
	assert.NotEmpty(t, opts.ProducerAPIKey)
	assert.NotEmpty(t, opts.ProducerAPISecret)
	assert.NotEmpty(t, opts.ClusterID)
}
