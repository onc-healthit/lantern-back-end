package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/capabilityquerier"
	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/workers"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	aq "github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

// queryArgs is a struct to hold the values necessary to set up workers for processing information
// (see endpointmanager/pkg/workers) as well as the arguments for the capabilityquerier.QuerierArgs
// struct that is used when calling capabilityquerier.GetAndSendCapabilityStatement
type queryArgs struct {
	workers     *workers.Workers
	ctx         context.Context
	client      *http.Client
	jobDuration time.Duration
	mq          *lanternmq.MessageQueue
	ch          *lanternmq.ChannelID
	qName       string
	userAgent   string
}

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

// queryEndpoints gets an endpoint from the queue message and queries it to get the Capability Statement.
// This function is expected to be called by the lanternmq ProcessMessages function.
// parameter message:  the queue message that is being processed by this function, which is just an endpoint.
// parameter args:     expected to be a map of the string "queryArgs" to the above queryArgs struct. It is formatted
// 					   this way because queue processing is generalized.
func queryEndpoints(message []byte, args *map[string]interface{}) error {
	// Get arguments
	qa, ok := (*args)["queryArgs"].(queryArgs)
	if !ok {
		return fmt.Errorf("unable to cast queryArgs from arguments")
	}

	urlString := string(message)

	jobArgs := make(map[string]interface{})

	jobArgs["querierArgs"] = capabilityquerier.QuerierArgs{
		FhirURL:      urlString,
		Client:       qa.client,
		MessageQueue: qa.mq,
		ChannelID:    qa.ch,
		QueueName:    qa.qName,
		UserAgent:    qa.userAgent,
	}

	job := workers.Job{
		Context:     qa.ctx,
		Duration:    qa.jobDuration,
		Handler:     (capabilityquerier.GetAndSendCapabilityStatement),
		HandlerArgs: &jobArgs,
	}

	err := qa.workers.Add(&job)
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

	// Read version file that is mounted
	version, err := ioutil.ReadFile("/etc/lantern/VERSION")
	failOnError(err)
	versionString := string(version)
	versionNum := strings.Split(versionString, "=")
	userAgent := "LANTERN/" + versionNum[1]
	userAgent = strings.TrimSuffix(userAgent, "\n")

	client := &http.Client{
		Timeout: time.Second * 35,
	}

	errs := make(chan error)

	numWorkers := viper.GetInt("capquery_numworkers")
	workers := workers.NewWorkers()
	ctx := context.Background()

	// Start workers and have then always running
	err = workers.Start(ctx, numWorkers, errs)
	failOnError(err)

	args := make(map[string]interface{})
	args["queryArgs"] = queryArgs{
		workers:     workers,
		ctx:         ctx,
		client:      client,
		jobDuration: 30 * time.Second,
		mq:          &mq,
		ch:          &ch,
		qName:       capQName,
		userAgent:   userAgent,
	}

	messages, err := mq.ConsumeFromQueue(ch, endptQName)
	failOnError(err)

	go mq.ProcessMessages(ctx, messages, queryEndpoints, &args, errs)

	for elem := range errs {
		log.Warn(elem)
	}
}
