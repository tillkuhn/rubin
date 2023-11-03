package rubin

import (
	"context"
	"testing"
	"time"

	"github.com/tillkuhn/rubin/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func init() {
	// log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func TestProduceMessageOK(t *testing.T) {
	hm := map[string]string{"heading": "for tomorrow"}
	et := "com.test.event"
	ctx := context.Background()
	srv := testutil.ServerMock(200) // "https://pkc-zpjg0.eu-central-1.aws.confluent.cloud:443"
	defer srv.Close()
	opts := &Options{
		RestEndpoint: srv.URL, ClusterID: testutil.ClusterID,
		ProducerAPIKey: "test.key", ProducerAPISecret: "test.pw",
		HTTPTimeout:  5 * time.Second,
		LogLevel:     "debug",
		DumpMessages: true,
	}
	cc := NewClient(opts)
	// strings.NewReader("hello world")
	resp, err := cc.Produce(ctx, testutil.Topic, "1234", "Hello Hase!", hm)
	assert.NoError(t, err)
	assert.Equal(t, int32(42), resp.Offset)
	assert.NotNil(t, resp.Timestamp)
	// assert.Equal(t, http.StatusOK, resp.ErrorCode)

	// test with default timeout and debug = true and empty key
	opts.HTTPTimeout = 0
	cc = NewClient(opts)
	_, err = cc.Produce(ctx, "public.welcome", "", "Hello Hase!", hm) // Simple String
	assert.NoError(t, err)
	_, err = cc.Produce(ctx, "public.welcome", "", `{"example": 1}`, hm) // valid json
	assert.NoError(t, err)

	event, err := NewCloudEvent("//testing/client", et, "", map[string]string{"heading": "for tomorrow"})
	assert.NoError(t, err)

	_, err = cc.Produce(ctx, "public.welcome", "abc/123", event, hm) // struct that can be unmarshalled
	assert.NoError(t, err)

	// test without auth (mock should return 401 is no user and pw and submitted in auth header
	opts.ProducerAPIKey = ""
	opts.ProducerAPISecret = ""
	opts.DumpMessages = true

	cc = NewClient(opts)
	_, err = cc.Produce(ctx, "public.welcome", "", "Hello Hase!", hm) // Simple String
	assert.ErrorContains(t, err, "unexpected http")
}
