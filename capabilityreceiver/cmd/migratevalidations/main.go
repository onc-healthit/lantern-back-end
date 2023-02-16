package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"time"

	"github.com/lib/pq"

	"github.com/onc-healthit/lantern-back-end/capabilityreceiver/pkg/capabilityhandler/validation"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/smartparser"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/workers"
)

var updateInfoValResStatement *sql.Stmt
var updateHistoryValResStatement *sql.Stmt

// Result is the value that is returned from getting the history data from the
// given URL
type Result struct {
	URL string
}

type workerArgs struct {
	fhirURL   string
	store     *postgresql.Store
	result    chan Result
	isHistory bool
}

type validationArgs struct {
	updatedTime       time.Time
	capStatByte       []byte
	tlsVersion        string
	mimeTypes         []string
	smartResponseByte []byte
}

// prepares the statements used in the "up" migration
func prepareUpStatements(s *postgresql.Store) error {
	var err error

	updateInfoValResStatement, err = s.DB.Prepare(`
		UPDATE fhir_endpoints_info
		SET
			validation_result_id = $1
		WHERE url = $2;`)
	if err != nil {
		return err
	}
	updateHistoryValResStatement, err = s.DB.Prepare(`
		UPDATE fhir_endpoints_info_history
		SET
			validation_result_id = $1
		WHERE updated_at = $2 AND url = $3;`)
	if err != nil {
		return err
	}

	return nil
}

func returnResult(wa workerArgs) error {
	result := Result{
		URL: wa.fhirURL,
	}
	wa.result <- result
	return nil
}

