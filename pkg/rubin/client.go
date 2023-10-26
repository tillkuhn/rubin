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

	"github.com/pkg/errors"

	"github.com/user/rubin/internal/log"
)

// New returns a new Rubin Client for http interaction
func New(options *Options) *Client {
	logger := log.NewAtLevel(os.Getenv("LOG_LEVEL"))
	logger.Infow("Rubin  Client configured",
		"endpoint", options.RestEndpoint, "useSecret", len(options.APIPassword) > 0)
	return &Client{
		options: options,
		logger:  *logger,
	}
}

// Produce produces records to the given topic, returning delivery reports for each record produced.
// Example URL: https://pkc-zpjg0.eu-central-1.aws.confluent.cloud:443/kafka/v3/clusters/lkc-gqmo5r/topics/public.welcome/records
func (c *Client) Produce(ctx context.Context, topic string, key string, data interface{}) (Response, error) {
	basicAuth := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.options.APIKey, c.options.APIPassword)))
	url := fmt.Sprintf("%s/kafka/v3/clusters/%s/topics/%s/records", c.options.RestEndpoint, c.options.ClusterID, topic)
	payload := NewTopicPayload([]byte(key), data)
	payloadJSON, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadJSON))
	req.Header.Set("Content-Type", "application/json") // don't add ;charset=UTF8 or server will complain
	req.Header.Add("Authorization", "Basic "+basicAuth)

	c.logger.Infof("Push record to %s", url)
	httpClient := &http.Client{Timeout: c.options.HTTPTimeout}
	if c.options.debug {
		reqDump, _ := httputil.DumpRequestOut(req, true)
		fmt.Printf("REQUEST:\n%s", string(reqDump)) // only for debug
	}
	var kResp Response
	res, err := httpClient.Do(req)
	if err != nil {
		return kResp, err
	}

	defer c.closeSilently(res.Body)
	body, err := io.ReadAll(res.Body)
	kResp.ErrorCode = res.StatusCode
	if err != nil {
		return kResp, err
	}

	if res.StatusCode != http.StatusOK {
		return kResp, errors.Wrap(err, fmt.Sprintf("unexpected status code %d for %s", res.StatusCode, url))
	}
	if err := json.Unmarshal(body, &kResp); err != nil {
		return kResp, errors.Wrap(err, fmt.Sprintf("unexpected topic api response: %s", string(body)))
	}
	c.logger.Infow("Record committed", "status", kResp.ErrorCode, "offset", kResp.Offset, "topic", kResp.TopicName)
	return kResp, nil
}

// CloseSilently avoids "Unhandled error warnings if you use defer to close Resources
func (c *Client) closeSilently(cl io.Closer) {
	if err := cl.Close(); err != nil {
		c.logger.Warn(err.Error())
	}
}
