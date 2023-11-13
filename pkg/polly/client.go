package polly

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

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

// HandleMessageFunc consumer will pass received messages to a function that matches this type
type HandleMessageFunc func(ctx context.Context, message kafka.Message)

// MessageReader interface that makes it easy to mock the real kafka.Reader for testing purposes
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

// ConsumeRequest to be passed to Consumer method, identifies the Kafka Topic
// and the ConsumerFunction to handle the messages
type ConsumeRequest struct {
	Topic   string
	Handler HandleMessageFunc
}

func NewClient(options *Options) *Client {
	logger := log.NewAtLevel("debug") // log.With().Str("logger", "kafka-consumer🐛").Logger()
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

// NewClientFromEnv returns a properly configured and ready-to-use Client
// that invoked the callback function for every received messages using the default KafkaConsumerTopic
// Spec: See https://github.com/segmentio/kafka-go#reader-
func NewClientFromEnv() (*Client, error) {
	options, err := NewOptionsFromEnv()
	if err != nil {
		return nil, err
	}
	return NewClient(options), nil
}

// CloseWait blocks until the Consumer WaitGroup counter is zero, or timeout is reached
func (c *Client) CloseWait() {
	c.logger.Debug("Waiting for Consumer(s) to go down")
	cTimeout := make(chan struct{})
	go func() {
		defer close(cTimeout)
		c.wg.Wait()
	}()
	select {
	case <-cTimeout:
		c.logger.Debugf("All Listeners went down within %v timeout", defaultCloseWaitTimeout)
	case <-time.After(defaultCloseWaitTimeout):
		c.logger.Debugf("Timeout %v reached, stop waiting for listener shutdown", defaultCloseWaitTimeout)
	}
}

// Consume uses kafka-go Reader which automatically handles reconnections and offset management,
// and exposes an API that supports asynchronous cancellations and timeouts using Go contexts.
// See https://github.com/segmentio/kafka-go#reader-
// and this nice tutorial https://www.sohamkamani.com/golang/working-with-kafka/
// doneChan chan<- struct{}
func (c *Client) Consume(ctx context.Context, req ConsumeRequest) error {
	c.logger.Infof("Let's consume some yummy Kafka Messages on topic=%s groupID=%s", req.Topic, c.options.ConsumerGroupID)
	dialer := &kafka.Dialer{
		SASLMechanism: plain.Mechanism{
			Username: c.options.ConsumerAPIKey,
			Password: c.options.ConsumerAPISecret,
		},
		Timeout: defaultDialTimeout, // todo make configurable
		TLS:     &tls.Config{MinVersion: tls.VersionTLS12},
	}

	r := c.readerFactory(kafka.ReaderConfig{
		Brokers: []string{c.options.BootstrapServers},
		GroupID: c.options.ConsumerGroupID,
		// Topic:   req.Topic,
		GroupTopics: []string{req.Topic}, // Can listen to multiple topics
		// kafka polls the cluster to check if there is any new data on the topic for the my-group kafka ID,
		// the cluster will only respond if there are at least 10 new bytes of information to send.
		MinBytes: minConsumeBytes,
		MaxBytes: maxConsumeBytes,
		Dialer:   dialer,
		// RetentionTime optionally sets the length of time the consumer group will be saved
		RetentionTime:  retentionTime,
		StartOffset:    c.options.StartOffset(), // see godoc for details
		CommitInterval: 1 * time.Second,         // flushes commits to Kafka every  x seconds
		Logger:         LoggerWrapper{delegate: c.logger},
		ErrorLogger:    ErrorLoggerWrapper{delegate: c.logger},
	})
	defer func() {
		c.logger.Debugf("Post-consume: closing reader stream for topic %s", req.Topic)
		if err := r.Close(); err != nil {
			c.logger.Warnf("Error closing reader stream: %v", err)
		}
		c.wg.Done() // decrement, CloseWait() will wait for this group and there may be multiple consumers
		c.logger.Debugf("Post-consume: %s ready for shutdown", req.Topic)
	}()
	c.wg.Add(1) // add to wait group to ensure graceful shutdown

	var rcvCount int32 // thx https://github.com/cloudevents/sdk-go/blob/main/samples/kafka/sender-receiver/main.go
	maxReceive := c.options.ConsumerMaxReceive
	for maxReceive < 0 || atomic.AddInt32(&rcvCount, 1) <= maxReceive {
		// ReadMessage reads and return the next message from the r. The method call
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
		req.Handler(ctx, msg)
	}
	return nil
}

// DumpMessage simple handler function that can be used as HandleMessageFunc and simply dumps information
// about the received Kafka Message and the payload container therein
func DumpMessage(_ context.Context, message kafka.Message) {
	fmt.Printf(" kafka.Message: %s %d/%d %s\n", message.Topic, message.Partition, message.Offset, string(message.Value))
}
