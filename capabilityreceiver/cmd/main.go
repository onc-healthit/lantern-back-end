package main

import (
	"context"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"

	"github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/capabilityreceiver/pkg/capabilityhandler"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/sharedfunctions"
)

func main() {
	err := config.SetupConfig()
	sharedfunctions.FailOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	sharedfunctions.FailOnError("", err)
	log.Info("Successfully connected to DB!")

	// Set up the queue for sending messages
	qName := viper.GetString("capquery_qname")
	messageQueue, channelID, err := accessqueue.ConnectToServerAndQueue(viper.GetString("quser"), viper.GetString("qpassword"), viper.GetString("qhost"), viper.GetString("qport"), qName)
	sharedfunctions.FailOnError("", err)
	log.Info("Successfully connected to Queue!")
	defer messageQueue.Close()

	ctx := context.Background()

	err = capabilityhandler.ReceiveCapabilityStatements(ctx, store, messageQueue, channelID, qName)
	sharedfunctions.FailOnError("", err)
}
