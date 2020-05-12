package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/config"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/querier"
	"github.com/spf13/viper"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

// Metrics collected inside built-in prometheus vector.
// Each METRIC has its own registrations
var httpCodesGaugeVec *prometheus.GaugeVec
var responseTimeGaugeVec *prometheus.GaugeVec
var totalUptimeChecksCounterVec *prometheus.CounterVec
var totalFailedUptimeChecksCounterVec *prometheus.CounterVec

// getHTTPRequestTiming records the http request characteristics for the endpoint specified by urlString
// Record the metrics into the appropriate prometheus register under the label specified by organizationName
func getHTTPRequestTiming(urlString string) {
	ctx := context.Background()
	// Closing context if HTTP request and response processing is not completed within 30 seconds.
	// This includes dropping the request connection if there's no reply within 30 seconds.
	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancelFunc()

	var resp, responseTime, err = querier.GetResponseAndTiming(ctx, urlString)

	if err != nil {
		log.WithFields(log.Fields{"url": urlString}).Warn("Error getting response charactaristics for endpoint.", err.Error())
	} else {
		responseTimeGaugeVec.WithLabelValues(urlString).Set(responseTime)

		if resp != nil && resp.StatusCode != http.StatusOK {
			totalFailedUptimeChecksCounterVec.WithLabelValues(urlString).Inc()
		}
		if resp != nil {
			httpCodesGaugeVec.WithLabelValues(urlString).Set(float64(resp.StatusCode))
		}
		totalUptimeChecksCounterVec.WithLabelValues(urlString).Inc()
	}
}

func initializeMetrics() {
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

	prometheus.MustRegister(httpCodesGaugeVec)
	prometheus.MustRegister(responseTimeGaugeVec)
	prometheus.MustRegister(totalUptimeChecksCounterVec)
	prometheus.MustRegister(totalFailedUptimeChecksCounterVec)

}

func setupServer() {
	// Setup hosted metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	var err = http.ListenAndServe(":"+viper.GetString("port"), nil)
	if err != nil {
		log.Fatal("HTTP Server Creation Error: ", err.Error())
	}
}

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func main() {
	config.SetupConfig()
	go setupServer()

	var endpointsFile string
	var source string
	if len(os.Args) == 3 {
		endpointsFile = os.Args[1]
		source = os.Args[2]
	} else if len(os.Args) == 2 {
		log.Error("missing endpoints list source command-line argument")
		return
	} else {
		log.Error("missing endpoints list command-line argument")
		return
	}
	// Data in resources/EndpointSources was taken from https://fhirfetcher.github.io/data.json
	var listOfEndpoints, err = fetcher.GetEndpointsFromFilepath(endpointsFile, source)
	if err != nil {
		log.Fatal("Endpoint List Parsing Error: ", err.Error())
	}
	initializeMetrics()

	var queryCount = 0
	// Infinite query loop
	for {
		for _, endpointEntry := range listOfEndpoints.Entries {
			// TODO: Distribute calls using a worker of some sort so that we are not sending out a million requests at once
			var urlString = endpointEntry.FHIRPatientFacingURI
			// Specifically query the FHIR endpoint metadata
			metadataURL, err := url.Parse(urlString)
			if err != nil {
				log.Warn("Endpoint URL Parsing Error: ", err.Error())
			} else {
				getHTTPRequestTiming(metadataURL.String())
			}
		}
		runtime.GC()

		// If the query interval is zero we will be continuously blasting out requests which causes broken connection issues
		// This is an issue in tests where we reduce the number of endpoint entries this introduces a minimum required pause time
		if viper.GetInt("query_interval") == 0 {
			time.Sleep(time.Duration(10 * time.Second))
		} else {
			time.Sleep(time.Duration(viper.GetInt("query_interval")) * time.Minute)
		}
		queryCount += 1
	}

}