// creates jobs for the workers so that each worker updates the correct object based
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
		jobArgs["workerArgs"] = workerArgs{
			fhirURL:   urls[index],
			store:     store,
			result:    ch,
			isHistory: isHistory,
		}

		handlerFunction := addToValidationTableInfo
		if isHistory {
			handlerFunction = addToValidationTableHistory
		}
		if migrateDirection == "down" {
			handlerFunction = addToValidationField
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

// addToValidationTableHistory gets the history table data for a given URL and creates the
// validation table rows based on each row's capability statement
func addToValidationTableHistory(ctx context.Context, args *map[string]interface{}) error {
	wa, ok := (*args)["workerArgs"].(workerArgs)
	if !ok {
		log.Warnf("unable to cast arguments to type workerArgs")
		result := Result{
			URL: "unknown",
		}
		wa.result <- result
		return nil
	}

	// Get validation information from the specified table table for the given URL
	selectHistory := `SELECT capability_statement, tls_version, mime_types,
			smart_response, updated_at AS INFO_UPDATED
		FROM fhir_endpoints_info_history
		WHERE url=$1;`
	historyRows, err := wa.store.DB.QueryContext(ctx, selectHistory, wa.fhirURL)
	if err != nil {
		log.Warnf("Failed getting the history rows for URL %s. Error: %s", wa.fhirURL, err)
		return returnResult(wa)
	}
	defer historyRows.Close()
	var validationVals []validationArgs
	for historyRows.Next() {
		var val validationArgs
		err = historyRows.Scan(&val.capStatByte,
			&val.tlsVersion,
			pq.Array(&val.mimeTypes),
			&val.smartResponseByte,
			&val.updatedTime)
		if err != nil {
			log.Warnf("Error while scanning the rows of the history table for URL %s. Error: %s", wa.fhirURL, err)
			continue
		}
		validationVals = append(validationVals, val)
	}

	for _, val := range validationVals {
		// Create the capability statement object
		capStat, err := capabilityparser.NewCapabilityStatement(val.capStatByte)
		if err != nil {
			log.Warnf("unable to parse CapabilityStatement out of message with url %s. Error: %s", wa.fhirURL, err)
		}

		// Create smart response object
		smartResp, err := smartparser.NewSMARTResp(val.smartResponseByte)
		if err != nil {
			log.Warnf("Error while unmarshalling the smart response for URL %s. Error: %s", wa.fhirURL, err)
		}

		fhirVersion := ""
		if capStat != nil {
			fhirVersion, _ = capStat.GetFHIRVersion()
		}

		validator := validation.ValidatorForFHIRVersion(fhirVersion)
		validationObj := validator.RunValidation(capStat, fhirVersion, val.tlsVersion, smartResp, "None", "None")
		valResID, err := wa.store.AddValidationResult(ctx)
		if err != nil {
			log.Warnf("Failed to add a new ID. Error: %s", err)
			return returnResult(wa)
		}
		// Then add each element of the validationObj array as a row to the table with that id
		err = wa.store.AddValidation(ctx, &validationObj, valResID)
		if err != nil {
			log.Warnf("Failed to add validation for URL %s. Error: %s", wa.fhirURL, err)
			return returnResult(wa)
		}

		// Then have to update the row in the history table with that id
		_, err = updateHistoryValResStatement.ExecContext(ctx, valResID, val.updatedTime, wa.fhirURL)
		if err != nil {
			log.Warnf("Error while updating the row of the table for URL %s at %s. Error: %s", wa.fhirURL, val.updatedTime.String(), err)
			return returnResult(wa)
		}
	}
	return returnResult(wa)
}

// since the current data in info table is also in the history table, get the ID
// that was generated for the associated history table row and use that for the
// info table
func addToValidationTableInfo(ctx context.Context, args *map[string]interface{}) error {
	wa, ok := (*args)["workerArgs"].(workerArgs)
	if !ok {
		log.Warnf("unable to cast arguments to type workerArgs")
		result := Result{
			URL: "unknown",
		}
		wa.result <- result
		return nil
	}

	selectHistory := `SELECT validation_result_id FROM fhir_endpoints_info_history
		WHERE url = $1
		ORDER BY entered_at DESC
		LIMIT 1;`
	valResRow := wa.store.DB.QueryRowContext(ctx, selectHistory, wa.fhirURL)
	valResID := 0
	err := valResRow.Scan(&valResID)
	if err != nil {
		log.Warnf("Failed to get the validation_result_id. Error: %s", err)
		return returnResult(wa)
	}
	_, err = updateInfoValResStatement.ExecContext(ctx, valResID, wa.fhirURL)
	if err != nil {
		log.Warnf("Error while updating the row of fhir_endpoints_info table for URL %s. Error: %s", wa.fhirURL, err)
	}
	return returnResult(wa)
}

// addToValidationField gets the table data for a given URL and creates the
// validation field data based on each row's capability statement
func addToValidationField(ctx context.Context, args *map[string]interface{}) error {
	wa, ok := (*args)["workerArgs"].(workerArgs)
	if !ok {
		log.Warnf("unable to cast arguments to type workerArgs")
		result := Result{
			URL: "unknown",
		}
		wa.result <- result
		return nil
	}

	databaseTable := "fhir_endpoints_info"
	if wa.isHistory {
		databaseTable = "fhir_endpoints_info_history"
	}

	updateValidationStatement := `
		UPDATE ` + databaseTable + `
		SET
			validation = $1
		WHERE updated_at = $2 AND url = $3;`

	// Get all necessary validation data from the specified table for the given URL
	selectHistory := `SELECT capability_statement, tls_version, mime_types,
		smart_response, updated_at AS INFO_UPDATED
		FROM ` + databaseTable + `
		WHERE url=$1;`
	historyRows, err := wa.store.DB.QueryContext(ctx, selectHistory, wa.fhirURL)
	if err != nil {
		log.Warnf("Failed getting the history rows for URL %s. Error: %s", wa.fhirURL, err)
		return returnResult(wa)
	}

	defer historyRows.Close()
	var validationVals []validationArgs
	for historyRows.Next() {
		var val validationArgs
		err = historyRows.Scan(&val.capStatByte,
			&val.tlsVersion,
			pq.Array(&val.mimeTypes),
			&val.smartResponseByte,
			&val.updatedTime)
		if err != nil {
			log.Warnf("Error while scanning the rows of the history table for URL %s. Error: %s", wa.fhirURL, err)
			continue
		}
		validationVals = append(validationVals, val)
	}

	for _, val := range validationVals {
		// Create the capability statement object
		capStat, err := capabilityparser.NewCapabilityStatement(val.capStatByte)
		if err != nil {
			log.Warnf("unable to parse CapabilityStatement out of message with url %s. Error: %s", wa.fhirURL, err)
		}

		// Create smart response object
		smartResp, err := smartparser.NewSMARTResp(val.smartResponseByte)
		if err != nil {
			log.Warnf("Error while unmarshalling the smart response for URL %s. Error: %s", wa.fhirURL, err)
		}

		fhirVersion := ""
		if capStat != nil {
			fhirVersion, _ = capStat.GetFHIRVersion()
		}
		validator := validation.ValidatorForFHIRVersion(fhirVersion)
		validationObj := validator.RunValidation(capStat, fhirVersion, val.tlsVersion, smartResp, "None", "None")
		validationJSON, err := json.Marshal(validationObj)
		if err != nil {
			log.Warnf("Error marshalling object to JSON. Error: %s", err)
			continue
		}

		// Add the validation object to the specified table
		_, err = wa.store.DB.ExecContext(ctx, updateValidationStatement,
			validationJSON,
			val.updatedTime,
			wa.fhirURL)
		if err != nil {
			log.Warnf("Failed to add validation for URL %s. Error: %s", wa.fhirURL, err)
			return returnResult(wa)
		}
	}

	return returnResult(wa)
}

// Migrate the fhir_endpoints_info table and fhir_endpoints_info_history table based
// on the given migration direction, either populating the validation table if
// it's an "up" migration or validation field in the table if it's a "down" migration
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

	if migrateDirection == "up" {
		err = prepareUpStatements(store)
		helpers.FailOnError("Error when preparing database statements. Error: ", err)
	}

	ctx := context.Background()

	// Get all URLs from the fhir_endpoints_info_history table
	sqlQuery := "SELECT DISTINCT url FROM fhir_endpoints_info_history;"
	rows, err := store.DB.QueryContext(ctx, sqlQuery)
	helpers.FailOnError("Make sure that the database is not empty. Error: ", err)

	var urls []string
	defer rows.Close()
	for rows.Next() {
		var currURL string
		err = rows.Scan(&currURL)
		helpers.FailOnError("Error scanning the row. Error:", err)

		urls = append(urls, currURL)
	}

	errs := make(chan error)
	numWorkers := 10
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
