package main

import (
	"context"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"

	"github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/capabilityreceiver/pkg/capabilityhandler"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

func setupCapStatReception(ctx context.Context, store *postgresql.Store ){
	// Set up the queue for sending messages
	qName := viper.GetString("capquery_qname")
	messageQueue, channelID, err := accessqueue.ConnectToServerAndQueue(viper.GetString("quser"), viper.GetString("qpassword"), viper.GetString("qhost"), viper.GetString("qport"), qName)
	helpers.FailOnError("", err)
	log.Info("Successfully connected to Capability Statemsnts Queue!")
	defer messageQueue.Close()

	err = capabilityhandler.ReceiveCapabilityStatements(ctx, store, messageQueue, channelID, qName)
	helpers.FailOnError("", err)
}

func setupVersionsReception(ctx context.Context, store *postgresql.Store ){
	// Set up the queue for sending messages
	qName := viper.GetString("versionsquery_response_qname")
	messageQueue, channelID, err := accessqueue.ConnectToServerAndQueue(viper.GetString("quser"), viper.GetString("qpassword"), viper.GetString("qhost"), viper.GetString("qport"), qName)
	helpers.FailOnError("", err)
	log.Info("Successfully connected to Versions Response Queue!")
	defer messageQueue.Close()

	capQueryQueue, capQueryChannelID, err := accessqueue.ConnectToServerAndQueue(viper.GetString("quser"), viper.GetString("qpassword"), viper.GetString("qhost"), viper.GetString("qport"), "endpoints-to-capability")
	helpers.FailOnError("", err)
	log.Info("Successfully connected to capabilityquerier Queue!")
	defer capQueryQueue.Close()

	err = capabilityhandler.ReceiveVersionResponses(ctx, store, messageQueue, channelID, qName, capQueryQueue, capQueryChannelID)
	helpers.FailOnError("", err)
}

func main() {
	err := config.SetupConfig()
	helpers.FailOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	log.Info("Successfully connected to DB!")

	ctx := context.Background()

	go setupVersionsReception(ctx, store)
	setupCapStatReception(ctx, store)

}
