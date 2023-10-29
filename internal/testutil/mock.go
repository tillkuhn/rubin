package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
)

const (
	sampleDir = "../../testdata"
	ClusterID = "abc-r2d2"
)

func ServerMock(responseCode int) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc(
		fmt.Sprintf("/kafka/v3/clusters/%s/topics/public.welcome/records", ClusterID),
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
