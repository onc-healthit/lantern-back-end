package main

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/lanternmq/rabbitmq"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

// connectToServerAndQueue creates a connection to an exchange at the given location with the given credentials.
func connectToServerAndQueue(qUser, qPassword, qHost, qPort, qName string) (lanternmq.MessageQueue, lanternmq.ChannelID, error) {
	mq := &rabbitmq.MessageQueue{}
	err := mq.Connect(qUser, qPassword, qHost, qPort)
	if err != nil {
		return nil, nil, err
	}
	ch, err := mq.CreateChannel()
	if err != nil {
		return nil, nil, err
	}
	return connectToQueue(mq, ch, qName)
}

func connectToQueue(mq lanternmq.MessageQueue, ch lanternmq.ChannelID, qName string) (lanternmq.MessageQueue, lanternmq.ChannelID, error) {
	exists, err := mq.QueueExists(ch, qName)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return nil, nil, errors.Errorf("queue %s does not exist", qName)
	}

	return mq, ch, nil
}

// sendToQueue publishes a message to the given queue
func sendToQueue(
	ctx context.Context,
	message string,
	mq *lanternmq.MessageQueue,
	ch *lanternmq.ChannelID,
	queueName string) error {

	// don't send the message if the context is done
	select {
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "unable to send message to queue - context ended")
	default:
		// ok
	}

	err := (*mq).PublishToQueue(*ch, queueName, message)
	if err != nil {
		return err
	}

	return nil
}

// getEndptsAndSend gets the current list of endpoints from the database and sends each one to the given queue
// it continues to repeat this action every time the given interval period has passed
func getEnptsAndSend(
	ctx context.Context,
	qName string,
	qInterval int,
	store endpointmanager.FHIREndpointStore,
	mq *lanternmq.MessageQueue,
	channelID *lanternmq.ChannelID) {

	for {
		listOfEndpoints, err := store.GetAllFHIREndpoints(ctx)
		failOnError(err)

		for i, endpt := range *listOfEndpoints {
			if i%10 == 0 {
				log.Infof("Processed %d/%d messages", i, len(*listOfEndpoints))
			}

			msgBytes, err := json.Marshal(endpt)
			failOnError(err)
			msgStr := string(msgBytes)
			err = sendToQueue(ctx, msgStr, mq, channelID, qName)
			failOnError(err)
		}

		log.Infof("Waiting %d minutes", qInterval)
		time.Sleep(time.Duration(qInterval) * time.Minute)
	}
}

func main() {
	log.Info("Started the endpoint manager.")

	err := config.SetupConfig()
	failOnError(err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError(err)
	log.Info("Successfully connected to DB!")

	// Set up the queue for sending messages to capabilityquerier and networkstatsquerier
	capQName := viper.GetString("enptinfo_capquery_qname")
	mq, channelID, err := connectToServerAndQueue(viper.GetString("quser"), viper.GetString("qpassword"), viper.GetString("qhost"), viper.GetString("qport"), capQName)
	failOnError(err)
	log.Info("Successfully connected to capabilityquerier Queue!")

	netQName := viper.GetString("enptinfo_netstats_qname")
	mq, channelID, err = connectToQueue(mq, channelID, netQName)
	failOnError(err)
	log.Info("Successfully connected to networkstatsquerier Queue!")

	// Infinite query loop
	var wg sync.WaitGroup
	ctx := context.Background()
	wg.Add(1)
	capInterval := viper.GetInt("capquery_qryintvl")
	go getEnptsAndSend(ctx, capQName, capInterval, store, &mq, &channelID)

	wg.Add(1)
	netInterval := viper.GetInt("endptqry_query_interval")
	go getEnptsAndSend(ctx, netQName, netInterval, store, &mq, &channelID)

	wg.Wait()
}
