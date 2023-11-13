package polly

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

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

func TestNew(t *testing.T) {
	polly, err2 := testClient(t)
	assert.NotNil(t, polly)
	assert.NoError(t, err2)

	ctx := context.WithValue(context.Background(), contextKeyTopic, testTopic)
	err := polly.Consume(ctx, ConsumeRequest{Topic: testTopic, Handler: DumpMessage})
	assert.NoError(t, err)
	polly.CloseWait()

	ctx = context.WithValue(context.Background(), contextKeyTopic, errorTopic)
	err = polly.Consume(ctx, ConsumeRequest{Topic: errorTopic, Handler: DumpMessage})
	assert.Error(t, err)
	polly.CloseWait()
}

func testClient(t *testing.T) (*Client, error) {
	// cf := func(ctx context.Context, msg kafka.Message) {}
	prefix := strings.ToUpper(envconfigDefaultPrefix)
	_ = os.Setenv(prefix+"_CONSUMER_TOPIC", "hase")
	_ = os.Setenv(prefix+"_KAFKA_CONSUMER_START_LAST", "true")
	defer func() {
		_ = os.Unsetenv(prefix + "_KAFKA_CONSUMER_START_LAST")
		_ = os.Unsetenv(prefix + "_CONSUMER_TOPIC")
	}()

	k, err := NewClientFromEnv()
	assert.NoError(t, err)
	k.readerFactory = mockMessageReader
	k.options.ConsumerMaxReceive = 2 // after max 2 messages, consumer will exit
	return k, err
}

func mockMessageReader(_ kafka.ReaderConfig) MessageReader {
	return &MockMessageReader{}
}

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
