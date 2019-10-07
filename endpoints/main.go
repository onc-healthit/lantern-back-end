package main

import (
	"github.com/onc-healthit/lantern-back-end/endpoints/querier"
	"github.com/onc-healthit/lantern-back-end/endpoints/fetcher"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/onc-healthit/lantern-back-end/fhir"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics collected inside built-in prometheus vector.
// Each METRIC has its own registrations
var httpCodesGaugeVec *prometheus.GaugeVec
var responseTimeGaugeVec *prometheus.GaugeVec
var tlsVersionGaugeVec *prometheus.GaugeVec
var fhirVersionGaugeVec *prometheus.GaugeVec
var totalUptimeChecksCounterVec *prometheus.CounterVec
var totalFailedUptimeChecksCounterVec *prometheus.CounterVec

func getHTTPRequestTiming(urlString string, organizationName string, recordLongRunningMetrics bool) {
	var resp, responeTime, err = querier.GetResponseAndTiming(urlString)

	responseTimeGaugeVec.WithLabelValues(organizationName).Set(responeTime)
	// Need to think about whether or not an errored request is considered a failed uptime check
	if err != nil || (resp != nil && resp.StatusCode != http.StatusOK) {
		totalFailedUptimeChecksCounterVec.WithLabelValues(organizationName).Inc()
	}
	if resp != nil {
		if resp.StatusCode == http.StatusOK && recordLongRunningMetrics {
			recordLongRunningStats(resp, organizationName)
		}
		httpCodesGaugeVec.WithLabelValues(organizationName).Set(float64(resp.StatusCode))
	}
	totalUptimeChecksCounterVec.WithLabelValues(organizationName).Inc()
}

func recordLongRunningStats(resp *http.Response, organizationName string) {
	var capabilityStatement = fhir.ParseConformanceStatement(resp)
	var fhirVersionString string = capabilityStatement.FhirVersion.Value
	var fhirVersionAsNumber, _ = strconv.Atoi(strings.Replace(fhirVersionString, ".", "", -1))
	fhirVersionGaugeVec.WithLabelValues(organizationName).Set(float64(fhirVersionAsNumber))
	tlsVersionGaugeVec.WithLabelValues(organizationName).Set(float64(resp.TLS.Version))
}

func initializeMetrics() {
	httpCodesGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "AllEndpoints",
			Name:      "http_request_responses",
			Help:      "HTTP request responses partitioned by orgName",
		},
		[]string{"orgName"})

	responseTimeGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "AllEndpoints",
			Name:      "http_response_time",
			Help:      "HTTP response time partitioned by orgName",
		},
		[]string{"orgName"})

	tlsVersionGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "AllEndpoints",
			Name:      "tls_version",
			Help:      "TLS version reported in the response header partitioned by orgName",
		},
		[]string{"orgName"})

	fhirVersionGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "AllEndpoints",
			Name:      "fhir_version",
			Help:      "FHIR version reported in the Capability statement partitioned by orgName",
		},
		[]string{"orgName"})

	totalUptimeChecksCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "AllEndpoints",
			Name:      "total_uptime_checks",
			Help:      "Total number of uptime checks partitioned by orgName",
		},
		[]string{"orgName"})

	totalFailedUptimeChecksCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "AllEndpoints",
			Name:      "total_failed_uptime_checks",
			Help:      "Total number of failed uptime checks partitioned by orgName",
		},
		[]string{"orgName"})

	prometheus.MustRegister(httpCodesGaugeVec)
	prometheus.MustRegister(responseTimeGaugeVec)
	prometheus.MustRegister(tlsVersionGaugeVec)
	prometheus.MustRegister(fhirVersionGaugeVec)
	prometheus.MustRegister(totalUptimeChecksCounterVec)
	prometheus.MustRegister(totalFailedUptimeChecksCounterVec)

}

func setupServer() {
	// Setup hosted metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	// TODO: Configure port in configureation file
	var err = http.ListenAndServe(":8443", nil)
	if err != nil {
		// TODO: Use a logging solution instead of println
		println("HTTP Request Error: ", err.Error())
	}
}

func main() {

	go setupServer()

	var endpointsFile string
	if len(os.Args) != 1 {
		endpointsFile = os.Args[1]
	} else {
		println("ERROR: Missing endpoints list command-line arguement")
		return
	}
	// Data in resources/EndpointSources was taken from https://fhirfetcher.github.io/data.json
	var listOfEndpoints = fetcher.GetListOfEndpoints(endpointsFile)
	initializeMetrics()

	var queryCount = 0
	// Infinite query loop
	for {
		for _, endpointEntry := range listOfEndpoints.Entries {
			// TODO: Distribute calls using a worker of some sort so that we are not sending out a million requests at once
			var url = endpointEntry.FHIRPatientFacingURI
			var orgName = endpointEntry.OrganizationName
			// Long polling stats will be queried for every 6 hours
			var longPollingInterval = (queryCount%72 == 0)
			getHTTPRequestTiming(url, orgName, longPollingInterval)
		}
		runtime.GC()
		// Polling interval, only necessary when running http calls asynchronously
		// time.Sleep(5 * time.Minute)
		queryCount += 1
	}

}
