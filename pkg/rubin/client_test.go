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
	assert.NotEmpty(t, cc.String())

	resp, err := cc.Produce(ctx, Request{
		Topic:   "public.hello",
		Data:    "Dragonfly out in the sun you know what I mean",
		Key:     "134-5678",
		Headers: map[string]string{"heading": "for tomorrow"},
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(42), resp.Offset)
	assert.NotNil(t, resp.Timestamp)
	// assert.Equal(t, http.StatusOK, resp.ErrorCode)

	// test with default timeout and debug = true and empty key
	opts.HTTPTimeout = 0
	cc = NewClient(opts)
	_, err = cc.Produce(ctx, Request{testutil.Topic, "Hello Hase!", "", hm}) // Simple String
	assert.NoError(t, err)
	json := Request{testutil.Topic, `{"example": 1}`, "134", hm}
	_, err = cc.Produce(ctx, json) // valid json
	assert.NoError(t, err)

	event, err := NewCloudEvent("//testing/client", et, "", map[string]string{"heading": "for tomorrow"})
	assert.NoError(t, err)

	_, err = cc.Produce(ctx, Request{testutil.Topic, event, "134", hm}) // struct that can be unmarshalled
	assert.NoError(t, err)

	// test without auth (mock should return 401 is no user and pw and submitted in auth header
	opts.ProducerAPIKey = ""
	opts.ProducerAPISecret = ""
	opts.DumpMessages = true

	cc = NewClient(opts)
	_, err = cc.Produce(ctx, json) // should fail since api key / secret are empty
	assert.ErrorContains(t, err, "unexpected http")
}
