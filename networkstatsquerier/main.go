package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
<<<<<<< 1216bac01245a9cc97266263dcdce796e3c1c7f3
	"os"
	"runtime"
=======
	"path"
>>>>>>> Networkstats querier now uses the queue to get the endpoint information.
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/workers"
	"github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
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
<<<<<<< e9c8d57d345fff6bc47875038460bdea130d3340
// Record the metrics into the appropriate prometheus register under the label specified by organizationName
func getHTTPRequestTiming(urlString string) {
	ctx := context.Background()
	// Closing context if HTTP request and response processing is not completed within 30 seconds.
	// This includes dropping the request connection if there's no reply within 30 seconds.
	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancelFunc()
=======
func getHTTPRequestTiming(message []byte, args *map[string]interface{}) error {
	// Get arguments
	wkrs, ok := (*args)["workers"].(*workers.Workers)
	if !ok {
		return fmt.Errorf("unable to cast Workers from arguments")
	}
	ctx, ok := (*args)["ctx"].(context.Context)
	if !ok {
		return fmt.Errorf("unable to cast context from arguments")
	}
	numWorkers, ok := (*args)["numWorkers"].(int)
	if !ok {
		return fmt.Errorf("unable to cast numWorkers to int from arguments")
	}
	errs, ok := (*args)["errs"].(chan error)
	if !ok {
		return fmt.Errorf("unable to cast errs to chan error from arguments")
	}
	jobDuration, ok := (*args)["jobDuration"].(time.Duration)
	if !ok {
		return fmt.Errorf("unable to cast jobDuration to time.Duration from arguments")
	}
>>>>>>> Networkstats querier now uses the queue to get the endpoint information.

	// Handle the start message that is sent before the endpoints and the stop message that is sent at the end
	if string(message) == "start" {
		err := wkrs.Start(ctx, numWorkers, errs)
		if err != nil {
			return err
		}
		return nil
	}
	if string(message) == "stop" {
		err := wkrs.Stop()
		if err != nil {
			return fmt.Errorf("error stopping queue workers: %s", err.Error())
		}
		return nil
	}

	urlString := string(message)
	// Specifically query the FHIR endpoint metadata
	metadataURL, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("Endpoint URL Parsing Error: %s", err.Error())
	}

	jobArgs := make(map[string]interface{})
	jobArgs["urlString"] = metadataURL.String()
	jobArgs["respTime"] = responseTimeGaugeVec
	jobArgs["failUptime"] = totalFailedUptimeChecksCounterVec
	jobArgs["httpCodes"] = httpCodesGaugeVec
	jobArgs["totalUptime"] = totalUptimeChecksCounterVec

	job := workers.Job{
		Context:     ctx,
		Duration:    jobDuration,
		Handler:     (querier.GetResponseAndTiming),
		HandlerArgs: &jobArgs,
	}

	err = wkrs.Add(&job)
	if err != nil {
		return fmt.Errorf("error adding job to workers: %s", err.Error())
	}

	return nil
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
	err := config.SetupConfig()
	failOnError(err)

	go setupServer()

	initializeMetrics()

<<<<<<< e9c8d57d345fff6bc47875038460bdea130d3340
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
=======
	// Set up the queue for receiving messages
	qUser := viper.GetString("quser")
	qPassword := viper.GetString("qpassword")
	qHost := viper.GetString("qhost")
	qPort := viper.GetString("qport")
	qName := viper.GetString("endptinfo_netstats_qname")
	mq, ch, err := accessqueue.ConnectToServerAndQueue(qUser, qPassword, qHost, qPort, qName)
	failOnError(err)
	defer mq.Close()

	messages, err := mq.ConsumeFromQueue(ch, qName)
	failOnError(err)

	// @TODO Setup numworkers environment variable for networkstatsquerier
	numWorkers := 10
	workers := workers.NewWorkers()
	ctx := context.Background()
>>>>>>> Networkstats querier now uses the queue to get the endpoint information.

	args := make(map[string]interface{})
	errs := make(chan error)
	args["workers"] = workers
	args["ctx"] = ctx
	args["numWorkers"] = numWorkers
	args["errs"] = errs
	args["jobDuration"] = 30 * time.Second

	go mq.ProcessMessages(ctx, messages, getHTTPRequestTiming, &args, errs)

	for elem := range errs {
		log.Warn(elem)
	}
}
