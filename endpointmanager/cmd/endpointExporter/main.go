package main

import (
	"context"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func failOnError(errString string, err error) {
	if err != nil {
		log.Fatalf("%s %s", errString, err)
	}
}

func main() {
	log.Info("Starting to export endpoints")

	err := config.SetupConfig()
	failOnError("Error setting up config", err)
	ctx := context.Background()

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError("Error creating store", err)

	sql_query := "COPY (SELECT * FROM endpoint_export) TO '/tmp/export.csv' DELIMITER ',' CSV HEADER;"
	_, err = store.DB.ExecContext(ctx, sql_query)
	failOnError("Error exporting csv", err)

}
