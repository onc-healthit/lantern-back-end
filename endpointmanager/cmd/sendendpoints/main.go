package main

import (
	"context"
	"sync"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	se "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/sendendpoints"
	"github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
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
	go se.GetEnptsAndSend(ctx, &wg, capQName, capInterval, store, &mq, &channelID, errs)

	wg.Add(1)
	netInterval := viper.GetInt("endptqry_query_interval")
	go se.GetEnptsAndSend(ctx, &wg, netQName, netInterval, store, &mq, &channelID, errs)

	for elem := range errs {
		log.Warn(elem)
	}

	wg.Wait()
}
