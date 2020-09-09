package main

import (
	"context"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	log.Info("Starting to export endpoints")

	err := config.SetupConfig()
	helpers.FailOnError("Error setting up config", err)
	ctx := context.Background()

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("Error creating store", err)
	// Copy entire contents of endpoint_export view into a csv which will be written to /tmp
	sql_query := "COPY (SELECT * FROM endpoint_export) TO '/tmp/export.csv' DELIMITER ',' CSV HEADER;"
	_, err = store.DB.ExecContext(ctx, sql_query)
	helpers.FailOnError("Error exporting csv", err)

}
