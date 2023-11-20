package polly

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/segmentio/kafka-go"
	"github.com/tillkuhn/rubin/internal/log"
	"go.uber.org/zap"

	"github.com/segmentio/kafka-go/sasl/plain"
)

const (
	defaultDialTimeout      = 5 * time.Second
	defaultCloseWaitTimeout = 10 * time.Second
	minConsumeBytes         = 10
	maxConsumeBytes         = 10e6 // 10 MB
	// retentionTime optionally sets the length of time the consumer group will be saved by the broker, Default 24h
	retentionTime = 10 * time.Minute
)

// errInvalidContentType used as static error for Kafka messages with unexpected or no content-type header
var errInvalidContentType = errors.New("invalid content-type")

// HandleMessageFunc consumer will pass received messages to a function that matches this type
type HandleMessageFunc func(ctx context.Context, message kafka.Message)

// MessageReader interface that makes it easy to mock the real kafka.Reader in Poll() for testing purposes
type MessageReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}

// defaultMessageReader returns the standard segmentio/kafka-go reader
func defaultMessageReader(config kafka.ReaderConfig) MessageReader {
	return kafka.NewReader(config)
}

// Client represents a high level Kafka Consumer Client
type Client struct {
	logger  *zap.SugaredLogger
	options *Options
	// readerFactory makes it easier to Mock readers as it can be overwritten by Tests
	readerFactory func(config kafka.ReaderConfig) MessageReader
	wg            sync.WaitGroup
}

func NewClient(options *Options) *Client {
	logger := log.New() // NewAtLevel("debug")
	c := &Client{
		options: options,
		logger:  logger,
		// errChan:      make(chan error, 10),
	}
	c.readerFactory = defaultMessageReader
	logger.Debugf("New Client initialized %s@%s consumerGroupId=%s",
		c.options.ConsumerAPIKey, c.options.BootstrapServers, c.options.ConsumerGroupID)
	return c
}

// NewClientFromEnv delegated to NewClient amd returns a properly configured and ready-to-use Client
// that invoked the callback function for every received messages using the default KafkaConsumerTopic
// Spec: See https://github.com/segmentio/kafka-go#reader-
func NewClientFromEnv() (*Client, error) {
	options, err := NewOptionsFromEnv()
	if err != nil {
		return nil, err
	}
	return NewClient(options), nil
}

// Poll uses kafka-go Reader which automatically handles reconnections and offset management,
// and exposes an API that supports asynchronous cancellations and timeouts using Go contexts.
// See https://github.com/segmentio/kafka-go#reader-
// and this nice tutorial https://www.sohamkamani.com/golang/working-with-kafka/
// doneChan chan<- struct{}
func (c *Client) Poll(ctx context.Context, rc kafka.ReaderConfig, msgHandler HandleMessageFunc) error {
	c.applyDefaults(&rc)
	topics := rc.GroupTopics
	c.logger.Infof("Let's consume some yummy Kafka Messages on topic=%s groupID=%s", topics, c.options.ConsumerGroupID)

	r := c.readerFactory(rc)
	defer func() {
		c.logger.Debugf("Post-consume: closing reader stream for topic %s", topics)
		if err := r.Close(); err != nil {
			c.logger.Warnf("Error closing reader stream: %v", err)
		}
		c.wg.Done() // decrement, WaitForClose() will wait for this group as there may be multiple consumers
		c.logger.Debugf("Post-consume: %s ready for shutdown", topics)
	}()
	c.wg.Add(1) // add to wait group to ensure graceful shutdown

	var rcvCount int32 // thx https://github.com/cloudevents/sdk-go/blob/main/samples/kafka/sender-receiver/main.go
	maxReceive := c.options.ConsumerMaxReceive
	for maxReceive < 0 || atomic.AddInt32(&rcvCount, 1) <= maxReceive {
		// SimpleMessageStream reads and return the next message from the r. The method call
		// blocks until a message becomes available, or an error occurs.
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				c.logger.Debugf("Reader has been closed (ctx err: %v)", ctx.Err())
			} else {
				c.logger.Errorf("Error on message read: %v", err)
				return err // really?, or better use c.errChan <- err
			}
			break
		}
		msgHandler(ctx, msg)
	}
	return nil
}

func (c *Client) applyDefaults(rc *kafka.ReaderConfig) {
	dialer := &kafka.Dialer{
		SASLMechanism: plain.Mechanism{
			Username: c.options.ConsumerAPIKey,
			Password: c.options.ConsumerAPISecret,
		},
		Timeout: defaultDialTimeout, // todo make configurable
		TLS:     &tls.Config{MinVersion: tls.VersionTLS12},
	}

	rc.Brokers = []string{c.options.BootstrapServers}
	rc.GroupID = c.options.ConsumerGroupID
	// Topic=   pr.Topic
	// rc.GroupTopics= []string{pr.Topic} // Can listen to multiple topics
	// kafka polls the cluster to check if there is any new data on the topic for the my-group kafka ID,
	// the cluster will only respond if there are at least 10 new bytes of information to send.
	rc.MinBytes = minConsumeBytes
	rc.MaxBytes = maxConsumeBytes
	rc.Dialer = dialer
	// RetentionTime optionally sets the length of time the consumer group will be saved
	rc.RetentionTime = retentionTime
	rc.StartOffset = c.options.StartOffset() // see godoc for details
	rc.CommitInterval = 1 * time.Second      // flushes commits to Kafka every  x seconds
	rc.Logger = LoggerWrapper{delegate: c.logger}
	rc.ErrorLogger = ErrorLoggerWrapper{delegate: c.logger}
}

// WaitForClose blocks until the Consumer WaitGroup counter is zero, or timeout is reached
func (c *Client) WaitForClose() {
	c.logger.Debug("Waiting for Consumer(s) to go down")
	cDone := make(chan struct{})
	go func() {
		defer close(cDone)
		c.wg.Wait()
	}()
	select {
	case <-cDone:
		c.logger.Debugf("All Listeners went down within %v timeout", defaultCloseWaitTimeout)
	case <-time.After(defaultCloseWaitTimeout):
		c.logger.Debugf("Timeout %v reached, stop waiting for listener shutdown", defaultCloseWaitTimeout)
	}
}

// DumpMessage simple handler function that can be used as HandleMessageFunc and simply dumps information
// about the received Kafka Message and the payload container therein
func DumpMessage(_ context.Context, message kafka.Message) {
	fmt.Printf(" kafka.Message: %s %d/%d %s\n", message.Topic, message.Partition, message.Offset, string(message.Value))
}

// AsCloudEvent Helper function to unmarshal Kafka Message into a CloudEvent
func AsCloudEvent(message kafka.Message) (cloudevents.Event, error) {
	// 	request.Headers["content-type"] = cloudevents.ApplicationCloudEventsJSON + "; charset=UTF-8"
	event := cloudevents.NewEvent()
	var cType string
	for _, h := range message.Headers {
		if h.Key == "content-type" {
			cType = string(h.Value)
		}
	}
	if !strings.HasPrefix(cType, cloudevents.ApplicationCloudEventsJSON) {
		return event, fmt.Errorf("%w value %s not supported, expected %s", errInvalidContentType, cType, cloudevents.ApplicationCloudEventsJSON)
	}
	err := json.Unmarshal(message.Value, &event)
	return event, err
}
