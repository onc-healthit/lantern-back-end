package main

import (
	"context"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"

	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/queue"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityhandler"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
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

	// Set up the queue for sending messages
	qName := viper.GetString("capquery_qname")
	messageQueue, channelID, err := queue.ConnectToQueue(viper.GetString("quser"), viper.GetString("qpassword"), viper.GetString("qhost"), viper.GetString("qport"), qName)
	failOnError(err)
	log.Info("Successfully connected to Queue!")
	defer messageQueue.Close()

	ctx := context.Background()

	err = capabilityhandler.ReceiveCapabilityStatements(ctx, store, messageQueue, channelID, qName)
	failOnError(err)
}
