// Package rubin experiments with Kafka REST Proxy API for sync Kafka communication
// Full docs: https://docs.confluent.io/cloud/current/api.html#tag/Records-(v3)
package rubin

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/pkg/errors"

	"github.com/tillkuhn/rubin/internal/log"
)

const defaultTimeout = 30 * time.Second

// errClient addresses linter err113: do not define dynamic errors, use wrapped static errors instead
// use like this:  fmt.Errorf("%w: unexpected http status code %d for %s", errClientResponse, res.StatusCode, url)
var errClientResponse = errors.New("kafka client response error")

// New returns a new Rubin Client for http interaction
func New(options *Options) *Client {
	logger := log.NewAtLevel(options.LogLevel)
	logger.Infow("Kafka REST Proxy Client configured",
		"endpoint", options.RestEndpoint, "useSecret", len(options.ProducerAPISecret) > 0)
	if options.HTTPTimeout.Seconds() < 1 {
		logger.Debugf("Timeout duration is zero or too low, using default %v", defaultTimeout)
		options.HTTPTimeout = defaultTimeout
	}

	return &Client{
		options: options,
		logger:  *logger,
	}
}

// Produce produces records to the given topic, returning delivery reports for each record produced.
// Example URL: https://pkc-zpjg0.eu-central-1.aws.confluent.cloud:443/kafka/v3/clusters/lkc-gqmo5r/topics/public.welcome/records
// See also: https://github.com/confluentinc/kafka-rest#produce-records-with-json-data
// and https://github.com/confluentinc/kafka-rest#produce-records-with-string-data
// and https://docs.confluent.io/platform/current/kafka-rest/api.html
// "If your data is JSON, you can use json as the embedded format and embed it directly:"
func (c *Client) Produce(ctx context.Context, topic string, key string, data interface{}, hm map[string]string) (kafkarestv3.ProduceResponse, error) {
	defer func() {
		_ = c.logger.Sync() // make sure any buffered log entries are flushed when Produce returns
	}()

	if key == "" {
		key = uuid.New().String()
		c.logger.Debugf("Using generated message key %s", key)
	}
	ts := time.Now().Round(time.Second) // make sure we round to .SSS
	basicAuth := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.options.ProducerAPIKey, c.options.ProducerAPISecret)))
	url := fmt.Sprintf("%s/kafka/v3/clusters/%s/topics/%s/records", c.options.RestEndpoint, c.options.ClusterID, topic)
	var keyData interface{}
	if len(key) > 0 {
		keyData = b64.StdEncoding.EncodeToString([]byte(key))
	}

	// handle prealloc warning (https://stackoverflow.com/a/59734761/4292075)
	apiHeaders := make([]kafkarestv3.ProduceRequestHeader, len(hm))
	mapCnt := 0
	for k, v := range hm {
		val := b64.StdEncoding.EncodeToString([]byte(v))
		apiHeaders[mapCnt] = kafkarestv3.ProduceRequestHeader{
			Name:  k,
			Value: &val,
		}
		mapCnt++
	}

	valueType, valueData := c.transformPayload(data)
	payload := kafkarestv3.ProduceRequest{
		// PartitionId: nil, // not needed
		Headers: apiHeaders,
		Key: &kafkarestv3.ProduceRequestData{
			Type: "BINARY",
			Data: &keyData,
		},
		Value: &kafkarestv3.ProduceRequestData{
			Type: valueType, // String or JSON
			Data: &valueData,
		},
		Timestamp: &ts,
	}
	payloadJSON, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadJSON))
	req.Header.Set("Content-Type", "application/json") // don't add ;charset=UTF8 or server will complain
	req.Header.Add("Authorization", "Basic "+basicAuth)

	c.logger.Infow("Push record", "url", url, "msg-len", len(payloadJSON), "headers", len(apiHeaders))
	httpClient := &http.Client{Timeout: c.options.HTTPTimeout}
	if c.options.DumpMessages {
		reqDump, _ := httputil.DumpRequestOut(req, true)
		fmt.Printf("Dump HTTP-Request:\n%s\n\n", string(reqDump)) // only for debug
	}
	var kResp kafkarestv3.ProduceResponse
	res, err := httpClient.Do(req)
	if err != nil {
		return kResp, err
	}
	if c.options.DumpMessages {
		resDump, _ := httputil.DumpResponse(res, true)
		fmt.Printf("Dump HTTP-Response:\n%s", string(resDump)) // only for debug
	}

	defer c.closeSilently(res.Body)
	body, err := io.ReadAll(res.Body)

	// kResp. = res.StatusCode
	if err != nil {
		return kResp, err
	}

	if res.StatusCode != http.StatusOK {
		return kResp, fmt.Errorf("%w: unexpected http status code %d for %s", errClientResponse, res.StatusCode, url)
	}
	if err := json.Unmarshal(body, &kResp); err != nil {
		return kResp, errors.Wrap(err, fmt.Sprintf("unexpected topic api response: %s", string(body)))
	}
	// if kResp.ErrorCode != http.StatusOK {
	//	return kResp, errors.Wrap(responseError, fmt.Sprintf("unexpected kafka response error code %d for %s", kResp.ErrorCode, url))
	//}
	c.logger.Infow("Record successfully committed", "key", kResp.Key, "topic", kResp.TopicName, "offset", kResp.Offset, "partition", kResp.PartitionId)
	return kResp, nil
}

// transformPayload inspects the payload, determines the valueType and handles JSON Strings
func (c *Client) transformPayload(data interface{}) (valueType string, valueData interface{}) {
	s, isString := data.(string)
	// .logger.Infof("Interface Data: %T", data) // returns *string, string or rubin.Event ...
	switch {
	case isString && json.Valid([]byte(s)):
		c.logger.Debugf("Record value is a string that contains valid JSON, using unmarshal")
		valueType = "JSON"
		err := json.Unmarshal([]byte(s), &valueData)
		if err != nil {
			c.logger.Errorf("%v", err)
		}
	case isString:
		valueType = "STRING"
		// "value":{"type":"STRING","data":"Hello String!"}
		c.logger.Debugf("Record value is a simple string")
		valueData = data
	default:
		valueType = "JSON"
		// real json "value":{"type":"JSON","data":{"action":"update/event",
		c.logger.Debug("Record value will be marshalled as embedded struct in ProduceRequest")
		valueData = data
	}
	return valueType, valueData
}

// CloseSilently avoids "Unhandled error warnings if you use defer to close Resources
func (c *Client) closeSilently(cl io.Closer) {
	if err := cl.Close(); err != nil {
		c.logger.Warn(err.Error())
	}
}
