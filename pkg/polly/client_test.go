package polly

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/tillkuhn/rubin/internal/testutil"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

var (
	testTopic  = "mock.hase"
	errorTopic = "mock.error"

	errTest = errors.New("test error")
)

// contextKey Fix "should not use basic type string as key in context.WithValue" go lint
type contextKey int

const (
	contextKeyTopic contextKey = iota
	// ...
)

func TestNewClient(t *testing.T) {
	polly, err2 := testClient(t)
	assert.NotNil(t, polly)
	assert.NoError(t, err2)

	ctx := context.Background()
	ctx = context.WithValue(ctx, contextKeyTopic, testTopic)
	err := polly.Poll(ctx, kafka.ReaderConfig{Topic: testTopic}, DumpMessage)
	assert.NoError(t, err)
	polly.WaitForClose(ctx)

	ctx = context.WithValue(context.Background(), contextKeyTopic, errorTopic)
	err = polly.Poll(ctx, kafka.ReaderConfig{Topic: errorTopic}, DumpMessage)
	assert.Error(t, err)
	polly.WaitForClose(ctx)
}

func TestCloudEvent(t *testing.T) {
	respBytes, err := os.ReadFile(testutil.TestDataDir + "/cloudevent-ci.json")
	assert.NoError(t, err)
	km := kafka.Message{
		Topic:     "ci.events",
		Partition: 0,
		Offset:    0,
		Key:       []byte("key/123"),
		Value:     respBytes,
		Headers:   nil,
		Time: time.Date(2020, time.April,
			11, 21, 34, 01, 0, time.UTC),
	}
	_, err = AsCloudEvent(km)
	assert.ErrorContains(t, err, "invalid content-type")

	// try again with the correct Header
	km.Headers = []kafka.Header{{Key: "content-type", Value: []byte(cloudevents.ApplicationCloudEventsJSON)}}
	ce, err := AsCloudEvent(km)
	assert.NoError(t, err)
	assert.Equal(t, "net.timafe.events.ci.published", ce.Type())
}
func TestRealReader(t *testing.T) {
	dr := defaultMessageReader(kafka.ReaderConfig{Brokers: []string{"hase"}, Topic: "horst"})
	assert.NotNil(t, dr)
}

func testClient(t *testing.T) (*Client, error) {
	// cf := func(ctx context.Context, msg kafka.Message) {}
	prefix := strings.ToUpper(envconfigDefaultPrefix)
	_ = os.Setenv(prefix+"_CONSUMER_TOPIC", "hase")
	_ = os.Setenv(prefix+"_CONSUMER_START_LAST", "true")
	defer func() {
		_ = os.Unsetenv(prefix + "_CONSUMER_START_LAST")
		_ = os.Unsetenv(prefix + "_CONSUMER_TOPIC")
	}()

	k, err := NewClientFromEnv()
	assert.NoError(t, err)
	assert.True(t, k.options.ConsumerStartLast)
	assert.Equal(t, "localhost:9092", k.options.BootstrapServers)
	assert.Contains(t, k.String(), "localhost:9092")
	k.readerFactory = mockMessageReader
	k.options.ConsumerMaxReceive = 2 // after max 2 messages, consumer will exit
	return k, err
}

func mockMessageReader(_ kafka.ReaderConfig) MessageReader {
	return &MockMessageReader{}
}

// MockMessageReader is used instead of the real kafka.Reader to produce a stream of kafka messages for unit testing
type MockMessageReader struct {
	offset int64
}

func (mr *MockMessageReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	mr.offset++
	topic := ctx.Value(contextKeyTopic)
	if topic == errorTopic {
		return kafka.Message{}, fmt.Errorf("%w this topic is bad, better use %s", errTest, testTopic)
	}
	return kafka.Message{
		Partition: 42,
		Topic:     topic.(string),
		Time:      time.Now().Round(time.Second),
		Value:     []byte("mock"),
		Offset:    mr.offset - 1,
	}, nil
}
func (mr *MockMessageReader) Close() error {
	return nil
}
