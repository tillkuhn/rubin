package rubin

import (
	b64 "encoding/base64"
	"go.uber.org/zap"
)

// Options keeps the settings to set up client connection.
type Options struct {
	// RestEndpoint of confluent cluster
	RestEndpoint string `yaml:"restEndpoint"`
	// ClusterID of confluent cluster
	ClusterID   string `yaml:"clusterID"`
	ApiKey      string `yaml:"apiKey"`
	ApiPassword string `yaml:"apiPassword"`
	// debug can be only activated from within this package (e.g. for integration testing)
	debug bool
}

// Client an instance of confluent.Client initialized with the given options
type Client struct {
	options *Options
	logger  zap.SugaredLogger
}

// Response simple wrapper around a Confluent Rest response
type Response struct {
	ErrorCode int     `json:"error_code"`
	Offset    float64 `json:"offset"`
	TopicName string  `json:"topic_name"`
}

type TopicPayloadElement struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

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
