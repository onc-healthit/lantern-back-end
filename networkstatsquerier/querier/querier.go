package querier

import (
	"context"
	"strings"
	"net/http"
	"net/http/httptrace"
	"time"
)

// Need to define timeout or else it is infinite
var netClient = &http.Client{
	Timeout: time.Second * 35,
}

// Prepends url with https:// and appends with metadata/ if needed
func normalizeURL(url string) string{
	normalized := url
    // for cases such as foobar.com
    if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://")  {
        normalized = "https://" + normalized
    }

	// for cases such as foobar.com/
	if !strings.HasSuffix(url, "/metadata") && !strings.HasSuffix(url, "/metadata/") {
		if !strings.HasSuffix(url, "/") {
			normalized = normalized + "/"
		}
		normalized = normalized + "metadata"
	}
    return normalized
}

// GetResponseAndTiming returns the http response, the reponse time, the context cancel function and any errors for an http request to the endpoint at urlString
func GetResponseAndTiming(ctx context.Context, urlString string) (*http.Response, float64, error) {
	// recover from fatal errors
	if err := recover(); err != nil {
		return nil, -1, err.(error)
	}

	normalizedURL := normalizeURL(urlString)

	req, err := http.NewRequest("GET", normalizedURL, nil)
	if err != nil {
		return nil, -1, err
	}

	var start time.Time
	trace := &httptrace.ClientTrace{}

	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))

	start = time.Now()
	resp, err := netClient.Do(req)

	if err != nil {
		return nil, -1, err
	}

	var responseTime = float64(time.Since(start).Seconds())

	return resp, responseTime, err
}
