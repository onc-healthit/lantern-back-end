package main

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"time"

	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/capabilityquerier"
	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/config"
	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/endpoints"
	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/queue"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func queryEndpoints(ctx context.Context,
	listOfEndpoints *fetcher.ListOfEndpoints,
	qw *capabilityquerier.QueueWorkers,
	numWorkers int,
	jobDuration time.Duration,
	mq *lanternmq.MessageQueue,
	ch *lanternmq.ChannelID,
	qName string,
	client *http.Client,
) {
	err := qw.Start(ctx, numWorkers)
	failOnError(err)

	for i, endpointEntry := range listOfEndpoints.Entries {
		if i%10 == 0 {
			log.Infof("Processed %d/%d messages", i, len(listOfEndpoints.Entries))
		}
		var urlString = endpointEntry.FHIRPatientFacingURI
		// Specifically query the FHIR endpoint metadata
		metadataURL, err := url.Parse(urlString)
		if err != nil {
			log.Warn("endpoint URL parsing error: ", err.Error())
		} else {
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
				log.Warn("error adding job to queue workers: ", err.Error())
				break
			}
		}
	}

	log.Info("Stopping queue workers")
	err = qw.Stop()
	failOnError(err)
	log.Info("Done retrieving and sending capability statement information")
	runtime.GC()
}

func main() {
	err := config.SetupConfig()
	failOnError(err)

	queryInterval := viper.GetInt("capquery_qryintvl")

	// TODO: continuing to use the list of endpoints and 'fetcher'. however, eventually we'll
	// be taking messages off of a queue and this code will be removed.
	listOfEndpoints, err := endpoints.GetEndpoints("../networkstatsquerier/resources/EndpointSources.json")
	failOnError(err)

	// Set up the queue for sending messages
	qUser := viper.GetString("quser")
	qPassword := viper.GetString("qpassword")
	qHost := viper.GetString("qhost")
	qPort := viper.GetString("qport")
	qName := viper.GetString("capquery_qname")
	mq, ch, err := queue.ConnectToQueue(qUser, qPassword, qHost, qPort, qName)
	failOnError(err)
	defer mq.Close()

	client := &http.Client{
		Timeout: time.Second * 35,
	}

	numWorkers := viper.GetInt("capquery_numworkers")
	qw := capabilityquerier.NewQueueWorkers()

	// Infinite query loop
	for {
		ctx := context.Background()

		queryEndpoints(ctx, listOfEndpoints, qw, numWorkers, 30*time.Second, &mq, &ch, qName, client)

		log.Infof("Waiting %d minutes", queryInterval)
		time.Sleep(time.Duration(queryInterval) * time.Minute)
	}
}
