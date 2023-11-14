package polly

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/segmentio/kafka-go"
)

const envconfigDefaultPrefix = "kafka"

// Options Kafka Context params populated by envconfig in NewClientFromEnv...()
type Options struct {
	BootstrapServers   string `required:"false" default:"localhost:9092" desc:"Kafka Bootstrap server(s)" split_words:"true"`
	ProducerClientID   string `required:"false" default:"kafkaClient" desc:"Client Id for Message Producer" split_words:"true"`
	ConsumerAPIKey     string `required:"false" default:"" desc:"Kafka API Key Key for consumer (user)"  split_words:"true"`
	ConsumerAPISecret  string `required:"false" default:"" desc:"Kafka API Secret for consumer (password)" split_words:"true"`
	ConsumerGroupID    string `required:"false" default:"app.local3" desc:"Used as default id for KafkaConsumerGroups" split_words:"true"`
	ConsumerMaxReceive int32  `required:"false" default:"-1" desc:"Max num of received messages, default -1 (unlimited), useful for dev" split_words:"true"`
	ConsumerStartLast  bool   `required:"false" default:"false" desc:"Whether to start consuming at the last offset (default: first)" split_words:"true"`
	Debug              bool   `default:"false" desc:"Debug mode, registers logger for kafka packages" split_words:"true"`
}

// NewOptionsFromEnv uses environment configuration with default prefix "kafka" to init Options
func NewOptionsFromEnv() (*Options, error) {
	return NewOptionsFromEnvWithPrefix(envconfigDefaultPrefix)
}

// NewOptionsFromEnvWithPrefix same as NewOptionsFromEnv but allows custom prefix
func NewOptionsFromEnvWithPrefix(prefix string) (*Options, error) {
	var options Options
	if err := envconfig.Process(prefix, &options); err != nil {
		return nil, err
	}
	return &options, nil
}

// StartOffset provides the reader options depending on ConsumerStartLast (true == first, else last)
// LastOffset  int64 = -1 // The most recent offset available for a partition.
// FirstOffset int64 = -2 // The least recent offset available for a partition.
func (o Options) StartOffset() int64 {
	if o.ConsumerStartLast {
		return kafka.LastOffset
	}
	return kafka.FirstOffset
}
