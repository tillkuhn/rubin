package rubin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	sampleDir = "../../testdata"
	clusterID = "abc-r2d2"
)

func init() {
	// log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func TestProduceMessageOK(t *testing.T) {
	srv := serverMock(200) // "https://pkc-zpjg0.eu-central-1.aws.confluent.cloud:443"
	defer srv.Close()
	cc := New(&Options{srv.URL, clusterID, "test.key", "test.pw", 5 * time.Second, false})
	// strings.NewReader("hello world")
	resp, err := cc.Produce(context.Background(), "public.welcome", "1234", "Hello Hase!")
	assert.NoError(t, err)
	assert.Equal(t, int32(42), resp.Offset)
	assert.NotNil(t, resp.Timestamp)
	assert.Equal(t, http.StatusOK, resp.ErrorCode)
}

func serverMock(responseCode int) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc(
		fmt.Sprintf("/kafka/v3/clusters/%s/topics/public.welcome/records", clusterID),
		mockHandler(fmt.Sprintf("%s/response-%d.json", sampleDir, responseCode)),
	)
	srv := httptest.NewServer(handler)
	return srv
}

func mockHandler(responseFile string) func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		respBytes, err := os.ReadFile(responseFile)
		if err != nil {
			panic(err.Error())
		}
		_, _ = w.Write(respBytes)
	}
}
