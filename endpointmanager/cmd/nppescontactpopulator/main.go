package main

import (
	"context"
	"math"
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/nppesquerier"
)

func main() {
	err := config.SetupConfig()
	helpers.FailOnError("", err)

	ctx := context.Background()
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	if len(os.Args) != 2 {
		log.Fatal("NPPES contact csv file not provided as argument.")
	}
	fname := os.Args[1]
	err = store.DeleteAllNPIContacts(ctx)
	helpers.FailOnError("", err)

	log.Info("Adding NPI FHIR URLs to database...")
	added, err := nppesquerier.ParseAndStoreNPIContactsFile(ctx, fname, store)
	log.Infof("Added %d NPI FHIR URLs to database\n", added)
	helpers.FailOnError("", err)

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
