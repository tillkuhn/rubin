package rubin

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/tillkuhn/rubin/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func init() {
	// log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

const dumpMessages = false

func TestProduceMessageOK(t *testing.T) {
	hm := map[string]string{"heading": "for tomorrow"}
	ctx := context.Background()
	srv := testutil.ServerMock() // "https://pkc-zpjg0.eu-central-1.aws.confluent.cloud:443"
	defer srv.Close()
	opts := &Options{
		RestEndpoint: srv.URL, ClusterID: testutil.ClusterID,
		ProducerAPIKey: "test.key", ProducerAPISecret: "test.pw",
		HTTPTimeout:  5 * time.Second,
		LogLevel:     "debug",
		DumpMessages: dumpMessages,
	}
	cc := NewClient(opts)
	assert.NotEmpty(t, cc.String())

	resp, err := cc.Produce(ctx, Request{
		Topic:   testutil.Topic(200),
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
	_, err = cc.Produce(ctx, Request{testutil.Topic(200), "Hello Hase!", "", hm}) // Simple String
	assert.NoError(t, err)

	json := Request{testutil.Topic(200), `{"example": 1}`, "134", hm}
	_, err = cc.Produce(ctx, json) // valid json
	assert.NoError(t, err)

	event, err := NewCloudEvent("//testing/client", "", map[string]string{"heading": "for tomorrow"})
	assert.NoError(t, err)

	_, err = cc.Produce(ctx, Request{testutil.Topic(200), event, "134", hm}) // struct that can be unmarshalled
	assert.NoError(t, err)

	// test without auth (mock should return 401 is no user and pw and submitted in auth header
	opts.ProducerAPIKey = ""
	opts.ProducerAPISecret = ""
	opts.DumpMessages = true

	cc = NewClient(opts)
	_, err = cc.Produce(ctx, json) // should fail since api key / secret are empty
	assert.ErrorContains(t, err, "unexpected http")
}

func TestProduceMessageError(t *testing.T) {
	// hm := map[string]string{"heading": "for tomorrow"}
	ctx := context.Background()
	srv := testutil.ServerMock() // "https://pkc-zpjg0.eu-central-1.aws.confluent.cloud:443"
	defer srv.Close()
	req := Request{
		Topic:   testutil.Topic(http.StatusForbidden),
		Data:    "Dragonfly out in the sun you know what I mean",
		Key:     "134-5678",
		Headers: map[string]string{"heading": "for tomorrow"},
	}
	opts := &Options{
		RestEndpoint: srv.URL, ClusterID: testutil.ClusterID,
		ProducerAPIKey: "test.key", ProducerAPISecret: "test.pw",
	}
	cc := NewClient(opts)
	_, err := cc.Produce(ctx, req)
	assert.ErrorContains(t, err, "Not authorized")
}
