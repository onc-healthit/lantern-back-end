package querier

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/prometheus/client_golang/prometheus"
)

// Need to define timeout or else it is infinite
var netClient = &http.Client{
	Timeout: time.Second * 35,
}

// GetResponseAndTiming gets the response and response time for an http request to the
// endpoint at urlString and records the metrics into the appropriate prometheus register
// under the label specified by urlString
func GetResponseAndTiming(ctx context.Context, args *map[string]interface{}) error {
	// Get arguments
	urlString, ok := (*args)["urlString"].(string)
	if !ok {
		return fmt.Errorf("unable to cast urlString to string from arguments")
	}
	responseTimeGaugeVec, ok := (*args)["respTime"].(*prometheus.GaugeVec)
	if !ok {
		return fmt.Errorf("unable to cast respTime to *prometheus.GaugeVec from arguments")
	}
	totalUptimeChecksCounterVec, ok := (*args)["totalUptime"].(*prometheus.CounterVec)
	if !ok {
		return fmt.Errorf("unable to cast totalUptime to *prometheus.CounterVec from arguments")
	}
	totalFailedUptimeChecksCounterVec, ok := (*args)["failUptime"].(*prometheus.CounterVec)
	if !ok {
		return fmt.Errorf("unable to cast failUptime to *prometheus.CounterVec from arguments")
	}
	httpCodesGaugeVec, ok := (*args)["httpCodes"].(*prometheus.GaugeVec)
	if !ok {
		return fmt.Errorf("unable to cast httpCodes to *prometheus.GaugeVec from arguments")
	}

	// recover from fatal errors
	if err := recover(); err != nil {
		return err.(error)
	}

	normalizedURL := endpointmanager.NormalizeEndpointURL(urlString)

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

	responseTimeGaugeVec.WithLabelValues(urlString).Set(responseTime)

	if resp != nil && resp.StatusCode != http.StatusOK {
		totalFailedUptimeChecksCounterVec.WithLabelValues(urlString).Inc()
	}
	if resp != nil {
		httpCodesGaugeVec.WithLabelValues(urlString).Set(float64(resp.StatusCode))
	}
	totalUptimeChecksCounterVec.WithLabelValues(urlString).Inc()

	return err
}
