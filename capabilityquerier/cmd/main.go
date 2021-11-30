package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/capabilityquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/historypruning"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/jsonexport"
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
	store       *postgresql.Store
}

// queryEndpointsCapabilityStatement gets an endpoint from the queue message and queries it to get the Capability Statement.
// This function is expected to be called by the lanternmq ProcessMessages function.
// parameter message:  the queue message that is being processed by this function, which is just an endpoint.
// parameter args:     expected to be a map of the string "queryArgs" to the above queryArgs struct. It is formatted
// 					   this way because queue processing is generalized.
func queryEndpointsCapabilityStatement(message []byte, args *map[string]interface{}) error {
	// Get arguments
	qa, ok := (*args)["queryArgs"].(queryArgs)
	if !ok {
		return fmt.Errorf("unable to cast queryArgs from arguments")
	}

	var msgJSON map[string]string
	err := json.Unmarshal(message, &msgJSON)
	if err != nil {
		return fmt.Errorf("Error parsing queryEndpointsCapabilityStatement message JSON: %s", err.Error())
	}

	urlString := msgJSON["url"]
	requestVersion := msgJSON["requestVersion"]
	defaultVersion := msgJSON["defaultVersion"]
	exportFileWait := viper.GetInt("exportfile_wait")

	if urlString == "FINISHED" {
		historypruning.PruneInfoHistory(qa.ctx, qa.store, true)
		time.Sleep(time.Duration(exportFileWait) * time.Second)
		err := jsonexport.CreateJSONExport(qa.ctx, qa.store, "/etc/lantern/exportfolder/fhir_endpoints_fields.json")
		return err
	}

	jobArgs := make(map[string]interface{})

	jobArgs["querierArgs"] = capabilityquerier.QuerierArgs{
		FhirURL:        urlString,
		RequestVersion: requestVersion,
		DefaultVersion: defaultVersion,
		Client:         qa.client,
		MessageQueue:   qa.mq,
		ChannelID:      qa.ch,
		QueueName:      qa.qName,
		UserAgent:      qa.userAgent,
		Store:          qa.store,
	}

	job := workers.Job{
		Context:     qa.ctx,
		Duration:    qa.jobDuration,
		Handler:     (capabilityquerier.GetAndSendCapabilityStatement),
		HandlerArgs: &jobArgs,
	}

	err = qa.workers.Add(&job)
	if err != nil {
		return fmt.Errorf("error adding job to workers: %s", err.Error())
	}

	return nil
}

// queryEndpointsVersionsOperation gets an endpoint from the queue message and queries it to get supported versions
// This function is expected to be called by the lanternmq ProcessMessages function.
// parameter message:  the queue message that is being processed by this function, which is just an endpoint.
// parameter args:     expected to be a map of the string "queryArgs" to the above queryArgs struct. It is formatted
// 					   this way because queue processing is generalized.
func queryEndpointsVersionsOperation(message []byte, args *map[string]interface{}) error {
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
		Store:        qa.store,
	}

	job := workers.Job{
		Context:     qa.ctx,
		Duration:    qa.jobDuration,
		Handler:     (capabilityquerier.GetAndSendVersionsResponse),
		HandlerArgs: &jobArgs,
	}

	err := qa.workers.Add(&job)
	if err != nil {
		return fmt.Errorf("error adding job to workers: %s", err.Error())
	}

	return nil
}

func setupQueue(store *postgresql.Store, userAgent string, client *http.Client, ctx context.Context, qName string, endptQName string, processFunc lanternmq.MessageHandler) {
	// Set up the queue for sending messages
	qUser := viper.GetString("quser")
	qPassword := viper.GetString("qpassword")
	qHost := viper.GetString("qhost")
	qPort := viper.GetString("qport")
	mq, ch, err := aq.ConnectToServerAndQueue(qUser, qPassword, qHost, qPort, qName)
	helpers.FailOnError("", err)

	mq, ch, err = aq.ConnectToQueue(mq, ch, endptQName)
	helpers.FailOnError("", err)

	defer mq.Close()

	errs := make(chan error)

	numWorkers := viper.GetInt("query_numworkers")
	workers := workers.NewWorkers()

	// Start workers and have them always running
	err = workers.Start(ctx, numWorkers, errs)
	helpers.FailOnError("", err)

	args := make(map[string]interface{})
	args["queryArgs"] = queryArgs{
		workers:     workers,
		ctx:         ctx,
		client:      client,
		jobDuration: 30 * time.Second,
		mq:          &mq,
		ch:          &ch,
		qName:       qName,
		userAgent:   userAgent,
		store:       store,
	}

	messages, err := mq.ConsumeFromQueue(ch, endptQName)
	helpers.FailOnError("", err)

	go mq.ProcessMessages(ctx, messages, processFunc, &args, errs)

	for elem := range errs {
		log.Warn(elem)
	}
}

func main() {
	err := config.SetupConfig()
	helpers.FailOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	log.Info("Successfully connected to DB!")

	var emptyJSON []byte
	if _, err := os.Stat("/etc/lantern/exportfolder/fhir_endpoints_fields.json"); os.IsNotExist(err) {
		err = ioutil.WriteFile("/etc/lantern/exportfolder/fhir_endpoints_fields.json", emptyJSON, 0644)
		helpers.FailOnError("Failed to create empty JSON export file", err)
	}

	// Read version file that is mounted
	version, err := ioutil.ReadFile("/etc/lantern/VERSION")
	helpers.FailOnError("", err)
	versionString := string(version)
	versionNum := strings.Split(versionString, "=")
	userAgent := "LANTERN/" + versionNum[1]
	userAgent = strings.TrimSuffix(userAgent, "\n")

	client := &http.Client{
		Timeout: time.Second * 35,
	}

	ctx := context.Background()

	versionResponseQName := viper.GetString("versionsquery_response_qname")
	versionEndptQName := viper.GetString("versionsquery_qname")
	go setupQueue(store, userAgent, client, ctx, versionResponseQName, versionEndptQName, queryEndpointsVersionsOperation)
	capQName := viper.GetString("capquery_qname")
	capQueryEndptQName := viper.GetString("endptinfo_capquery_qname")
	setupQueue(store, userAgent, client, ctx, capQName, capQueryEndptQName, queryEndpointsCapabilityStatement)

}
