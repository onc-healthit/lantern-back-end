package main

import (
	"context"
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fetcher"
	endptQuerier "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fhirendpointquerier"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/sharedfunctions"
)

func main() {
	var endpointsFile string
	var source string

	if len(os.Args) == 3 {
		endpointsFile = os.Args[1]
		source = os.Args[2]
	} else if len(os.Args) == 2 {
		log.Fatalf("ERROR: Missing endpoints list source command-line argument")
	} else {
		log.Fatalf("ERROR: Missing endpoints list command-line argument")
	}

	listOfEndpoints, err := fetcher.GetEndpointsFromFilepath(endpointsFile, source)
	sharedfunctions.FailOnError("Endpoint List Parsing Error: ", err)

	if len(listOfEndpoints.Entries) != 0 {
		err = config.SetupConfig()
		sharedfunctions.FailOnError("", err)

		ctx := context.Background()
		store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
		sharedfunctions.FailOnError("", err)
		log.Info("Successfully connected to DB!")

		dbErr := endptQuerier.AddEndpointData(ctx, store, &listOfEndpoints)
		sharedfunctions.FailOnError("Saving in fhir_endpoints database error: ", dbErr)
	}
}
