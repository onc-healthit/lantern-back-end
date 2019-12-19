package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/capabilityquerier"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/lanternmq/rabbitmq"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func setupConfig() {
	var err error
	viper.SetEnvPrefix("lantern")
	viper.AutomaticEnv()

	err = viper.BindEnv("quser")
	failOnError(err)
	err = viper.BindEnv("qpassword")
	failOnError(err)
	err = viper.BindEnv("qhost")
	failOnError(err)
	err = viper.BindEnv("qport")
	failOnError(err)
	err = viper.BindEnv("qcapstatq")
	failOnError(err)
	err = viper.BindEnv("capquery_numworkers")
	failOnError(err)

	viper.SetDefault("quser", "capabilityquerier")
	viper.SetDefault("qpassword", "capabilityquerier")
	viper.SetDefault("qhost", "localhost")
	viper.SetDefault("qport", "5672")
	viper.SetDefault("qcapstatq", "capability-statements")
	viper.SetDefault("capquery_numworkers", 10)
}

func connectToQueue(qName string) (lanternmq.MessageQueue, lanternmq.ChannelID, error) {
	mq := &rabbitmq.MessageQueue{}
	err := mq.Connect(viper.GetString("quser"), viper.GetString("qpassword"), viper.GetString("qhost"), viper.GetString("qport"))
	if err != nil {
		return nil, nil, err
	}
	ch, err := mq.CreateChannel()
	if err != nil {
		return nil, nil, err
	}
	exists, err := mq.QueueExists(ch, qName)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return nil, nil, errors.Errorf("queue %s does not exist", qName)
	}

	return mq, ch, nil
}

func getEndpoints() (*fetcher.ListOfEndpoints, error) {
	var endpointsFile string
	if len(os.Args) != 1 {
		endpointsFile = os.Args[1]
	} else {
		endpointsFile = "../../endpointnetworkquerier/resources/EndpointSources.json"
		//return nil, errors.New("Missing endpoints list command-line argument")
	}
	var listOfEndpoints, err = fetcher.GetListOfEndpoints(endpointsFile)
	if err != nil {
		log.Fatal("Endpoint List Parsing Error: ", err.Error())
	}
	return &listOfEndpoints, nil
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

	for _, endpointEntry := range listOfEndpoints.Entries {
		print(".")
		var urlString = endpointEntry.FHIRPatientFacingURI
		// Specifically query the FHIR endpoint metadata
		metadataURL, err := url.Parse(urlString)
		if err != nil {
			log.Warn("endpoint URL parsing error: ", err.Error())
		} else {
			metadataURL.Path = path.Join(metadataURL.Path, "metadata")

			// err = capabilityquerier.GetAndSendCapabilityStatement(ctx, metadataURL, client, mq, ch, qName)
			// if err != nil {
			// 	log.Warn(err.Error())
			// }

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

	println("## Stopping")
	err = qw.Stop()
	failOnError(err)
	println("## Done with round")
	runtime.GC()
}

func main() {
	setupConfig()

	// TODO: continuing to use the list of endpoints and 'fetcher'. however, eventually we'll
	// be taking messages off of a queue and this code will be removed.
	listOfEndpoints, err := getEndpoints()
	failOnError(err)

	// Set up the queue for sending messages
	qName := viper.GetString("qcapstatq")
	mq, ch, err := connectToQueue(qName)
	failOnError(err)
	defer mq.Close()

	client := &http.Client{
		Timeout: time.Second * 35,
	}

	numWorkers := viper.GetInt("capquery_numworkers")
	qw := capabilityquerier.NewQueueWorkers()
	//jobDuration, err := time.ParseDuration("30s")
	failOnError(err)

	// Infinite query loop
	for {
		print("*")
		ctx := context.Background()

		queryEndpoints(ctx, listOfEndpoints, qw, numWorkers, 30*time.Second, &mq, &ch, qName, client)

		time.Sleep(5 * time.Minute)
	}
}
