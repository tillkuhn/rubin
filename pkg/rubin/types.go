package rubin

import (
	"time"

	"go.uber.org/zap"
)

// Options keeps the settings to set up client connection.
type Options struct {
	// RestEndpoint of confluent cluster
	RestEndpoint string `yaml:"rest_endpoint" default:"" required:"false" desc:"Kafka REST Proxy Endpoint"  split_words:"true"`
	// ClusterID of confluent cluster
	ClusterID    string        `yaml:"cluster_id" default:"" required:"false" desc:"Kafka Cluster ID"  split_words:"true"`
	APIKey       string        `yaml:"api_key" default:"" required:"false" desc:"Kafka API Key with Producer Privileges"  split_words:"true"`
	APISecret    string        `yaml:"api_secret" default:"" required:"false" desc:"Kafka API Secret with Producer Privileges"  split_words:"true"`
	HTTPTimeout  time.Duration `yaml:"http_timeout" default:"10s" required:"false" desc:"Timeout for HTTP Client" split_words:"true"`
	DumpMessages bool          `yaml:"dump_messages" default:"false" required:"false" desc:"Print http request/response to stdout" split_words:"true"`
	LogLevel     string        `yaml:"log_level" default:"info" required:"false" desc:"Min LogLevel debug,info,warn,error" split_words:"true"`
}

// Client an instance of confluent.Client initialized with the given options
type Client struct {
	options *Options
	logger  zap.SugaredLogger
}

type Event struct {
	Action  string `json:"action,omitempty"`
	Message string `json:"message,omitempty"`
	// Time.MarshalJSON returns
	// "The time is a quoted string in RFC 3339 format, with sub-second precision added if present."
	Time     time.Time `json:"time,omitempty"`
	Source   string    `json:"source,omitempty"`
	EntityID string    `json:"entityId,omitempty"`
}

/*
payload := `{
	 "key": {
	   "type": "BINARY",
	   "data": "Zm9vYmFy"
	 },
	 "value": {
	   "type": "JSON",
	   "data": "Bonjour le monde, de nouveau! Let's go!'"
	 }
	}`
*/
