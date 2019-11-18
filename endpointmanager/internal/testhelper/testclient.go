package testhelper

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
)

// TestClient is test wrapper around an http client, allowing http responses to be mocked.
type TestClient struct {
	teardown func()
	http.Client
}

// NewTestClient creates a new TestClient using an httptest TLS Server. Any http requests using
// this client will be handled by 'handler'.
func NewTestClient(handler http.Handler) *TestClient {
	httpcli, teardown := testingHTTPClient(handler)

	tc := TestClient{
		teardown,
		*httpcli,
	}

	return &tc
}

// NewTestClientWithResponse creates a new TestClient using an httptest TLS Server, and requests
// are responded to using the given response byte string.
func NewTestClientWithResponse(response []byte) *TestClient {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(response)
	})

	tc := NewTestClient(h)

	return tc
}

// NewTestClientWith404 creates a new TestClient using an httptest TLS Server, and requests
// are responded to with a 404 response code.
func NewTestClientWith404() *TestClient {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "sample 404 error", http.StatusNotFound)
	})

	tc := NewTestClient(h)

	return tc
}

// NewTestClientWith502 creates a new TestClient using an httptest TLS Server, and requests
// are responded to with a 502 response code.
func NewTestClientWith502() *TestClient {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "sample 502 error", http.StatusBadGateway)
	})

	tc := NewTestClient(h)

	return tc
}

// Close closes resources associated with the test client and should be called after every
// instantiation of the client.
func (tc *TestClient) Close() {
	tc.teardown()
}

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewTLSServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return cli, s.Close
}
