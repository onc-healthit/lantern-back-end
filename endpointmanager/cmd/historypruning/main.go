package main

import (
	"context"
	"strconv"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/historypruning"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"

	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

func main() {
	err := config.SetupConfig()
	helpers.FailOnError("", err)

	ctx := context.Background()
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)

	thresholdInt := viper.GetInt("pruning_threshold")
	threshold := strconv.Itoa(thresholdInt)
	queryInterval := ""

	historypruning.HistoryPruningCheck(ctx, store, threshold, queryInterval)
}
