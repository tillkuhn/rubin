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
	srv := testutil.ServerMock(200) // "https://pkc-zpjg0.eu-central-1.aws.confluent.cloud:443"
	defer srv.Close()
	cc := New(&Options{srv.URL, testutil.ClusterID, "test.key", "test.pw", 5 * time.Second, false})
	// strings.NewReader("hello world")
	resp, err := cc.Produce(context.Background(), "public.welcome", "1234", "Hello Hase!")
	assert.NoError(t, err)
	assert.Equal(t, int32(42), resp.Offset)
	assert.NotNil(t, resp.Timestamp)
	// assert.Equal(t, http.StatusOK, resp.ErrorCode)

	// test with default timeout and debug = true and empty key
	cc = New(&Options{srv.URL, testutil.ClusterID, "test.key", "test.pw", 0, true})
	_, err = cc.Produce(context.Background(), "public.welcome", "", "Hello Hase!")
	assert.NoError(t, err)
}
