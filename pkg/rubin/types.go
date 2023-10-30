package rubin

import (
	b64 "encoding/base64"
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
}

// Client an instance of confluent.Client initialized with the given options
type Client struct {
	options *Options
	logger  zap.SugaredLogger
}

//// Response simple wrapper around a Confluent Rest response
//// Check https://github.com/confluentinc/kafka-rest-sdk-go/blob/master/kafkarestv3/docs/ProduceResponse.md
//// https://github.com/confluentinc/kafka-rest-sdk-go/blob/master/kafkarestv3/model_produce_response.go
// type Response struct {
//	ErrorCode   int        `json:"error_code"`
//	Offset      int32      `json:"offset"`
//	TopicName   string     `json:"topic_name"`
//	PartitionID int32      `json:"partition_id"`
//	Timestamp   *time.Time `json:"timestamp"`
//}

type TopicPayloadElement struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// TopicPayload content for new topic record
// Check https://github.com/confluentinc/kafka-rest-sdk-go/blob/master/kafkarestv3/docs/ProduceRequest.md
type TopicPayload struct {
	Key   TopicPayloadElement `json:"key"`
	Value TopicPayloadElement `json:"value"`
}

func NewTopicPayload(key []byte, data interface{}) TopicPayload {
	return TopicPayload{
		Key:   TopicPayloadElement{Type: "BINARY", Data: b64.StdEncoding.EncodeToString(key)},
		Value: TopicPayloadElement{Type: "JSON", Data: data},
	}
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
