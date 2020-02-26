package main

import (
	"context"
	"os"
	"log"
	"github.com/spf13/viper"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointlinker"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
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

	err := config.SetupConfig()
	failOnError("Error setting up config", err)
	ctx := context.Background()

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError("Error creating store", err)

	err = endpointlinker.LinkAllOrgsAndEndpoints(ctx, store, verbose);
	failOnError("Error linking all orgs and enpoints", err)


}