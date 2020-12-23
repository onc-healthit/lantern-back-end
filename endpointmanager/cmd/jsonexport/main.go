package main

import (
	"context"
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/jsonexport"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	var exportFile string

	if len(os.Args) >= 1 {
		exportFile = os.Args[1]
	} else {
		log.Fatalf("ERROR: Missing export file name command-line argument")
	}

	err := config.SetupConfig()
	helpers.FailOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	ctx := context.Background()
	log.Info("Successfully connected to DB!")

	err = jsonexport.CreateJSONExport(ctx, store, exportFile)
	helpers.FailOnError("", err)
}
