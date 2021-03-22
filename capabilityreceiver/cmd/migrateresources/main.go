package main

import (
	"context"
	"database/sql"
	"fmt"
	"encoding/json"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/capabilityreceiver/pkg/capabilityhandler"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/workers"
)

var updateFHIREndpointInfoHistoryStatement *sql.Stmt

// Result is the value that is returned from getting the history data from the
// given URL
type Result struct {
	URL  string
}

type historyArgs struct {
	fhirURL string
	store   *postgresql.Store
	result  chan Result
}

// @TODO creates jobs for the workers so that each worker gets the history data
// for a specified url
func createJobs(ctx context.Context,
	ch chan Result,
	urls []string,
	store *postgresql.Store,
	allWorkers *workers.Workers) {
	for index := range urls {
		jobArgs := make(map[string]interface{})
		jobArgs["historyArgs"] = historyArgs{
			fhirURL: urls[index],
			store:   store,
			result:  ch,
		}

		job := workers.Job{
			Context:     ctx,
			Duration:    time.Duration(120) * time.Second,
			Handler:     getHistory,
			HandlerArgs: &jobArgs,
		}

		err := allWorkers.Add(&job)
		if err != nil {
			log.Warnf("Error while adding job for getting history for URL %s, %s", urls[index], err)
		}
	}
}

// getHistory gets the database history of a specified url
func getHistory(ctx context.Context, args *map[string]interface{}) error {
	ha, ok := (*args)["historyArgs"].(historyArgs)
	if !ok {
		return fmt.Errorf("unable to cast arguments to type historyArgs")
	}

	updateFHIREndpointInfoHistoryStatement, err := ha.store.DB.Prepare(`
		UPDATE fhir_endpoints_info_history
		SET
			operation_resource = $1
		WHERE entered_at = $2;`)
	if err != nil {
		return fmt.Errorf("unable to prepare FHIR Endpoint History Update statement: %s", err)
	}
	defer updateFHIREndpointInfoHistoryStatement.Close()

	// Get everything from the fhir_endpoints_info_history table for the given URL
	selectHistory := `SELECT entered_at, capability_statement
		FROM fhir_endpoints_info_history
		WHERE url=$1;`
	historyRows, err := ha.store.DB.QueryContext(ctx, selectHistory, ha.fhirURL)
	if err != nil {
		log.Warnf("Failed getting the history rows for URL %s. Error: %s", ha.fhirURL, err)
		result := Result{
			URL:  ha.fhirURL,
		}
		ha.result <- result
		return nil
	}

	// Puts the rows in an array and sends it back on the channel to be processed
	defer historyRows.Close()
	for historyRows.Next() {
		var enteredTime time.Time
		var capStat []byte
		err = historyRows.Scan(&enteredTime, &capStat)
		if err != nil {
			log.Warnf("Error while scanning the rows of the history table for URL %s. Error: %s", ha.fhirURL, err)
			continue
		}

		/**
			@TODO Figure out how to convert capability statement to supported resources
			1. convert []byte to map[string]interface{}
			2. Call RunSupportedResourcesChecks
			3. Put the operation_resource value into the history table at the given entered_at value
		*/
		var capInt map[string]interface{}
		if capStat != nil && len(capStat) > 0 {
			_ = json.Unmarshal(capStat, &capInt)
		}
		_, operationResource := capabilityhandler.RunSupportedResourcesChecks(capInt)
		operResourceJSON, err := json.Marshal(operationResource)
		if err != nil {
			log.Warnf("Error while convering operationResource to JSON, %+v, URL %s at %s. Error: %s", operationResource, err)
			continue
		}
		_, err = updateFHIREndpointInfoHistoryStatement.ExecContext(ctx, operResourceJSON, enteredTime)
		if err != nil {
			log.Warnf("Error while updating the row of the history table for URL %s at %s. Error: %s", ha.fhirURL, enteredTime.String(), err)
		}
	}
	result := Result{
		URL:  ha.fhirURL,
	}
	ha.result <- result
	return nil
}


func main() {
	err := config.SetupConfig()
	helpers.FailOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	log.Info("Successfully connected to DB!")

	ctx := context.Background()

	// Get everything from the fhir_endpoints_info table
	sqlQuery := "SELECT DISTINCT url FROM fhir_endpoints_info_history;"
	rows, err := store.DB.QueryContext(ctx, sqlQuery)
	helpers.FailOnError("Make sure that the database is not empty. Error:", err)

	// Put into an object
	var urls []string
	defer rows.Close()
	for rows.Next() {
		var currURL string
		err = rows.Scan(&currURL)
		helpers.FailOnError("Error scanning the row. Error:", err)

		urls = append(urls, currURL)
	}

	// @TODO: Go through all elements of the history database, will have to use workers
	errs := make(chan error)
	numWorkers := 50
	allWorkers := workers.NewWorkers()

	// Start workers
	err = allWorkers.Start(ctx, numWorkers, errs)
	helpers.FailOnError("Error from starting workers. Error:", err)

	resultCh := make(chan Result)
	go createJobs(ctx, resultCh, urls, store, allWorkers)

	// Add the results from createJobs to mapURLHistory
	count := 0
	for _ = range resultCh {
		if count == len(urls)-1 {
			close(resultCh)
		}
		count++
	}
}
