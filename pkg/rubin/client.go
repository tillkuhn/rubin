package rubin

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/google/uuid"

	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/tillkuhn/rubin/internal/log"
)

// defaultTimeout for http communication
const defaultTimeout = 30 * time.Second

// errClient addresses linter err113: do not define dynamic errors, use wrapped static errors instead
// use like this:  fmt.Errorf("%w: unexpected http status code %d for %s", errClientResponse, res.StatusCode, url)
var errClientResponse = errors.New("kafka client response error")

// Client to communicate with the Kafka REST Endpoint Topic/Records API
type Client struct {
	options    *Options
	httpClient *http.Client
	logger     zap.SugaredLogger
}

// NewClient returns a new Rubin Client for http interaction
func NewClient(options *Options) *Client {
	logger := log.NewAtLevel(options.LogLevel)
	logger.Infow("Kafka REST Proxy Client configured",
		"endpoint", options.RestEndpoint, "useSecret", len(options.ProducerAPISecret) > 0)
	if options.HTTPTimeout.Seconds() < 1 {
		logger.Debugf("Timeout duration is zero or too low, using default %v", defaultTimeout)
		options.HTTPTimeout = defaultTimeout
	}

	return &Client{
		options:    options,
		httpClient: &http.Client{Timeout: options.HTTPTimeout},
		logger:     *logger,
	}
}

func (c *Client) String() string {
	return fmt.Sprintf("rubin-http-client@%s/%s", c.options.RestEndpoint, c.options.ClusterID)
}

func (c *Client) Produce(ctx context.Context, topic string, key string, data interface{}, hm map[string]string) (kafkarestv3.ProduceResponse, error) {
	defer func() {
		_ = c.logger.Sync() // make sure any buffered log entries are flushed when Produce returns
	}()
	keyData := c.messageKeyData(key)
	basicAuth := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.options.ProducerAPIKey, c.options.ProducerAPISecret)))
	url := fmt.Sprintf("%s/kafka/v3/clusters/%s/topics/%s/records", c.options.RestEndpoint, c.options.ClusterID, topic)
	apiHeaders := messageHeaders(hm)

	var kResp kafkarestv3.ProduceResponse
	valueType, valueData, err := transformPayload(data)
	if err != nil {
		return kResp, fmt.Errorf("%w: unable to extract paylos (%s)", errClientResponse, err.Error())
	}
	ts := time.Now().Round(time.Second)
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
	c.checkDumpRequest(req)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return kResp, fmt.Errorf("%w: cannot send http request %s", errClientResponse, err.Error())
	}
	c.checkDumpResponse(res)

	defer closeSilently(res.Body)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return kResp, fmt.Errorf("%w: cannot parse response body %s", errClientResponse, err.Error())
	}

	if res.StatusCode != http.StatusOK {
		return kResp, fmt.Errorf("%w: unexpected http status code %d for %s", errClientResponse, res.StatusCode, url)
	}

	if err := json.Unmarshal(body, &kResp); err != nil {
		return kResp, errors.Wrap(err, fmt.Sprintf("unexpected topic api response: %s", string(body)))
	}
	// todo check if kResp.ErrorCode != http.StatusOK (  "error_code": 200 ), but ErrorCode ist not in ProduceResponse
	c.logger.Infow("Record successfully committed", "key", kResp.Key, "topic", kResp.TopicName, "offset", kResp.Offset, "partition", kResp.PartitionId)

	return kResp, nil
}

func (c *Client) messageKeyData(key string) interface{} {
	if key == "" {
		key = uuid.New().String()
		c.logger.Debugf("Using generated message key %s", key)
	}
	keyData := b64.StdEncoding.EncodeToString([]byte(key))
	return keyData
}

func messageHeaders(hm map[string]string) []kafkarestv3.ProduceRequestHeader {
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
	return apiHeaders
}

// transformPayload inspects the payload, determines the valueType and handles JSON Strings
func transformPayload(data interface{}) (valueType string, valueData interface{}, err error) {
	s, isString := data.(string)
	switch {
	case isString && json.Valid([]byte(s)):
		valueType = "JSON"
		err := json.Unmarshal([]byte(s), &valueData)
		if err != nil {
			return valueType, valueData, err
		}
	case isString:
		valueType = "STRING"
		// "value":{"type":"STRING","data":"Hello String!"}
		valueData = data
	default:
		valueType = "JSON"
		// real json "value":{"type":"JSON","data":{"action":"update/event",
		valueData = data
	}
	return valueType, valueData, nil
}

func (c *Client) checkDumpRequest(req *http.Request) {
	if c.options.DumpMessages {
		reDump, _ := httputil.DumpRequest(req, true)
		fmt.Printf("Dump HTTP-Request:\n%s", string(reDump)) // only for debug
	}
}
func (c *Client) checkDumpResponse(res *http.Response) {
	if c.options.DumpMessages {
		resDump, _ := httputil.DumpResponse(res, true)
		fmt.Printf("Dump HTTP-Response:\n%s", string(resDump)) // only for debug
	}
}

// CloseSilently avoids "Unhandled error warnings if you use defer to close Resources
func closeSilently(cl io.Closer) {
	if err := cl.Close(); err != nil {
		log.NewAtLevel("warn").Warn(err.Error())
	}
}
