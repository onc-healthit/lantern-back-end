package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/archivefile"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	var dateStart string
	var dateEnd string
	var writeFile string

	if len(os.Args) >= 3 {
		dateStart = os.Args[1]
		dateEnd = os.Args[2]
		writeFile = os.Args[3]
	} else {
		log.Fatalf("ERROR: Missing command-line arguments")
	}

	err := config.SetupConfig()
	helpers.FailOnError("", err)

	// Verify that given dates are in the correct format
	layout := "2006-01-02"
	_, err = time.Parse(layout, dateStart)
	helpers.FailOnError("ERROR: Start date not in correct format", err)
	_, err = time.Parse(layout, dateEnd)
	helpers.FailOnError("ERROR: End date not in correct format", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	log.Info("Successfully connected to DB!")

	ctx := context.Background()

	entries, err := archivefile.CreateArchive(ctx, store, dateStart, dateEnd)
	helpers.FailOnError("", err)

	// Format as JSON
	finalFormatJSON, err := json.MarshalIndent(entries, "", "\t")
	helpers.FailOnError("", err)

	// Write to the given file
	err = ioutil.WriteFile(writeFile, finalFormatJSON, 0644)
	helpers.FailOnError("ERROR: Writing to given file failed", err)
}
