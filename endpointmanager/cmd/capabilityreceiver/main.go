package main

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityhandler"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

func failOnError(errString string, err error) {
	if err != nil {
		log.Fatalf("%s %s", errString, err)
	}
}

func main() {
	err := config.SetupConfig()
	failOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError("", err)
	log.Info("Successfully connected to DB!")

	capabilityhandler.CapabilityReceiver(store)
}
