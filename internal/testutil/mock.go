package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
)

const (
	sampleDir   = "../../testdata"
	ClusterID   = "abc-r2d2"
	topicPrefix = "public.hello"
)

func ServerMock() *httptest.Server {
	handler := http.NewServeMux()
	// http.StatusUnauthorized /* 401 */ is html !!!
	for _, code := range []int{http.StatusOK, http.StatusBadRequest /*400*/, http.StatusForbidden /*403*/} {
		handler.HandleFunc(
			fmt.Sprintf("/kafka/v3/clusters/%s/topics/%s/records", ClusterID, Topic(code)),
			mockHandler(fmt.Sprintf("%s/response-%d.json", sampleDir, code)),
		)
	}
	srv := httptest.NewServer(handler)
	return srv
}

// Topic expects a response file testdata/response-<statuscode>
func Topic(statusCode int) string {
	return fmt.Sprintf("%s-%d", topicPrefix, statusCode)
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
