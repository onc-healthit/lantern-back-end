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
	"github.com/onc-healthit/lantern-back-end/endpointmanager/sharedfunctions"
)

func main() {
	err := config.SetupConfig()
	sharedfunctions.FailOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	sharedfunctions.FailOnError("", err)
	log.Info("Successfully connected to DB!")

	// Set up the queue for sending messages to capabilityquerier
	capQName := viper.GetString("enptinfo_capquery_qname")
	mq, channelID, err := accessqueue.ConnectToServerAndQueue(viper.GetString("quser"), viper.GetString("qpassword"), viper.GetString("qhost"), viper.GetString("qport"), capQName)
	sharedfunctions.FailOnError("", err)
	log.Info("Successfully connected to capabilityquerier Queue!")

	errs := make(chan error)

	// Infinite query loop
	var wg sync.WaitGroup
	ctx := context.Background()
	wg.Add(1)
	capInterval := viper.GetInt("capquery_qryintvl")
	go se.GetEnptsAndSend(ctx, &wg, capQName, capInterval, store, &mq, &channelID, errs)

	for elem := range errs {
		log.Warn(elem)
	}

	wg.Wait()
}
