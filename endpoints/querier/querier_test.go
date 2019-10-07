package querier

import (
	"testing"
	"net/http/httptest"
	"net/http"
)

func MetadataResponseStub(t *testing.T) *httptest.Server {
	var resp string
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	  switch r.RequestURI {
		case "/metadata":
		  resp = "foo"
		default:
		  http.Error(w, "not found", http.StatusNotFound)
		  return
	  }
	  var _, err = w.Write([]byte(resp))
	  if err != nil {
		t.Errorf("Error in test web mock %s", err.Error())
	  }
	}))
}

func Test_GetHTTP200Response(t *testing.T) {
	server := MetadataResponseStub(t)
	defer server.Close()
	var resp, responseTime, err = GetResponseAndTiming(server.URL)

	if err != nil {
		t.Errorf("GetResponseAndTiming should not return an error, recieved error %s", err.Error())
	}

	if resp.StatusCode != http.StatusOK{
		t.Errorf("GetResponseAndTiming is expected to recieve a 200 OK got: %d", resp.StatusCode)
	}

	if responseTime <= 0 {
		t.Errorf("GetResponseAndTiming should return a response time")
	}

}
