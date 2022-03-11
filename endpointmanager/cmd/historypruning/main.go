package main

import (
	"context"
	"os"
	"strconv"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/historypruning"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"

	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

func main() {
	var err error
	pruningLimit := false

	if len(os.Args) > 1 {
		pruningLimit, err = strconv.ParseBool(os.Args[1])
		helpers.FailOnError("", err)
	}

	err = config.SetupConfig()
	helpers.FailOnError("", err)

	ctx := context.Background()
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)

	historypruning.PruneInfoHistory(ctx, store, pruningLimit)
}
