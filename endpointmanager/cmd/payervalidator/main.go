// cmd/payervalidator/main.go
package main

import (
	"context"
	"flag"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/payervalidator"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	var (
		dryRun = flag.Bool("dry-run", false, "Perform validation without updating database")
	)
	flag.Parse()

	var err error

	// Setup configuration
	err = config.SetupConfig()
	helpers.FailOnError("Failed to setup config", err)

	// Establish database connection
	store, err := postgresql.NewStore(
		viper.GetString("dbhost"),
		viper.GetInt("dbport"),
		viper.GetString("dbuser"),
		viper.GetString("dbpassword"),
		viper.GetString("dbname"),
		viper.GetString("dbsslmode"),
	)
	helpers.FailOnError("Failed to connect to database", err)
	defer store.Close()
	log.Info("Successfully connected to database!")

	// Initialize payer validator with the store
	validator, err := payervalidator.NewValidatorWithStore(store)
	helpers.FailOnError("Failed to initialize payer validator", err)
	defer validator.Close()

	ctx := context.Background()

	if *dryRun {
		log.Info("Running in dry-run mode - no database updates will be made")
		err = validator.ValidateRegistrationsDryRun(ctx)
	} else {
		err = validator.ValidateAndProcessRegistrations(ctx)
	}

	helpers.FailOnError("Payer validation failed", err)
	log.Info("Payer registration validation completed successfully")
}
