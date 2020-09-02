package main

import (
	"context"
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/nppesquerier"
	"github.com/onc-healthit/lantern-back-end/sharedfunctions"
)

func main() {
	err := config.SetupConfig()
	sharedfunctions.failOnError("", err)

	ctx := context.Background()
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	sharedfunctions.failOnError("", err)
	if len(os.Args) != 2 {
		log.Fatal("NPPES contact csv file not provided as argument.")
	}
	fname := os.Args[1]
	err = store.DeleteAllNPIContacts(ctx)
	sharedfunctions.failOnError("", err)

	log.Info("Adding NPI FHIR URLs to database...")
	added, err := nppesquerier.ParseAndStoreNPIContactsFile(ctx, fname, store)
	log.Infof("Added %d NPI FHIR URLs to database\n", added)
	sharedfunctions.failOnError("", err)
}
