package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/capabilityreceiver/pkg/capabilityhandler"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/workers"
)

// Result is the value that is returned from getting the history data from the
// given URL
type Result struct {
	URL string
}

type historyArgs struct {
	fhirURL   string
	store     *postgresql.Store
	result    chan Result
	isHistory bool
}

// creates jobs for the workers so that each worker updates the correct field based
// on the given migrateDirection
func createJobs(ctx context.Context,
	ch chan Result,
	urls []string,
	store *postgresql.Store,
	allWorkers *workers.Workers,
	migrateDirection string,
	isHistory bool) {
	for index := range urls {
		jobArgs := make(map[string]interface{})
		jobArgs["historyArgs"] = historyArgs{
			fhirURL:   urls[index],
			store:     store,
			result:    ch,
			isHistory: isHistory,
		}

		handlerFunction := updateOperationResource
		if migrateDirection == "down" {
			handlerFunction = updateSupportedResources
		}

		job := workers.Job{
			Context:     ctx,
			Duration:    time.Duration(480) * time.Second,
			Handler:     handlerFunction,
			HandlerArgs: &jobArgs,
		}

		err := allWorkers.Add(&job)
		if err != nil {
			log.Warnf("Error while adding job for getting history for URL %s, %s", urls[index], err)
		}
	}
}

// updateOperationResource gets the history data for a given URL and creates the
// operation_resource field data based on each row's capability statement
func updateOperationResource(ctx context.Context, args *map[string]interface{}) error {
	ha, ok := (*args)["historyArgs"].(historyArgs)
	if !ok {
		log.Warnf("unable to cast arguments to type historyArgs")
		result := Result{
			URL: "unknown",
		}
		ha.result <- result
		return nil
	}

	databaseTable := "fhir_endpoints_info"
	if ha.isHistory {
		databaseTable = "fhir_endpoints_info_history"
	}

	updateFHIREndpointInfoHistoryStatement, err := ha.store.DB.Prepare(`
		UPDATE ` + databaseTable + `
		SET
			operation_resource = $1
		WHERE updated_at = $2 AND url = $3;`)
	if err != nil {
		log.Warnf("unable to prepare FHIR Endpoint History Update statement %s. Error: %s", ha.fhirURL, err)
		result := Result{
			URL: ha.fhirURL,
		}
		ha.result <- result
		return nil
	}
	defer updateFHIREndpointInfoHistoryStatement.Close()

	// Get everything from the fhir_endpoints_info_history table for the given URL
	selectHistory := `SELECT updated_at, capability_statement
		FROM ` + databaseTable + `
		WHERE url=$1;`
	historyRows, err := ha.store.DB.QueryContext(ctx, selectHistory, ha.fhirURL)
	if err != nil {
		log.Warnf("Failed getting the history rows for URL %s. Error: %s", ha.fhirURL, err)
		result := Result{
			URL: ha.fhirURL,
		}
		ha.result <- result
		return nil
	}

	defer historyRows.Close()
	for historyRows.Next() {
		var updatedTime time.Time
		var capStat []byte
		err = historyRows.Scan(&updatedTime, &capStat)
		if err != nil {
			log.Warnf("Error while scanning the rows of the history table for URL %s. Error: %s", ha.fhirURL, err)
			continue
		}

		// Unmarshal the capStat, we continue if there's an error because RunSupportedResourceChecks
		// handles a nil value and it makes more sense to put an empty array in the database than
		// a nil value
		var capInt map[string]interface{}
		if len(capStat) > 0 {
			err = json.Unmarshal(capStat, &capInt)
			if err != nil {
				log.Warnf("Error while unmarshalling the rows of the history table for URL %s. Error: %s", ha.fhirURL, err)
			}
		}
		operationResource := capabilityhandler.RunSupportedResourcesChecks(capInt)
		operResourceJSON, err := json.Marshal(operationResource)
		if err != nil {
			log.Warnf("Error while convering operationResource to JSON, %+v, Error: %s", operationResource, err)
			continue
		}
		_, err = updateFHIREndpointInfoHistoryStatement.ExecContext(ctx, operResourceJSON, updatedTime, ha.fhirURL)
		if err != nil {
			log.Warnf("Error while updating the row of the history table for URL %s at %s. Error: %s", ha.fhirURL, updatedTime.String(), err)
		}
	}
	result := Result{
		URL: ha.fhirURL,
	}
	ha.result <- result
	return nil
}

