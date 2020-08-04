package main

import (
	"context"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointlinker"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

func failOnError(errString string, err error) {
	if err != nil {
		log.Fatalf("%s %s", errString, err)
	}
}

func main() {
	var verbose = false
	if len(os.Args) > 1 && os.Args[1] == "--verbose" {
		verbose = true
	}
	log.Info("Starting to link FHIR endpoints to npi organizations")

	err := config.SetupConfig()
	failOnError("Error setting up config", err)
	ctx := context.Background()

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError("Error creating store", err)

	err = endpointlinker.LinkAllOrgsAndEndpoints(ctx, store, "/go/src/app/resources/linkerMatchesWhitelist.json", "/go/src/app/resources/linkerMatchesBlacklist.json", verbose)
	failOnError("Error linking all orgs and enpoints", err)

}
