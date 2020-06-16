package main

import (
	"context"
	"fmt"
	"net/http"
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

// workerArgs is a struct to hold the values necessary to set up workers for processing information
// (see endpointmanager/pkg/workers)
type workerArgs struct {
	workers     *workers.Workers
	ctx         context.Context
	numWorkers  int
	errs        chan error
	jobDuration time.Duration
}

// Metrics collected inside built-in prometheus vector.
// Each METRIC has its own registrations
var httpCodesGaugeVec *prometheus.GaugeVec
var responseTimeGaugeVec *prometheus.GaugeVec
var totalUptimeChecksCounterVec *prometheus.CounterVec
var totalFailedUptimeChecksCounterVec *prometheus.CounterVec

// getHTTPRequestTiming records the http request characteristics an endpoint given in the message variable
// This function is expected to be called by the lanternmq ProcessMessages function.
// parameter message:  the queue message that is being processed by this function, which is just an endpoint.
// parameter args:     expected to be a map of the string "workerArgs" to the above workerArgs struct. It is formatted
// 					   this way because queue processing is generalized.
func getHTTPRequestTiming(message []byte, args *map[string]interface{}) error {
	// Get arguments
	wa, ok := (*args)["workerArgs"].(workerArgs)
	if !ok {
		return fmt.Errorf("unable to cast workerArgs from arguments")
	}

	// Handle the start message that is sent before the endpoints and the stop message that is sent at the end
	if string(message) == "start" {
		err := wa.workers.Start(wa.ctx, wa.numWorkers, wa.errs)
		if err != nil {
			return err
		}
		return nil
	}
	if string(message) == "stop" {
		err := wa.workers.Stop()
		if err != nil {
			return fmt.Errorf("error stopping queue workers: %s", err.Error())
		}
		return nil
	}

	urlString := string(message)

	jobArgs := make(map[string]interface{})
	jobArgs["promArgs"] = querier.PrometheusArgs{
		URLString:                         urlString,
		ResponseTimeGaugeVec:              responseTimeGaugeVec,
		TotalFailedUptimeChecksCounterVec: totalFailedUptimeChecksCounterVec,
		HTTPCodesGaugeVec:                 httpCodesGaugeVec,
		TotalUptimeChecksCounterVec:       totalUptimeChecksCounterVec,
	}

	job := workers.Job{
		Context:     wa.ctx,
		Duration:    wa.jobDuration,
		Handler:     (querier.GetResponseAndTiming),
		HandlerArgs: &jobArgs,
	}

	err := wa.workers.Add(&job)
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

	numWorkers := viper.GetInt("numworkers")
	workers := workers.NewWorkers()
	ctx := context.Background()
	errs := make(chan error)

	args := make(map[string]interface{})
	args["workerArgs"] = workerArgs{
		workers:     workers,
		ctx:         ctx,
		numWorkers:  numWorkers,
		errs:        errs,
		jobDuration: 30 * time.Second,
	}

	go mq.ProcessMessages(ctx, messages, getHTTPRequestTiming, &args, errs)

	for elem := range errs {
		log.Warn(elem)
	}
}