// updateSupportedResources gets the history data for a given URL and creates the
// supported_resources field data based on each row's capability statement
func updateSupportedResources(ctx context.Context, args *map[string]interface{}) error {
	ha, ok := (*args)["historyArgs"].(historyArgs)
	if !ok {
		log.Warnf("unable to cast arguments to type historyArgs")
		result := Result{
			URL: "unknown",
		}
		ha.result <- result
		return nil
	}

	databaseTable := "fhir_endpoints_info"
	if ha.isHistory {
		databaseTable = "fhir_endpoints_info_history"
	}

	updateFHIREndpointInfoHistoryStatement, err := ha.store.DB.Prepare(`
		UPDATE ` + databaseTable + `
		SET
			supported_resources = $1
		WHERE updated_at = $2 AND url = $3;`)
	if err != nil {
		log.Warnf("unable to prepare FHIR Endpoint History Update statement %s. Error: %s", ha.fhirURL, err)
		result := Result{
			URL: ha.fhirURL,
		}
		ha.result <- result
		return nil
	}
	defer updateFHIREndpointInfoHistoryStatement.Close()

	// Get everything from the fhir_endpoints_info_history table for the given URL
	selectHistory := `SELECT updated_at, capability_statement
		FROM ` + databaseTable + `
		WHERE url=$1;`
	historyRows, err := ha.store.DB.QueryContext(ctx, selectHistory, ha.fhirURL)
	if err != nil {
		log.Warnf("Failed getting the history rows for URL %s. Error: %s", ha.fhirURL, err)
		result := Result{
			URL: ha.fhirURL,
		}
		ha.result <- result
		return nil
	}

	defer historyRows.Close()
	for historyRows.Next() {
		var updatedTime time.Time
		var capStat []byte
		err = historyRows.Scan(&updatedTime, &capStat)
		if err != nil {
			log.Warnf("Error while scanning the rows of the history table for URL %s. Error: %s", ha.fhirURL, err)
			continue
		}

		// Unmarshal the capStat, we continue if there's an error because createSupportedResources
		// handles a nil value and it makes more sense to put an empty array in the database than
		// a nil value
		var capInt map[string]interface{}
		if len(capStat) > 0 {
			err = json.Unmarshal(capStat, &capInt)
			if err != nil {
				log.Warnf("Error while unmarshalling the rows of the history table for URL %s. Error: %s", ha.fhirURL, err)
			}
		}
		supportedResources := createSupportedResources(capInt)
		_, err = updateFHIREndpointInfoHistoryStatement.ExecContext(ctx, pq.Array(supportedResources), updatedTime, ha.fhirURL)
		if err != nil {
			log.Warnf("Error while updating the row of the history table for URL %s at %s. Error: %s", ha.fhirURL, updatedTime.String(), err)
		}
	}
	result := Result{
		URL: ha.fhirURL,
	}
	ha.result <- result
	return nil
}

// createSupportedResources creates the supported_resources field data based on the
// given capability statement
func createSupportedResources(capInt map[string]interface{}) []string {
	if capInt == nil {
		return nil
	}
	var supportedResources []string

	if capInt["rest"] == nil {
		return nil
	}
	restArr := capInt["rest"].([]interface{})
	restInt := restArr[0].(map[string]interface{})
	if restInt["resource"] == nil {
		return nil
	}
	resourceArr := restInt["resource"].([]interface{})

	for _, resource := range resourceArr {
		resourceInt := resource.(map[string]interface{})
		if resourceInt["type"] == nil {
			return nil
		}
		resourceType := resourceInt["type"].(string)
		supportedResources = append(supportedResources, resourceType)
	}

	return supportedResources
}

