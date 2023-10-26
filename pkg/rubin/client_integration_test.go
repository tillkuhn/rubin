//go:build integration

// Experiment with "Separate Your Go Tests with Build Tags (go:build)"
// https://mickey.dev/posts/go-build-tags-testing/
// go test --tags=integration ./...
package rubin

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"testing"
)

const (
	intOptionsFile = ".integration-test-options.yaml"
)

func TestProduceMessageRealConfluentAPI(t *testing.T) {
	topic := "public.hello"

	ctx := context.Background()
	id := uuid.New().String()

	intOptions, err := initOptions()
	intOptions.debug = true
	if err != nil {
		t.Log(fmt.Sprintf("This test requires gitignored file %s in package directory with confluent producer api token", intOptionsFile))
		t.FailNow()
	}
	payloadData := []byte("Bonjour le monde, de nouveau! Let's go!'")

	// This should succeed
	cc := New(intOptions)
	resp, err := cc.Produce(ctx, topic, id, payloadData)
	assert.Greater(t, resp.Offset, float64(0))
	assert.Equal(t, topic, resp.TopicName)
	assert.NoError(t, err)

	newCar := struct {
		Make    string `json:"make"`
		Model   string
		Mileage int
	}{
		Make:    "BMW",
		Model:   "Taurus",
		Mileage: 200000,
	}
	resp, err = cc.Produce(ctx, topic, "my-car-123", newCar)
	//t.Log(string(resp))
	assert.Equal(t, http.StatusOK, resp.ErrorCode)
	assert.NoError(t, err)

	// This will fail (wrong password
	cc = New(&Options{intOptions.RestEndpoint, intOptions.ClusterID, "nobody", "failed", false})
	resp, err = cc.Produce(ctx, topic, id, payloadData)
	assert.Equal(t, http.StatusUnauthorized, resp.ErrorCode)
	assert.ErrorContains(t, err, "unexpected status code")
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
