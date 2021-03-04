package main

import (
	"math"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

func main() {
	err := config.SetupConfig()
	helpers.FailOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	log.Info("Successfully connected to DB!")

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
