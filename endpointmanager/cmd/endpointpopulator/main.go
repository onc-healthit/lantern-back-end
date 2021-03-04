package main

import (
	"context"
	"math"
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fetcher"
	endptQuerier "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fhirendpointquerier"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

func main() {
	var endpointsFile string
	var source string
	var listURL string

	if len(os.Args) == 3 {
		endpointsFile = os.Args[1]
		source = os.Args[2]
	} else if len(os.Args) == 4 {
		endpointsFile = os.Args[1]
		source = os.Args[2]
		listURL = os.Args[3]
	} else if len(os.Args) == 2 {
		log.Fatalf("ERROR: Missing endpoints list source command-line argument")
	} else {
		log.Fatalf("ERROR: Endpoints list command-line arguments are not correct")
	}

	listOfEndpoints, err := fetcher.GetEndpointsFromFilepath(endpointsFile, source, listURL)
	helpers.FailOnError("Endpoint List Parsing Error: ", err)

	if len(listOfEndpoints.Entries) != 0 {
		err = config.SetupConfig()
		helpers.FailOnError("", err)

		ctx := context.Background()
		store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
		helpers.FailOnError("", err)
		log.Info("Successfully connected to DB!")

		dbErr := endptQuerier.AddEndpointData(ctx, store, &listOfEndpoints)
		helpers.FailOnError("Saving in fhir_endpoints database error: ", dbErr)

		queryInterval := viper.GetInt("capquery_qryintvl")
		maxEndpoints := int(math.Round(float64(queryInterval*60) / float64(1.5)))

		var endpointTotal int
		endpointCountQuery := "SELECT COUNT(*) from fhir_endpoints;"
		err = store.DB.QueryRow(endpointCountQuery).Scan(&endpointTotal)
		helpers.FailOnError("", err)

		if endpointTotal >= maxEndpoints {
			log.Warn("The current number of endpoints exceeds the maximum amount of endpoints that can be queried within the given Lantern query interval. Make sure to either scale out the capability querier service as defined in the README, or define a longer query threshold.")
		}
	}
}
