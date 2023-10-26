// Package rubin experiments with Confluent REST API for sync Kafka communication
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

	"github.com/user/rubin/internal/log"
)

// New returns a new Confluent Client for http interaction
func New(options *Options) *Client {
	logger := log.NewAtLevel(os.Getenv("LOG_LEVEL"))
	logger.Infof("Confluent  Client configured for %s@%s (using password: %v)",
		options.ApiKey, options.RestEndpoint, len(options.ApiPassword) > 0)
	return &Client{
		options: options,
		logger:  *logger,
	}
}

// Produce produces records to the given topic, returning delivery reports for each record produced.
// Example URL: https://pkc-zpjg0.eu-central-1.aws.confluent.cloud:443/kafka/v3/clusters/lkc-gqmo5r/topics/public.welcome/records
func (c *Client) Produce(ctx context.Context, topic string, key string, data interface{}) (Response, error) {
	basicAuth := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.options.ApiKey, c.options.ApiPassword)))
	url := fmt.Sprintf("%s/kafka/v3/clusters/%s/topics/%s/records", c.options.RestEndpoint, c.options.ClusterID, topic)
	payload := NewTopicPayload([]byte(key), data)
	payloadJson, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadJson))
	req.Header.Set("Content-Type", "application/json") // don't add ;charset=UTF8 or server will complain
	req.Header.Add("Authorization", "Basic "+basicAuth)
	c.logger.Infof("Send message to %s", url)
	httpClient := &http.Client{Timeout: 5 * time.Second}
	if c.options.debug {
		reqDump, _ := httputil.DumpRequestOut(req, true)
		fmt.Printf("REQUEST:\n%s", string(reqDump)) // only for debug
	}
	var kResponse Response
	res, err := httpClient.Do(req)
	if err != nil {
		return kResponse, err
	}

	defer c.closeSilently(res.Body)
	body, err := io.ReadAll(res.Body)
	kResponse.ErrorCode = res.StatusCode
	if err != nil {
		return kResponse, err
	}

	if res.StatusCode != http.StatusOK {
		return kResponse, fmt.Errorf("unexpected status code %d for %s", res.StatusCode, url)
	}
	if err := json.Unmarshal(body, &kResponse); err != nil {
		return kResponse, fmt.Errorf("unexpected topic api response: %s", string(body))
	}
	c.logger.Infof("Response: offset %v topic %v", kResponse.Offset, kResponse.TopicName)
	return kResponse, nil

}

// CloseSilently avoids "Unhandled error warnings if you use defer to close Resources
func (c *Client) closeSilently(cl io.Closer) {
	if err := cl.Close(); err != nil {
		c.logger.Warn(err.Error())
	}
}
