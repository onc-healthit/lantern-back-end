package main

import (
	"context"
	"os"
	"strconv"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/jsonexport"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	var exportFile string
	monthlyExport := false
	var err error

	if len(os.Args) == 2 {
		exportFile = os.Args[1]
	} else if len(os.Args) > 2 {
		exportFile = os.Args[1]
		monthlyExport, err = strconv.ParseBool(os.Args[2])
		helpers.FailOnError("", err)
	} else {
		log.Fatalf("ERROR: Missing export file name command-line argument")
	}

	err = config.SetupConfig()
	helpers.FailOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	ctx := context.Background()
	log.Info("Successfully connected to DB!")

	err = jsonexport.CreateJSONExport(ctx, store, exportFile, monthlyExport)
	helpers.FailOnError("", err)
}
