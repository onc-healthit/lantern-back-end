package querier

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusArgs is a struct of the prometheus collectors that are used to save the values like
// response time and response code from the URL request and the URLString used to make the request and
// access the collectors
type PrometheusArgs struct {
	URLString                         string
	ResponseTimeGaugeVec              *prometheus.GaugeVec
	TotalUptimeChecksCounterVec       *prometheus.CounterVec
	TotalFailedUptimeChecksCounterVec *prometheus.CounterVec
	HTTPCodesGaugeVec                 *prometheus.GaugeVec
}

// Need to define timeout or else it is infinite
var netClient = &http.Client{
	Timeout: time.Second * 35,
}

// GetResponseAndTiming gets the response and response time for an http request to the
// endpoint at a given urlString and records the metrics into the appropriate prometheus
// register under the label specified by urlString
// The args are expected to be a map of the string "promArgs" to the above PrometheusArgs struct. It is formatted
// this way in order for it to be able to be called by a worker (see endpointmanager/pkg/workers)
func GetResponseAndTiming(ctx context.Context, args *map[string]interface{}) error {
	// Get arguments
	promArgs, ok := (*args)["promArgs"].(PrometheusArgs)
	if !ok {
		return fmt.Errorf("unable to case promArgs to type PrometheusArgs from arguments")
	}

	// recover from fatal errors
	if err := recover(); err != nil {
		return err.(error)
	}

	// Specifically query the FHIR endpoint metadata
	metadataURL, err := url.Parse(promArgs.URLString)
	if err != nil {
		return fmt.Errorf("Endpoint URL Parsing Error: %s", err.Error())
	}
	normalizedURL := endpointmanager.NormalizeEndpointURL(metadataURL.String())
	// Add a short time buffer before sending HTTP request to reduce burden on servers hosting multiple endpoints
	time.Sleep(time.Duration(500 * time.Millisecond))
	req, err := http.NewRequest("GET", normalizedURL, nil)
	if err != nil {
		return err
	}

	var start time.Time
	trace := &httptrace.ClientTrace{}

	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))

	start = time.Now()
	resp, err := netClient.Do(req)

	if err != nil {
		return err
	}

	var responseTime = float64(time.Since(start).Seconds())

	promArgs.ResponseTimeGaugeVec.WithLabelValues(promArgs.URLString).Set(responseTime)

	if resp != nil && resp.StatusCode != http.StatusOK {
		promArgs.TotalFailedUptimeChecksCounterVec.WithLabelValues(promArgs.URLString).Inc()
	}
	if resp != nil {
		promArgs.HTTPCodesGaugeVec.WithLabelValues(promArgs.URLString).Set(float64(resp.StatusCode))
	}
	promArgs.TotalUptimeChecksCounterVec.WithLabelValues(promArgs.URLString).Inc()

	return err
}