// Migrate the fhir_endpoints_info table and fhir_endpoints_info_history table based
// on the given migration direction, either populating the operation_resource field if
// it's an "up" migration or supported_resources if it's a "down" migration
func main() {
	var migrateDirection string

	if len(os.Args) >= 1 {
		migrateDirection = os.Args[1]
	} else {
		log.Fatalf("ERROR: Missing command-line arguments")
	}

	err := config.SetupConfig()
	helpers.FailOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	log.Info("Successfully connected to DB!")

	ctx := context.Background()

	// Get all URLs from the fhir_endpoints_info_history table
	sqlQuery := "SELECT DISTINCT url FROM fhir_endpoints_info_history;"
	rows, err := store.DB.QueryContext(ctx, sqlQuery)
	helpers.FailOnError("Make sure that the database is not empty. Error:", err)

	var urls []string
	defer rows.Close()
	for rows.Next() {
		var currURL string
		err = rows.Scan(&currURL)
		helpers.FailOnError("Error scanning the row. Error:", err)

		urls = append(urls, currURL)
	}

	errs := make(chan error)
	numWorkers := 25
	allWorkers := workers.NewWorkers()

	// Start workers
	err = allWorkers.Start(ctx, numWorkers, errs)
	helpers.FailOnError("Error from starting workers. Error:", err)

	resultCh := make(chan Result)
	go createJobs(ctx, resultCh, urls, store, allWorkers, migrateDirection, true)

	// Close the channel once we've received all results
	count := 0
	for range resultCh {
		if count == len(urls)-1 {
			close(resultCh)
		}
		count++
	}

	// Disable the add_fhir_endpoint_info_history_trigger so updating the fhir_endpoints_info
	// data does not add another entry in the fhir_endpoints_info_history table
	infoHistoryTriggerDisable := `
	ALTER TABLE fhir_endpoints_info
	DISABLE TRIGGER add_fhir_endpoint_info_history_trigger;`
	_, err = store.DB.ExecContext(ctx, infoHistoryTriggerDisable)
	helpers.FailOnError("Error from disabling trigger. Error:", err)

	// Disable the set_timestamp_fhir_endpoints_info so updating the fhir_endpoints_info
	// data does not change the "updated_at" field for each row
	infoTimeTriggerDisable := `
	ALTER TABLE fhir_endpoints_info
	DISABLE TRIGGER set_timestamp_fhir_endpoints_info;`
	_, err = store.DB.ExecContext(ctx, infoTimeTriggerDisable)
	helpers.FailOnError("Error from disabling time trigger. Error:", err)

	sqlQuery = "SELECT DISTINCT url FROM fhir_endpoints_info;"
	rows, err = store.DB.QueryContext(ctx, sqlQuery)
	helpers.FailOnError("Make sure that the database is not empty. Error:", err)

	var urls2 []string
	defer rows.Close()
	for rows.Next() {
		var currURL string
		err = rows.Scan(&currURL)
		helpers.FailOnError("Error scanning the row. Error:", err)

		urls2 = append(urls2, currURL)
	}

	resultCh2 := make(chan Result)
	go createJobs(ctx, resultCh2, urls2, store, allWorkers, migrateDirection, false)

	// Close the channel once we've received all results
	count = 0
	for range resultCh2 {
		if count == len(urls2)-1 {
			close(resultCh2)
		}
		count++
	}

	infoHistoryTriggerEnable := `
	ALTER TABLE fhir_endpoints_info
	ENABLE TRIGGER add_fhir_endpoint_info_history_trigger;`
	_, err = store.DB.ExecContext(ctx, infoHistoryTriggerEnable)
	helpers.FailOnError("Error from enabling trigger. Error:", err)

	infoTimeTriggerEnable := `
	ALTER TABLE fhir_endpoints_info
	ENABLE TRIGGER set_timestamp_fhir_endpoints_info;`
	_, err = store.DB.ExecContext(ctx, infoTimeTriggerEnable)
	helpers.FailOnError("Error from disabling time trigger. Error:", err)

	log.Info("Successfully migrated data!")
}
