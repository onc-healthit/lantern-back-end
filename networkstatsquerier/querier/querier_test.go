package querier

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics collected inside built-in prometheus vector.
// Each METRIC has its own registrations
var httpCodesGaugeVec *prometheus.GaugeVec
var responseTimeGaugeVec *prometheus.GaugeVec
var totalUptimeChecksCounterVec *prometheus.CounterVec
var totalFailedUptimeChecksCounterVec *prometheus.CounterVec
var reg *prometheus.Registry

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
	setupPrometheus(t)
	server := MetadataResponseStub(t)
	defer server.Close()
	var capabilityStatmentURL = server.URL
	ctx := context.Background()
	// Drop connection if no reply within 30 seconds
	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancelFunc()

	args := make(map[string]interface{})
	promArgs := PrometheusArgs{
		URLString:                         capabilityStatmentURL,
		ResponseTimeGaugeVec:              responseTimeGaugeVec,
		TotalUptimeChecksCounterVec:       totalUptimeChecksCounterVec,
		TotalFailedUptimeChecksCounterVec: totalFailedUptimeChecksCounterVec,
		HTTPCodesGaugeVec:                 httpCodesGaugeVec,
	}
	args["promArgs"] = promArgs
	var err = GetResponseAndTiming(ctx, &args)

	if err != nil {
		t.Errorf("GetResponseAndTiming should not return an error, recieved error %s", err.Error())
	}
	teardownPrometheus(t)
}

// Create a test prometheus Registry and add collectors
func setupPrometheus(t *testing.T) {
	reg = prometheus.NewRegistry()

	httpCodesGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "AllEndpoints",
			Name:      "http_request_responses",
			Help:      "HTTP request responses partitioned by url",
		},
		[]string{"url"})

	responseTimeGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "AllEndpoints",
			Name:      "http_response_time",
			Help:      "HTTP response time partitioned by url",
		},
		[]string{"url"})

	totalUptimeChecksCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "AllEndpoints",
			Name:      "total_uptime_checks",
			Help:      "Total number of uptime checks partitioned by url",
		},
		[]string{"url"})

	totalFailedUptimeChecksCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "AllEndpoints",
			Name:      "total_failed_uptime_checks",
			Help:      "Total number of failed uptime checks partitioned by url",
		},
		[]string{"url"})

	err := reg.Register(httpCodesGaugeVec)
	th.Assert(t, err == nil, err)
	err = reg.Register(responseTimeGaugeVec)
	th.Assert(t, err == nil, err)
	err = reg.Register(totalUptimeChecksCounterVec)
	th.Assert(t, err == nil, err)
	err = reg.Register(totalFailedUptimeChecksCounterVec)
	th.Assert(t, err == nil, err)
}

// Unregister collectors
func teardownPrometheus(t *testing.T) {
	isUnregistered := reg.Unregister(httpCodesGaugeVec)
	th.Assert(t, isUnregistered == true, "prometheus could not unregister httpCodesGaugeVec")
	isUnregistered = reg.Unregister(responseTimeGaugeVec)
	th.Assert(t, isUnregistered == true, "prometheus could not unregister responseTimeGaugeVec")
	isUnregistered = reg.Unregister(totalUptimeChecksCounterVec)
	th.Assert(t, isUnregistered == true, "prometheus could not unregister totalUptimeChecksCounterVec")
	isUnregistered = reg.Unregister(totalFailedUptimeChecksCounterVec)
	th.Assert(t, isUnregistered == true, "prometheus could not unregister totalFailedUptimeChecksCounterVec")
}
