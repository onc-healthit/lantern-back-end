package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpoints/fetcher"
	"github.com/onc-healthit/lantern-back-end/endpoints/querier"
	"github.com/onc-healthit/lantern-back-end/fhir"
	"github.com/spf13/viper"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

// Metrics collected inside built-in prometheus vector.
// Each METRIC has its own registrations
var httpCodesGaugeVec *prometheus.GaugeVec
var responseTimeGaugeVec *prometheus.GaugeVec
var tlsVersionGaugeVec *prometheus.GaugeVec
var fhirVersionGaugeVec *prometheus.GaugeVec
var totalUptimeChecksCounterVec *prometheus.CounterVec
var totalFailedUptimeChecksCounterVec *prometheus.CounterVec

// getHTTPRequestTiming records the http request characteristics for the endpoint specified by urlString
// Record the metrics into the appropriate prometheus register under the label specified by organizationName
// recordLongRunningMetrics specifies wether or not to record information contained in the capability statment
func getHTTPRequestTiming(urlString string, organizationName string, recordLongRunningMetrics bool) {
	ctx := context.Background()
	// Closing context if HTTP request and response processing is not completed within 30 seconds.
	// This includes dropping the request connection if there's no reply within 30 seconds.
	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancelFunc()

	var resp, responseTime, err = querier.GetResponseAndTiming(ctx, urlString)

	if err != nil {
		log.WithFields(log.Fields{"organization": organizationName, "url": urlString}).Warn("Error getting response charactaristics for endpoint.", err.Error())
	} else {
		responseTimeGaugeVec.WithLabelValues(organizationName).Set(responseTime)

		if resp != nil && resp.StatusCode != http.StatusOK {
			totalFailedUptimeChecksCounterVec.WithLabelValues(organizationName).Inc()
		}
		if resp != nil {
			if resp.StatusCode == http.StatusOK && recordLongRunningMetrics {
				recordLongRunningStats(resp, organizationName, urlString)
			}
			httpCodesGaugeVec.WithLabelValues(organizationName).Set(float64(resp.StatusCode))
		}
		totalUptimeChecksCounterVec.WithLabelValues(organizationName).Inc()
	}
}

// Records information gathered from the capability statment into prometheus
func recordLongRunningStats(resp *http.Response, organizationName string, urlString string) {
	var capabilityStatement, err = fhir.ParseCapabilityStatement(resp)
	if err != nil {
		log.WithFields(log.Fields{"organization": organizationName, "url": urlString}).Warn("Capability Statement Response Parsing Error: ", err.Error())
	} else {
		var fhirVersionString string = capabilityStatement.FhirVersion.Value
		var fhirVersionAsNumber, _ = strconv.Atoi(strings.Replace(fhirVersionString, ".", "", -1))
		fhirVersionGaugeVec.WithLabelValues(organizationName).Set(float64(fhirVersionAsNumber))
		tlsVersionGaugeVec.WithLabelValues(organizationName).Set(float64(resp.TLS.Version))
	}
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
	var err = http.ListenAndServe(":"+viper.GetString("port"), nil)
	if err != nil {
		log.Fatal("HTTP Server Creation Error: ", err.Error())
	}
}

func setupConfig() {
	var err error
	viper.SetEnvPrefix("lantern_endptqry")
	viper.AutomaticEnv()

	err = viper.BindEnv("port")
	failOnError(err)
	err = viper.BindEnv("logfile")
	failOnError(err)

	viper.SetDefault("port", 8443)
	viper.SetDefault("logfile", "endpointQuerierLog.json")
}

func initializeLogger() {
	log.SetFormatter(&log.JSONFormatter{})
	f, err := os.OpenFile(viper.GetString("logfile"), os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal("LogFile creation error: ", err.Error())
	}
	log.SetOutput(f)
}

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func main() {
	setupConfig()
	initializeLogger()
	go setupServer()

	var endpointsFile string
	if len(os.Args) != 1 {
		endpointsFile = os.Args[1]
	} else {
		println("ERROR: Missing endpoints list command-line arguement")
		return
	}
	// Data in resources/EndpointSources was taken from https://fhirfetcher.github.io/data.json
	var listOfEndpoints, err = fetcher.GetListOfEndpoints(endpointsFile)
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
			var orgName = endpointEntry.OrganizationName
			// Long polling stats will be queried for every 6 hours
			//TODO: Config file
			var longPollingInterval = (queryCount%72 == 0)
			// Specifically query the FHIR endpoint metadata
			metadataURL, err := url.Parse(urlString)
			if err != nil {
				log.Warn("Endpoint URL Parsing Error: ", err.Error())
			} else {
				metadataURL.Path = path.Join(metadataURL.Path, "metadata")
				getHTTPRequestTiming(metadataURL.String(), orgName, longPollingInterval)
			}
		}
		runtime.GC()
		// Polling interval, only necessary when running http calls asynchronously
		// TODO: Config file
		// time.Sleep(5 * time.Minute)
		queryCount += 1
	}

}
