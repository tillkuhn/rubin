package rubin

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tillkuhn/rubin/internal/testutil"
)

func TestRunEnv(t *testing.T) {
	os.Clearenv()
	os.Args = []string{"noop", "-topic", testutil.Topic(200), "-record", "Horst Tester"}
	mock := testutil.ServerMock()
	prefix := strings.ToUpper(envconfigDefaultPrefix)
	_ = os.Setenv(prefix+"_REST_ENDPOINT", mock.URL)
	_ = os.Setenv(prefix+"_CLUSTER_ID", testutil.ClusterID)
	_ = os.Setenv(prefix+"_PRODUCER_API_KEY", "hase")
	_ = os.Setenv(prefix+"_PRODUCER_API_SECRET", "friedrich")
	opts, err := NewOptionsFromEnv()
	assert.NoError(t, err)
	assert.Equal(t, mock.URL, opts.RestEndpoint)
	assert.NotEmpty(t, opts.ProducerAPIKey)
	assert.NotEmpty(t, opts.ProducerAPISecret)
	assert.NotEmpty(t, opts.ClusterID)
	assert.Contains(t, opts.String(), mock.URL)
}

func TestMust(t *testing.T) {
	_ = os.Setenv(strings.ToUpper(envconfigDefaultPrefix)+"_REST_ENDPOINT", "//hase")
	opts := Must[*Options](NewOptionsFromEnv())
	assert.Equal(t, "//hase", opts.RestEndpoint)
}
