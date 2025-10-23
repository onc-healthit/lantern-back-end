package main

import (
	"context"
	"fmt"
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
	log.Info("Running data validation check")

	queryInterval := viper.GetInt("capquery_qryintvl")

	ctx := context.Background()

	// Remove all endpoints from fhir_endpoints_info table that are not in the fhir_endpoints table
	err = store.DeleteFHIREndpointInfoOldEntries(ctx)
	helpers.FailOnError("", err)

	// Divide query interval (in seconds) by an average of 1.5 seconds per request to get the maximum number of endpoints that can be queried within query interval
	maxEndpoints := int(math.Floor(float64(queryInterval*60) / float64(1.5)))

	var endpointTotal int
	endpointCountQuery := "SELECT COUNT(*) from fhir_endpoints;"
	err = store.DB.QueryRow(endpointCountQuery).Scan(&endpointTotal)
	helpers.FailOnError("", err)

	if endpointTotal >= maxEndpoints {
		querierScale := int(math.Ceil(float64(endpointTotal) / float64(maxEndpoints)))
		queryIntervalIncrease := int(math.Ceil(float64(float64(endpointTotal)*float64(1.5)) / float64(60)))
		log.Warn(fmt.Sprintf("The current number of endpoints (%d) exceeds the maximum amount of endpoints that can be queried within the given Lantern query interval (%d minutes). Make sure to either scale out the capability querier service as defined in the README, or define a longer query threshold. With current query interval make sure you have at least %d querier instances, or with one querier instance make sure you increase query interval to at least %d minutes", endpointTotal, queryInterval, querierScale, queryIntervalIncrease))
	}
}
