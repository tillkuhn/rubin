// Package rubin experiments with Kafka REST Proxy API for sync Kafka communication
// Full docs: https://docs.confluent.io/cloud/current/api.html#tag/Records-(v3)
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
	"os"
	"time"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/pkg/errors"

	"github.com/tillkuhn/rubin/internal/log"
)

const defaultTimeout = 30 * time.Second

// New returns a new Rubin Client for http interaction
func New(options *Options) *Client {
	logger := log.NewAtLevel(os.Getenv("LOG_LEVEL"))
	logger.Infow("Kafka REST Proxy Client configured",
		"endpoint", options.RestEndpoint, "useSecret", len(options.APISecret) > 0)
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
func (c *Client) Produce(ctx context.Context, topic string, key string, data interface{}) (kafkarestv3.ProduceResponse, error) {
	defer func() {
		_ = c.logger.Sync() // flushed any buffered log entries
	}()
	basicAuth := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.options.APIKey, c.options.APISecret)))
	url := fmt.Sprintf("%s/kafka/v3/clusters/%s/topics/%s/records", c.options.RestEndpoint, c.options.ClusterID, topic)
	payload := NewTopicPayload([]byte(key), data)
	payloadJSON, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadJSON))
	req.Header.Set("Content-Type", "application/json") // don't add ;charset=UTF8 or server will complain
	req.Header.Add("Authorization", "Basic "+basicAuth)

	c.logger.Infow("Push record", "url", url)
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
		fmt.Printf("Dump HTTP-Response:\n%s\n\n", string(resDump)) // only for debug
	}

	defer c.closeSilently(res.Body)
	body, err := io.ReadAll(res.Body)

	// kResp. = res.StatusCode
	if err != nil {
		return kResp, err
	}

	// deal with err113: do not define dynamic errors, (re-)use wrapped static errors instead:
	responseError := errors.New("unexpected rest proxy api response")

	if res.StatusCode != http.StatusOK {
		return kResp, errors.Wrap(responseError, fmt.Sprintf("unexpected http response status code %d for %s", res.StatusCode, url))
	}
	if err := json.Unmarshal(body, &kResp); err != nil {
		return kResp, errors.Wrap(err, fmt.Sprintf("unexpected topic api response: %s", string(body)))
	}
	// if kResp.ErrorCode != http.StatusOK {
	//	return kResp, errors.Wrap(responseError, fmt.Sprintf("unexpected kafka response error code %d for %s", kResp.ErrorCode, url))
	//}
	c.logger.Infow("Record committed", "key", kResp.Key, "topic", kResp.TopicName, "offset", kResp.Offset, "partition", kResp.PartitionId)
	return kResp, nil
}

// CloseSilently avoids "Unhandled error warnings if you use defer to close Resources
func (c *Client) closeSilently(cl io.Closer) {
	if err := cl.Close(); err != nil {
		c.logger.Warn(err.Error())
	}
}
