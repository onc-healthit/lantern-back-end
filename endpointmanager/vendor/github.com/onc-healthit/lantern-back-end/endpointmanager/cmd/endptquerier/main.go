package main

import (
	"context"
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	endptQuerier "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/fhirendpointquerier"
	"github.com/onc-healthit/lantern-back-end/endpoints/fetcher"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

func failOnError(errString string, err error) {
	if err != nil {
		log.Fatalf("%s %s", errString, err)
	}
}

func main() {
	var endpointsFile string
	if len(os.Args) != 1 {
		endpointsFile = os.Args[1]
	} else {
		log.Fatalf("ERROR: Missing endpoints list command-line argument")
	}

	// Data in resources/EndpointSources was taken from https://fhirfetcher.github.io/data.json
	var listOfEndpoints, err = fetcher.GetListOfEndpoints(endpointsFile)
	failOnError("Endpoint List Parsing Error: ", err)

	err = config.SetupConfig()
	failOnError("", err)

	ctx := context.Background()
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError("", err)
	log.Info("Successfully connected to DB!")

	dbErr := endptQuerier.AddEndpointData(ctx, store, &listOfEndpoints)
	failOnError("Saving in fhir_endpoints database error: ", dbErr)
}
