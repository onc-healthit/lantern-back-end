package main

import (
	"context"
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/nppesquerier"
)

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func main() {
	config.SetupConfig()

	ctx := context.Background()
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError(err)
	if len(os.Args) != 2 {
		log.Fatal("NPPES csv file not provided as argument.")
	}
	fname := os.Args[1]
	err = store.DeleteAllNPIOrganizations(ctx)
	failOnError(err)
	_, err = nppesquerier.ParseAndStoreNPIFile(ctx, fname, store)
	failOnError(err)
}
