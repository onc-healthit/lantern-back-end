package main

import (
	"context"
	"os"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/datacleanup"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	err := config.SetupConfig()
	helpers.FailOnError("Error setting up config", err)

	ctx := context.Background()
	store, err := postgresql.NewStore(
		viper.GetString("dbhost"),
		viper.GetInt("dbport"),
		viper.GetString("dbuser"),
		viper.GetString("dbpassword"),
		viper.GetString("dbname"),
		viper.GetString("dbsslmode"),
	)
	helpers.FailOnError("Error connecting to DB", err)
	defer store.Close()

	log.Info("Starting stale CHPL data cleanup...")

	// Check if start time was passed as argument
	var cutoffTime time.Time
	if len(os.Args) >= 2 {
		// argument1: name of file & argument2: population start time
		startTimeStr := os.Args[1]
		cutoffTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			log.Fatalf("Invalid start time format: %v. Expected RFC3339 format.", err)
		}
		log.Infof("Using provided population start time as cutoff: %v", cutoffTime)
	} else {
		// Fallback for manual testing/debugging (only runs when no time is provided with the main.go file)
		cutoffTime = time.Now().Add(-1 * time.Hour) //Delete anything older than 1 hour
		log.Warnf("No start time provided, using fallback cutoff: %v", cutoffTime)
	}

	// Optional safety check
	if time.Since(cutoffTime) > 24*time.Hour {
		log.Warnf("Cutoff time is more than 24 hours ago (%v), this might delete a lot of data", cutoffTime)
	}

	err = datacleanup.CleanupStaleData(ctx, store, cutoffTime)
	if err != nil {
		log.Fatalf("Failed to cleanup stale data: %v", err)
	}

	log.Info("Stale CHPL data cleanup completed successfully.")
}
