package rubin

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/tillkuhn/rubin/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func init() {
	// log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func TestProduceMessageOK(t *testing.T) {
	hm := make(map[string]string)
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
	cc := New(opts)
	// strings.NewReader("hello world")
	resp, err := cc.Produce(ctx, "public.welcome", "1234", "Hello Hase!", hm)
	assert.NoError(t, err)
	assert.Equal(t, int32(42), resp.Offset)
	assert.NotNil(t, resp.Timestamp)
	// assert.Equal(t, http.StatusOK, resp.ErrorCode)

	// test with default timeout and debug = true and empty key
	opts.HTTPTimeout = 0
	cc = New(opts)
	_, err = cc.Produce(ctx, "public.welcome", "", "Hello Hase!", hm) // Simple String
	assert.NoError(t, err)
	_, err = cc.Produce(ctx, "public.welcome", "", `{"example": 1}`, hm) // valid json
	assert.NoError(t, err)

	event := Event{
		Action:  "update/event",
		Message: "go with me",
		Time:    time.Now().Round(time.Second), // make sure we round to .SSS
		Source:  os.Args[0],
	}

	_, err = cc.Produce(ctx, "public.welcome", "abc/123", event, hm) // struct that can be unmarshalled
	assert.NoError(t, err)

	// test without auth (mock should return 401 is no user and pw and submitted in auth header
	opts.ProducerAPIKey = ""
	opts.ProducerAPISecret = ""
	opts.DumpMessages = true

	cc = New(opts)
	_, err = cc.Produce(ctx, "public.welcome", "", "Hello Hase!", hm) // Simple String
	assert.ErrorContains(t, err, "unexpected http")
}
