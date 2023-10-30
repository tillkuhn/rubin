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
	Topic     = "public.welcome"
)

func ServerMock(responseCode int) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc(
		fmt.Sprintf("/kafka/v3/clusters/%s/topics/%s/records", ClusterID, Topic),
		mockHandler(fmt.Sprintf("%s/response-%d.json", sampleDir, responseCode)),
	)
	srv := httptest.NewServer(handler)
	return srv
}

func mockHandler(responseFile string) func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		respBytes, err := os.ReadFile(responseFile)
		u, p, _ := req.BasicAuth()
		// fmt.Printf("U=%s p=%s", u, p)
		// no username and no password? go away
		if u == "" && p == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err != nil {
			panic(err.Error())
		}
		_, _ = w.Write(respBytes)
	}
}
