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
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/google/uuid"

	"github.com/pkg/errors"
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
	// logger     *zerolog.Logger
}

// NewClient returns a new Rubin Client for http interaction
func NewClient(options *Options) *Client {
	// logger := zerolog.Logger{}//log.NewAtLevel(options.LogLevel)
	// logger.Info().Msgf("Kafka REST Proxy Client configured with endpoint=%s secret=%v",
	//	options.RestEndpoint, len(options.ProducerAPISecret) > 0)
	if options.HTTPTimeout.Seconds() < 1 {
		// logger.Printf("Timeout duration is zero or too low, using default %v", defaultTimeout)
		options.HTTPTimeout = defaultTimeout
	}

	return &Client{
		options:    options,
		httpClient: &http.Client{Timeout: options.HTTPTimeout},
		// logger:     &logger,
	}
}

// NewClientFromEnv Convenience function using default envconfig prefix for
//
//	opts, err := NewOptionsFromEnv()
//	client := NewClient(opts)
func NewClientFromEnv() (*Client, error) {
	opts, err := NewOptionsFromEnv()
	if err != nil {
		return nil, err
	}
	return NewClient(opts), err
}

// LogLevel allows dynamic configuration of LogLevel after the client has been initialized
// func (c *Client) LogLevel(levelStr string) {
//	logger = log.NewAtLevel(levelStr)
//}

// String representation of the client instance
func (c *Client) String() string {
	return fmt.Sprintf("rubin-http-client@%s", c.options.String())
}

// RecordRequest holds the data to build the Kafka Message Payload plus optional Key and Headers
type RecordRequest struct {
	Topic   string
	Data    interface{}
	Key     string
	Headers map[string]string
	// AsCloudEvent section for CloudEvents specific attributes
	AsCloudEvent bool
	Source       string
	Type         string
	Subject      string
}

// RecordResponse Wrapper around the external RecordResponse
type RecordResponse struct {
	ErrorCode int `json:"error_code"`
	kafkarestv3.ProduceResponse
}

// Produce produces a Kafka Record into the given Topic
func (c *Client) Produce(ctx context.Context, request RecordRequest) (RecordResponse, error) {
	logger := log.Ctx(ctx).With().Str("logger", "producer").Logger()
	// defer log.SyncSilently(logger)
	keyData := c.messageKeyData(request.Key)
	url := c.options.RecordEndpoint(request.Topic)

	var prodResp RecordResponse
	if request.AsCloudEvent {
		// wrap data into a Cloud Event
		ce, err := NewCloudEvent(request.Source, request.Type, request.Data)
		if err != nil {
			return prodResp, err
		}
		ce.SetSubject(request.Subject)
		request.Data = ce
	}

	valueType, valueData, err := transformPayload(request.Data)
	if err != nil {
		return prodResp, fmt.Errorf("%w: unable to extract paylos (%s)", errClientResponse, err.Error())
	}
	// handle message headers, add content type for cloud events
	if request.Headers == nil {
		request.Headers = map[string]string{}
	}

	// todo improve CE detection, use alternative content-type headers for JSON and STRING
	_, isCE := request.Data.(event.Event)
	if isCE {
		request.Headers["content-type"] = cloudevents.ApplicationCloudEventsJSON + "; charset=UTF-8"
	}
	apiHeaders := messageHeaders(request.Headers)

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
	req.Header.Add("Authorization", "Basic "+c.options.BasicAuth())

	logger.Info().Msgf("TopicURL=%s type=%s ce=%v len=%d hd=%d", url,
		fmt.Sprintf("%T", request.Data), request.AsCloudEvent, len(payloadJSON), len(apiHeaders),
	)
	c.checkDumpRequest(req)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return prodResp, fmt.Errorf("%w: cannot send http request %s", errClientResponse, err.Error())
	}
	c.checkDumpResponse(res)
	prodResp, err = parseResponse(res)
	if err != nil {
		return prodResp, err
	}

	logger.Info().Msgf("RecordRequest successfully committed code=%d key=%v topic=%s offset=%d part=%d", prodResp.ErrorCode, prodResp.Key, prodResp.TopicName, prodResp.Offset, prodResp.PartitionId)

	return prodResp, nil
}

func parseResponse(res *http.Response) (RecordResponse, error) {
	defer closeSilently(res.Body)
	var prodResp RecordResponse
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return prodResp, fmt.Errorf("%w: cannot parse response body %s", errClientResponse, err.Error())
	}

	if res.StatusCode != http.StatusOK {
		return prodResp, fmt.Errorf("%w: unexpected http status code %d", errClientResponse, res.StatusCode)
	}

	if err := json.Unmarshal(body, &prodResp); err != nil {
		return prodResp, errors.Wrap(err, fmt.Sprintf("unexpected topic api response: %s", string(body)))
	}
	if prodResp.ErrorCode != http.StatusOK {
		// error_code must be 200, other values indicate an error but could be also 5 digit (e.g. 40301)
		return prodResp, fmt.Errorf("%w: unexpected error_code %d in response %s", errClientResponse, prodResp.ErrorCode, string(body))
	}
	return prodResp, nil
}

func (c *Client) messageKeyData(key string) interface{} {
	if key == "" {
		key = uuid.New().String()
		// logger.Printf("Using generated message key %s", key)
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
// returned valueType is either STRING or JSON
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
		// marshals to string value "value":{"type":"STRING","data":"Hello String!"} }
		valueData = data
	default:
		valueType = "JSON"
		// marshals to json in json "value":{"type":"JSON","data":{"action":"update/event",
		valueData = data
	}
	return valueType, valueData, nil
}

func (c *Client) checkDumpRequest(req *http.Request) {
	if c.options.DumpMessages {
		reDump, _ := httputil.DumpRequest(req, true)
		sanitizedReq := strings.Replace(string(reDump), c.options.BasicAuth(), "************", 1)
		fmt.Printf("Dump HTTP-RecordRequest:\n%s", sanitizedReq) // only for debug
	}
}
func (c *Client) checkDumpResponse(res *http.Response) {
	if c.options.DumpMessages {
		resDump, _ := httputil.DumpResponse(res, true)

		fmt.Printf("\nDump HTTP-RecordResponse:\n%s", string(resDump)) // only for debug
	}
}

// CloseSilently avoids "Unhandled error warnings if you use defer to close Resources
func closeSilently(cl io.Closer) {
	if err := cl.Close(); err != nil {
		log.Warn().Msg(err.Error())
	}
}

// Must helper, see https://stackoverflow.com/a/73584801/4292075
// a generic version of https://go.dev/src/text/template/helper.go?s=576:619
// panics if the error is non-nil and intended for use in variable initializations
//
//	opts := Must[*Options](NewOptionsFromEnv())
func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}
