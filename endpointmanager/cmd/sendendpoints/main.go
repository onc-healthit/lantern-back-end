package main

import (
	"context"
	"sync"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

// GetEnptsAndSend gets the current list of endpoints from the database and sends each one to the given queue
// it continues to repeat this action every time the given interval period has passed
func GetEnptsAndSend(
	ctx context.Context,
	wg *sync.WaitGroup,
	qName string,
	qInterval int,
	store *postgresql.Store,
	mq *lanternmq.MessageQueue,
	channelID *lanternmq.ChannelID,
	errs chan<- error) {

	defer wg.Done()

	for {
		listOfEndpoints, err := store.GetAllFHIREndpoints(ctx)
		if err != nil {
			errs <- err
		}

		err = accessqueue.SendToQueue(ctx, "start", mq, channelID, qName)
		if err != nil {
			errs <- err
		}

		for i, endpt := range listOfEndpoints {
			if i%10 == 0 {
				log.Infof("Processed %d/%d messages", i, len(listOfEndpoints))
			}

			err = accessqueue.SendToQueue(ctx, endpt.URL, mq, channelID, qName)
			if err != nil {
				errs <- err
			}
		}

		err = accessqueue.SendToQueue(ctx, "stop", mq, channelID, qName)
		if err != nil {
			errs <- err
		}

		log.Infof("Waiting %d minutes", qInterval)
		time.Sleep(time.Duration(qInterval) * time.Minute)
	}
}

func main() {
	err := config.SetupConfig()
	failOnError(err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError(err)
	log.Info("Successfully connected to DB!")

	// Set up the queue for sending messages to capabilityquerier and networkstatsquerier
	capQName := viper.GetString("enptinfo_capquery_qname")
	mq, channelID, err := accessqueue.ConnectToServerAndQueue(viper.GetString("quser"), viper.GetString("qpassword"), viper.GetString("qhost"), viper.GetString("qport"), capQName)
	failOnError(err)
	log.Info("Successfully connected to capabilityquerier Queue!")

	netQName := viper.GetString("enptinfo_netstats_qname")
	mq, channelID, err = accessqueue.ConnectToQueue(mq, channelID, netQName)
	failOnError(err)
	log.Info("Successfully connected to networkstatsquerier Queue!")

	errs := make(chan error)

	// Infinite query loop
	var wg sync.WaitGroup
	ctx := context.Background()
	wg.Add(1)
	capInterval := viper.GetInt("capquery_qryintvl")
	go GetEnptsAndSend(ctx, &wg, capQName, capInterval, store, &mq, &channelID, errs)

	wg.Add(1)
	netInterval := viper.GetInt("endptqry_query_interval")
	go GetEnptsAndSend(ctx, &wg, netQName, netInterval, store, &mq, &channelID, errs)

	for elem := range errs {
		log.Warn(elem)
	}

	wg.Wait()
}
