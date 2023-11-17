package rubin

import (
	"encoding/base64"
	"net/url"
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
	assert.Equal(t, "user:password", mustB64Decode(o.BasicAuth()))

	// test with user encoded in url
	o.ProducerTopicURL = mustParseURL("https://friendOfSomeUser@some.cloud:443/kafka/v3/clusters/abc-932/topics/ciao.world")
	assert.Equal(t, "friendOfSomeUser:password", mustB64Decode(o.BasicAuth()))

	// test with user + pass encoded in url
	o.ProducerTopicURL = mustParseURL("https://friendOfSomeUser:notYourBusiness@some.cloud:443/kafka/v3/clusters/abc-932/topics/ciao.world")
	assert.Equal(t, "friendOfSomeUser:notYourBusiness", mustB64Decode(o.BasicAuth()))
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
	o.ProducerTopicURL = mustParseURL("https://some.cloud:443/kafka/v3/clusters/abc-932/topics/ciao.world")
	a2 := o.RecordEndpoint("")
	assert.Equal(t, "https://some.cloud:443/kafka/v3/clusters/abc-932/topics/ciao.world/records", a2)

	// default topic configured with topic URL, with request specific topic (topic name should be overwritten)
	assert.Equal(t, "https://some.cloud:443/kafka/v3/clusters/abc-932/topics/small.world/records", o.RecordEndpoint("small.world"))

	// make sure that user and pw data is stripped from url
	o.ProducerTopicURL = mustParseURL("https://user123:nix@some.cloud:443/kafka/v3/clusters/abc-932/topics/user.world")
	assert.Equal(t, "https://some.cloud:443/kafka/v3/clusters/abc-932/topics/user.world/records", o.RecordEndpoint(""))

	// test via env
	defer os.Clearenv()
	_ = os.Setenv(strings.ToUpper(envconfigDefaultPrefix)+"_PRODUCER_TOPIC_URL", "https://env123:nixEnv@env.cloud:443/kafka/v3/clusters/env-932/topics/env.world")
	oe, err := NewOptionsFromEnv()
	oe.ProducerAPISecret = "viaAPISecret" // should not matter if pw is part of the URL
	assert.NoError(t, err)
	assert.Equal(t, "https://env.cloud:443/kafka/v3/clusters/env-932/topics/env.world/records", oe.RecordEndpoint(""))
	assert.Equal(t, "env123:nixEnv", mustB64Decode(oe.BasicAuth()))

	// but now ProducerAPISecret should kick in
	_ = os.Setenv(strings.ToUpper(envconfigDefaultPrefix)+"_PRODUCER_TOPIC_URL", "https://env123@env.cloud:443/kafka/v3/clusters/env-932/topics/env.world")
	assert.Equal(t, "env123:nixEnv", mustB64Decode(oe.BasicAuth()))
	oe2, err := NewOptionsFromEnv()
	assert.NoError(t, err)
	oe2.ProducerAPISecret = "viaAPISecret" // should not matter if pw is part of the URL
	assert.Equal(t, "env123:viaAPISecret", mustB64Decode(oe2.BasicAuth()))
}

func mustParseURL(rawURL string) url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err.Error())
	}
	return *u
}

func mustB64Decode(encStr string) string {
	dStr, err := base64.StdEncoding.DecodeString(encStr)
	if err != nil {
		panic(err.Error())
	}
	return string(dStr)
}
