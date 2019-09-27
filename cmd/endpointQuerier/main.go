package main

import (
	"context"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"../../internal/endpoints"
	"../../internal/fhir"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics collected inside built-in prometheus vector. Each METRIC has its own registrations
var httpCodesGaugeVec *prometheus.GaugeVec
var responseTimeGaugeVec *prometheus.GaugeVec
var tlsVersionGaugeVec *prometheus.GaugeVec
var fhirVersionGaugeVec *prometheus.GaugeVec
var totalUptimeChecksCounterVec *prometheus.CounterVec
var totalFailedUptimeChecksCounterVec *prometheus.CounterVec

// Need to define timeout or else it is infinite
var netClient = &http.Client{
	Timeout: time.Second * 35,
}

func getHTTPRequestTimingFor(urlString string, organizationName string, recordLongRunningMetrics bool) {
	// recover from fatal errors
	if err := recover(); err != nil {
		println(err)
	}
	// Specifically query the FHIR endpoint metadata
	u, err := url.Parse(urlString)
	if err != nil {
		println("URL Parsing Error: ", err.Error())
	}
	u.Path = path.Join(u.Path, "metadata")

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
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
	responseTimeGaugeVec.WithLabelValues(organizationName).Set(float64(time.Since(start).Seconds()))

	// Need to think about wether or not an errored request is considered a failed uptime check
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
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	var conformanceStatement fhir.Conformance
	var err = xml.Unmarshal(bodyBytes, &conformanceStatement)
	if err != nil {
		println("Conformance Statement Parsing Error: ", err.Error())
	}
	var fhirVersionString string = conformanceStatement.FhirVersion.Value
	var fhirVersionAsNumber, _ = strconv.Atoi(strings.Replace(fhirVersionString, ".", "", -1))
	fhirVersionGaugeVec.WithLabelValues(organizationName).Set(float64(fhirVersionAsNumber))
	tlsVersionGaugeVec.WithLabelValues(organizationName).Set(float64(resp.TLS.Version))
}

func initializeMetrics(listOfEndpoints endpoints.ListOfEndpoints) {
	httpCodesGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "AllEndpoints",
			Name:      "http_request_responses",
			Help:      "HTTP requests partitioned by OrgName",
		},
		[]string{"orgName"})

	responseTimeGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "AllEndpoints",
			Name:      "http_response_time",
			Help:      "HTTP response time partitioned by OrgName",
		},
		[]string{"orgName"})

	tlsVersionGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "AllEndpoints",
			Name:      "tls_version",
			Help:      "TLS version reported in the response header",
		},
		[]string{"orgName"})

	fhirVersionGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "AllEndpoints",
			Name:      "fhir_version",
			Help:      "FHIR version reported in the conformance statement",
		},
		[]string{"orgName"})

	totalUptimeChecksCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "AllEndpoints",
			Name:      "total_uptime_checks",
			Help:      "Total number of uptime checks partitioned by OrgName",
		},
		[]string{"orgName"})

	totalFailedUptimeChecksCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "AllEndpoints",
			Name:      "total_failed_uptime_checks",
			Help:      "Total number of failed uptime checks partitioned by OrgName",
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
	var err = http.ListenAndServe(":8443", nil)
	if err != nil {
		println("HTTP Request Error: ", err.Error())
	}
}

func main() {

	go setupServer()

	var endpointsFile = os.Args[1]
	// Data in resources/EndpointSources was taken from https://fhirendpoints.github.io/data.json
	var listOfEndpoints = endpoints.GetListOfEndpoints(endpointsFile)
	initializeMetrics(listOfEndpoints)

	var queryCount = 0
	// Infinite query loop
	for {
		for _, endpointEntry := range listOfEndpoints.Entries {
			// TODO: Distribute calls using a worker of some sort so that we are not sending out a million requests at once
			var url = endpointEntry.FHIRPatientFacingURI
			var orgName = endpointEntry.OrganizationName
			// Long polling stats will be queried for every 6 hours
			var longPollingInterval = (queryCount%72 == 0)
			getHTTPRequestTimingFor(url, orgName, longPollingInterval)
		}
		runtime.GC()
		// Polling interval, only necessary when running http calls asynchronously
		// time.Sleep(5 * time.Minute)
		queryCount += 1
	}

}
