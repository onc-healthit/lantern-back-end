package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/capabilityquerier"
	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/workers"
	aq "github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func queryEndpoints(message []byte, args *map[string]interface{}) error {
	// Get arguments
	wkrs, ok := (*args)["workers"].(*workers.Workers)
	if !ok {
		return fmt.Errorf("unable to cast capabilityquerier QueueWorkers from arguments")
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
	fhirURL, err := url.Parse(urlString)
	if err != nil {
		log.Warnf("Error parsing URL string %s\n", urlString)
		return fmt.Errorf("endpoint URL parsing error: %s", err.Error())
	}

	jobArgs := make(map[string]interface{})
	jobArgs["FHIRURL"] = fhirURL
	jobArgs["client"] = (*args)["client"]
	jobArgs["mq"] = (*args)["mq"]
	jobArgs["ch"] = (*args)["ch"]
	jobArgs["qName"] = (*args)["qName"]

	job := workers.Job{
		Context:     ctx,
		Duration:    jobDuration,
		Handler:     (capabilityquerier.GetAndSendCapabilityStatement),
		HandlerArgs: &jobArgs,
	}

	err = wkrs.Add(&job)
	if err != nil {
		return fmt.Errorf("error adding job to workers: %s", err.Error())
	}

	return nil
}

func main() {
	err := config.SetupConfig()
	failOnError(err)

	// Set up the queue for sending messages
	qUser := viper.GetString("quser")
	qPassword := viper.GetString("qpassword")
	qHost := viper.GetString("qhost")
	qPort := viper.GetString("qport")
	capQName := viper.GetString("capquery_qname")
	mq, ch, err := aq.ConnectToServerAndQueue(qUser, qPassword, qHost, qPort, capQName)
	failOnError(err)

	endptQName := viper.GetString("endptinfo_capquery_qname")
	mq, ch, err = aq.ConnectToQueue(mq, ch, endptQName)
	failOnError(err)

	defer mq.Close()

	client := &http.Client{
		Timeout: time.Second * 35,
	}

	errs := make(chan error)

	numWorkers := viper.GetInt("capquery_numworkers")
	workers := workers.NewWorkers()
	ctx := context.Background()

	args := make(map[string]interface{})
	args["workers"] = workers
	args["ctx"] = ctx
	args["numWorkers"] = numWorkers
	args["errs"] = errs
	args["client"] = client
	args["jobDuration"] = 30 * time.Second
	args["mq"] = &mq
	args["ch"] = &ch
	args["qName"] = capQName

	messages, err := mq.ConsumeFromQueue(ch, endptQName)
	failOnError(err)

	go mq.ProcessMessages(ctx, messages, queryEndpoints, &args, errs)

	for elem := range errs {
		log.Warn(elem)
	}
}
