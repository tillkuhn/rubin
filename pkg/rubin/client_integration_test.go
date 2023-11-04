//go:build integration

// Experiment with "Separate Your Go Tests with Build Tags (go:build)"
// https://mickey.dev/posts/go-build-tags-testing/
// go test --tags=integration ./...
package rubin

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	intOptionsFile = ".test-int-options.yaml"
	topicUnderTest = "public.hello"
)

func TestProduceMessageRealConfluentAPI(t *testing.T) {

	ctx := context.Background()
	// id := uuid.New().String()

	intOptions, err := initOptions()
	if err != nil {
		t.Log(fmt.Sprintf("This test requires gitignored file %s in package directory with confluent producer api token", intOptionsFile))
		t.FailNow()
	}
	payloadData := []byte("Bonjour le monde, de nouveau! Let's go!'")

	// This should succeed
	cc := NewClient(intOptions)
	hm := map[string]string{"heading": "for tomorrow"}
	rd := Request{topicUnderTest, payloadData, "134", hm}
	resp, err := cc.Produce(ctx, rd)
	assert.NoError(t, err)
	assert.Greater(t, resp.Offset, int32(0))
	assert.Equal(t, topicUnderTest, resp.TopicName)
	assert.Equal(t, http.StatusOK, resp.ErrorCode)

	// this should also succeed
	newCar := struct {
		Make    string `json:"make"`
		Model   string
		Mileage int
	}{
		Make:    "BMW",
		Model:   "Taurus",
		Mileage: 200000,
	}
	rd.Data = newCar
	resp, err = cc.Produce(ctx, rd)
	//t.Log(string(resp))
	// assert.Equal(t, http.StatusOK, resp.ErrorCode)
	assert.NoError(t, err)

	event, err := NewCloudEvent("//testing/ci-test", "int.test", hm)
	assert.NoError(t, err)

	rd.Data = event
	resp, err = cc.Produce(ctx, rd)
	assert.NoError(t, err)

	// This will fail (wrong password)
	cc = NewClient(&Options{
		RestEndpoint:      intOptions.RestEndpoint,
		ClusterID:         intOptions.ClusterID,
		ProducerAPIKey:    "nobody",
		ProducerAPISecret: "failed",
		HTTPTimeout:       1 * time.Second,
		DumpMessages:      true,
	})
	resp, err = cc.Produce(ctx, rd)
	// assert.Equal(t, http.StatusUnauthorized, resp.ErrorCode)
	assert.ErrorContains(t, err, "unexpected http")
	assert.Empty(t, resp.TopicName)

	t.Log("Real Kafka Integration Test is happy")

}

func initOptions() (*Options, error) {
	yamlConfig, err := os.ReadFile(intOptionsFile)
	if err != nil {
		return nil, err
	}
	var intOptions Options
	err = yaml.Unmarshal([]byte(yamlConfig), &intOptions)
	return &intOptions, err
}
