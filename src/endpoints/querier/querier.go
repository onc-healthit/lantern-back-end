package querier

import (
	"context"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"path"
	"time"
)

// Need to define timeout or else it is infinite
var netClient = &http.Client{
	Timeout: time.Second * 35,
}

// Returns the http response, the reponse time and any errors for an http request to the endpoint at urlString
func GetResponseAndTiming(urlString string) (*http.Response, float64, error) {
	// recover from fatal errors
	if err := recover(); err != nil {
		// TODO: Use a logging solution instead of println
		println(err)
	}
	// Specifically query the FHIR endpoint metadata
	u, err := url.Parse(urlString)
	if err != nil {
		// TODO: Use a logging solution instead of println
		println("URL Parsing Error: ", err.Error())
	}
	u.Path = path.Join(u.Path, "metadata")

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		// TODO: Use a logging solution instead of println
		println("HTTP Request Error: ", err.Error())
	}

	var start time.Time
	trace := &httptrace.ClientTrace{}

	// Drop connection if no reply within 30 seconds
	ctx, cancel := context.WithDeadline(req.Context(), time.Now().Add(30*time.Second))
	// Cancel the context once we are done with this function so that context does not remain in memory (causing a leak)
	defer cancel()
	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))

	start = time.Now()
	resp, err := netClient.Do(req)

	if err != nil {
		// TODO: Use a logging solution instead of println
		println("HTTP Request Error: ", err.Error())
	}

	var responseTime = float64(time.Since(start).Seconds())

	return resp, responseTime, err
}
