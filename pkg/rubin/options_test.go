package rubin

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tillkuhn/rubin/internal/testutil"
)

var prefix = strings.ToUpper(envconfigDefaultPrefix)

func TestRunEnv(t *testing.T) {
	defer os.Clearenv()
	os.Args = []string{"noop", "-topic", testutil.Topic(200), "-record", "Horst Tester"}
	mock := testutil.ServerMock()
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
	defer os.Clearenv()
	_ = os.Setenv(strings.ToUpper(envconfigDefaultPrefix)+"_REST_ENDPOINT", "//hase")
	opts := Must[*Options](NewOptionsFromEnv())
	assert.Equal(t, "//hase", opts.RestEndpoint)
}

func TestOptionsError(t *testing.T) {
	defer os.Clearenv()
	_ = os.Setenv(prefix+"_HTTP_TIMEOUT", "this is not a duration")
	_, err := NewOptionsFromEnv()
	assert.ErrorContains(t, err, "invalid duration") // envconfig.Process: strconv.ParseInt: ...
}

func TestBasicAuth(t *testing.T) {
	o := Options{
		ProducerAPIKey:    "user",
		ProducerAPISecret: "password",
	}
	assert.Equal(t, "dXNlcjpwYXNzd29yZA==", o.BasicAuth())

	// test with use encoded in url
	o.TopicURL = "https://friendOfSomeUser@some.cloud:443/kafka/v3/clusters/abc-932/topics/ciao.world"
	assert.Equal(t, "ZnJpZW5kT2ZTb21lVXNlcjpwYXNzd29yZA==", o.BasicAuth())
}

func TestEndpoint(t *testing.T) {
	o := Options{
		RestEndpoint: "https://some.cloud:443",
		ClusterID:    "lka-123",
	}
	// request specific topic, cluster configured with endpoint and cluster id
	a := o.RecordEndpoint("hello.world")
	assert.Equal(t, "https://some.cloud:443/kafka/v3/clusters/lka-123/topics/hello.world/records", a)

	// default topic configured with topic URL, no request specific topic
	o.TopicURL = "https://some.cloud:443/kafka/v3/clusters/abc-932/topics/ciao.world"
	a2 := o.RecordEndpoint("")
	assert.Equal(t, "https://some.cloud:443/kafka/v3/clusters/abc-932/topics/ciao.world/records", a2)

	// default topic configured with topic URL, with request specific topic (topic name should be overwritten)
	a3 := o.RecordEndpoint("small.world")
	assert.Equal(t, "https://some.cloud:443/kafka/v3/clusters/abc-932/topics/small.world/records", a3)
}
