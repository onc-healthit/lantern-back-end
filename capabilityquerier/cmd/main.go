package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/capabilityquerier"
	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/config"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
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
	qw, ok := (*args)["qw"].(*capabilityquerier.QueueWorkers)
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
	client, ok := (*args)["client"].(*http.Client)
	if !ok {
		return fmt.Errorf("unable to cast client to http.Client from arguments")
	}
	jobDuration, ok := (*args)["jobDuration"].(time.Duration)
	if !ok {
		return fmt.Errorf("unable to cast jobDuration to time.Duration from arguments")
	}
	mq, ok := (*args)["mq"].(*lanternmq.MessageQueue)
	if !ok {
		return fmt.Errorf("unable to cast mq to MessageQueue from arguments")
	}
	ch, ok := (*args)["ch"].(*lanternmq.ChannelID)
	if !ok {
		return fmt.Errorf("unable to cast ch to ChannelID from arguments")
	}
	qName, ok := (*args)["qName"].(string)
	if !ok {
		return fmt.Errorf("unable to cast qName to string from arguments")
	}

	// Handle the start message that is sent before the endpoints and the stop message that is sent at the end
	if string(message) == "start" {
		err := qw.Start(ctx, numWorkers, errs)
		failOnError(err)
		return nil
	}
	if string(message) == "stop" {
		err := qw.Stop()
		if err != nil {
			return fmt.Errorf("error stopping queue workers: %s", err.Error())
		}
		return nil
	}

	urlString := string(message)
	// Specifically query the FHIR endpoint metadata
	metadataURL, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("endpoint URL parsing error: %s", err.Error())
	}

	metadataURL.Path = path.Join(metadataURL.Path, "metadata")

	job := capabilityquerier.Job{
		Context:      ctx,
		Duration:     jobDuration,
		FHIRURL:      metadataURL,
		Client:       client,
		MessageQueue: mq,
		Channel:      ch,
		QueueName:    qName,
	}

	err = qw.Add(&job)
	if err != nil {
		return fmt.Errorf("error adding job to queue workers: %s", err.Error())
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
	qw := capabilityquerier.NewQueueWorkers()
	ctx := context.Background()

	args := make(map[string]interface{})
	args["qw"] = qw
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

	go mq.ProcessMessages(messages, queryEndpoints, &args, errs)

	for elem := range errs {
		log.Warn(elem)
	}
}
