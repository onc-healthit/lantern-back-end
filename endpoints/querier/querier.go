package querier

import (
	"context"
	"net/http"
	"net/http/httptrace"
	"time"
)

// Need to define timeout or else it is infinite
var netClient = &http.Client{
	Timeout: time.Second * 35,
}

// GetResponseAndTiming returns the http response, the reponse time, the context cancel function and any errors for an http request to the endpoint at urlString
func GetResponseAndTiming(ctx context.Context, urlString string) (*http.Response, float64, error) {
	// recover from fatal errors
	if err := recover(); err != nil {
		return nil, -1, err.(error)
	}

	req, err := http.NewRequest("GET", urlString, nil)
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
