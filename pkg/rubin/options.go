package rubin

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

const (
	envconfigDefaultPrefix = "kafka"
)

// Options keeps the settings to set up client connection.
type Options struct {
	// RestEndpoint for kafka rest proxy api
	RestEndpoint string `yaml:"rest_endpoint" default:"" required:"false" desc:"Kafka REST Proxy Endpoint"  split_words:"true"`
	// ClusterID of kafka cluster (which becomes part of the URL)
	ClusterID         string        `yaml:"cluster_id" default:"" required:"false" desc:"Kafka Cluster ID"  split_words:"true"`
	ProducerAPIKey    string        `yaml:"api_key" default:"" required:"false" desc:"Kafka API Key with Producer Privileges"  split_words:"true"`
	ProducerAPISecret string        `yaml:"api_secret" default:"" required:"false" desc:"Kafka API Secret with Producer Privileges"  split_words:"true"`
	HTTPTimeout       time.Duration `yaml:"http_timeout" default:"10s" required:"false" desc:"Timeout for HTTP Client" split_words:"true"`
	DumpMessages      bool          `yaml:"dump_messages" default:"false" required:"false" desc:"Print http request/response to stdout" split_words:"true"`
	LogLevel          string        `yaml:"log_level" default:"info" required:"false" desc:"Min LogLevel debug,info,warn,error" split_words:"true"`
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

// String returns a String representation of the object (but hides sensitive information)
func (o Options) String() string {
	return fmt.Sprintf("%s/%s hasKey=%v hasSecret=%v", o.RestEndpoint, o.ClusterID, len(o.ProducerAPIKey) > 0, len(o.ProducerAPISecret) > 0)
}
