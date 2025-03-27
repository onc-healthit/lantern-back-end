package main

import (
	"context"
	"os"
	"sync"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	se "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/sendendpoints"
	"github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	err := config.SetupConfig()
	helpers.FailOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	log.Info("Successfully connected to DB!")

	// Set up the queue for sending messages to capabilityquerier
	capQName := viper.GetString("versionsquery_qname")
	mq, channelID, err := accessqueue.ConnectToServerAndQueue(viper.GetString("quser"), viper.GetString("qpassword"), viper.GetString("qhost"), viper.GetString("qport"), capQName)
	helpers.FailOnError("", err)
	log.Info("Successfully connected to capabilityquerier Queue!")

	errs := make(chan error)

	var emptyJSON []byte
	if _, err := os.Stat("/etc/lantern/exportfolder/fhir_endpoints_fields.json"); os.IsNotExist(err) {
		err = os.WriteFile("/etc/lantern/exportfolder/fhir_endpoints_fields.json", emptyJSON, 0644)
		helpers.FailOnError("Failed to create empty JSON export file", err)
	}

	// Infinite query loop
	var wg sync.WaitGroup
	ctx := context.Background()
	wg.Add(1)
	capInterval := viper.GetInt("capquery_qryintvl")
	go se.GetEnptsAndSend(ctx, &wg, capQName, capInterval, store, &mq, &channelID, errs)

	wg.Add(1)
	go se.HistoryPruning(ctx, &wg, capInterval, store, errs)

	wg.Add(1)
	go se.ValidationPruning(ctx, &wg, store, errs)

	for elem := range errs {
		log.Warn(elem)
	}

	wg.Wait()
}
