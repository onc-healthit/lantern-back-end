package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/capabilityreceiver/pkg/capabilityhandler/validation"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

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

type validationArgs struct {
	updatedTime       time.Time
	capStatByte       []byte
	tlsVersion        string
	mimeTypes         []string
	httpResponse      int
	smartHttpResponse int
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
		jobArgs["historyArgs"] = historyArgs{
			fhirURL:   urls[index],
			store:     store,
			result:    ch,
			isHistory: isHistory,
		}

		handlerFunction := addToValidationTable
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

// addToValidationTable gets the history data for a given URL and creates the
// validation table rows based on each row's capability statement
func addToValidationTable(ctx context.Context, args *map[string]interface{}) error {
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

	addValidationResultStatement, err := ha.store.DB.Prepare(`
		Insert into validation_results (id)
		VALUES (DEFAULT)
		RETURNING id;`)
	if err != nil {
		log.Warnf("unable to prepare Add Validation Result statement %s. Error: %s", ha.fhirURL, err)
		result := Result{
			URL: ha.fhirURL,
		}
		ha.result <- result
		return nil
	}
	defer addValidationResultStatement.Close()

	addValidationStatement, err := ha.store.DB.Prepare(`
		INSERT INTO validations (
			rule_name,
			valid,
			expected,
			actual,
			comment,
			reference,
			implementation_guide,
			validation_result_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)
	if err != nil {
		log.Warnf("unable to prepare Add Validation statement %s. Error: %s", ha.fhirURL, err)
		result := Result{
			URL: ha.fhirURL,
		}
		ha.result <- result
		return nil
	}
	defer addValidationStatement.Close()

	updateFHIREndpointValResStatement, err := ha.store.DB.Prepare(`
		UPDATE ` + databaseTable + `
		SET
			validation_result_id = $1
		WHERE updated_at = $2 AND url = $3;`)
	if err != nil {
		log.Warnf("unable to prepare FHIR Endpoint History Update statement %s. Error: %s", ha.fhirURL, err)
		result := Result{
			URL: ha.fhirURL,
		}
		ha.result <- result
		return nil
	}
	defer updateFHIREndpointValResStatement.Close()

	// Get everything from the fhir_endpoints_info_history table for the given URL
	// @TODO Could change this to endpoint_export for fhir_endpoints_info
	selectHistory := `SELECT endpts_info.capability_statement,
			endpts_info.tls_version, endpts_info.mime_types, endpts_metadata.http_response,
			endpts_metadata.smart_http_response,
			endpts_info.updated_at AS INFO_UPDATED
		FROM ` + databaseTable + ` AS endpts_info
		LEFT JOIN fhir_endpoints_metadata AS endpts_metadata ON endpts_info.metadata_id = endpts_metadata.id
		WHERE endpts_info.url=$1;`
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
	var validationVals []validationArgs
	for historyRows.Next() {
		// @TODO Change this?
		var val validationArgs
		err = historyRows.Scan(&val.capStatByte,
			&val.tlsVersion,
			pq.Array(&val.mimeTypes),
			&val.httpResponse,
			&val.smartHttpResponse,
			&val.updatedTime)
		if err != nil {
			log.Warnf("Error while scanning the rows of the history table for URL %s. Error: %s", ha.fhirURL, err)
			continue
		}
		validationVals = append(validationVals, val)
	}

	for _, val := range validationVals {
		// Create the capability statement object
		var capStat capabilityparser.CapabilityStatement
		var capInt map[string]interface{}
		if len(val.capStatByte) > 0 {
			err = json.Unmarshal(val.capStatByte, &capInt)
			if err != nil {
				log.Warnf("Error while unmarshalling the rows of the history table for URL %s. Error: %s", ha.fhirURL, err)
			}

			capStat, err = capabilityparser.NewCapabilityStatementFromInterface(capInt)
			if err != nil {
				log.Warnf("unable to parse CapabilityStatement out of message with url %s. Error: %s", ha.fhirURL, err)
			}
		}

		// Get the FHIR Version from that
		// @TODO update this based on Emily's PR
		fhirVersion := ""
		if capStat != nil {
			fhirVersion, _ = capStat.GetFHIRVersion()
		}
		// Create validator
		validator := validation.ValidatorForFHIRVersion(fhirVersion)
		// RunValidation
		validationObj := validator.RunValidation(capStat, val.httpResponse, val.mimeTypes, fhirVersion, val.tlsVersion, val.smartHttpResponse)
		// Somehow have to get an ID from the validation_results table
		var valResID int
		valRow, err := addValidationResultStatement.QueryContext(ctx)
		if err == nil {
			defer valRow.Close()
			count := 0
			for valRow.Next() {
				if count > 0 {
					log.Warnf("More than 1 ID added for URL", ha.fhirURL)
					break
				}
				err = valRow.Scan(&valResID)
				count++
			}
		}
		if err != nil {
			log.Warnf("Failed to add a new ID. Error: %s", err)
			result := Result{
				URL: ha.fhirURL,
			}
			ha.result <- result
			return nil
		}
		log.Infof("ID: %d", valResID)
		log.Infof("url: %s", ha.fhirURL)
		// Then add each element of this array as a row to the table with that id
		for _, ruleInfo := range validationObj.Results {
			_, err = addValidationStatement.QueryContext(ctx,
				ruleInfo.RuleName,
				ruleInfo.Valid,
				ruleInfo.Expected,
				ruleInfo.Actual,
				ruleInfo.Comment,
				ruleInfo.Reference,
				ruleInfo.ImplGuide,
				valResID)
			if err != nil {
				log.Warnf("Failed to add validation for URL %s. Error: %s", ha.fhirURL, err)
			}
		}
		// Then have to update the row in the history table with that id
		log.Infof("---------------")
		_, err = updateFHIREndpointValResStatement.ExecContext(ctx, valResID, val.updatedTime, ha.fhirURL)
		if err != nil {
			log.Warnf("Error while updating the row of the table for URL %s at %s. Error: %s", ha.fhirURL, val.updatedTime.String(), err)
		}
	}
	result := Result{
		URL: ha.fhirURL,
	}
	ha.result <- result
	return nil
}

// addToValidationField gets the history data for a given URL and creates the
// validation field data based on each row's capability statement
func addToValidationField(ctx context.Context, args *map[string]interface{}) error {
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

	// @TODO Update based on Emily's PR
	updateFHIREndpointInfoHistoryStatement, err := ha.store.DB.Prepare(`
		UPDATE ` + databaseTable + `
		SET
			validation = $1
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
	selectHistory := `SELECT endpts_info.capability_statement,
			endpts_info.tls_version, endpts_info.mime_types, endpts_metadata.http_response,
			endpts_metadata.smart_http_response,
			endpts_info.updated_at AS INFO_UPDATED
		FROM ` + databaseTable + ` AS endpts_info
		LEFT JOIN fhir_endpoints_metadata AS endpts_metadata ON endpts_info.metadata_id = endpts_metadata.id
		WHERE endpts_info.url=$1;`
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
	var validationVals []validationArgs
	for historyRows.Next() {
		// @TODO Change this?
		var val validationArgs
		err = historyRows.Scan(&val.capStatByte,
			&val.tlsVersion,
			pq.Array(&val.mimeTypes),
			&val.httpResponse,
			&val.smartHttpResponse,
			&val.updatedTime)
		if err != nil {
			log.Warnf("Error while scanning the rows of the history table for URL %s. Error: %s", ha.fhirURL, err)
			continue
		}
		validationVals = append(validationVals, val)
	}

	for _, val := range validationVals {
		// Create the capability statement object
		var capStat capabilityparser.CapabilityStatement
		var capInt map[string]interface{}
		if len(val.capStatByte) > 0 {
			err = json.Unmarshal(val.capStatByte, &capInt)
			if err != nil {
				log.Warnf("Error while unmarshalling the rows of the history table for URL %s. Error: %s", ha.fhirURL, err)
			}

			capStat, err = capabilityparser.NewCapabilityStatementFromInterface(capInt)
			if err != nil {
				log.Warnf("unable to parse CapabilityStatement out of message with url %s. Error: %s", ha.fhirURL, err)
			}
		}

		// Get the FHIR Version from that
		// @TODO update this based on Emily's PR
		fhirVersion := ""
		if capStat != nil {
			fhirVersion, _ = capStat.GetFHIRVersion()
		}
		// Create validator
		validator := validation.ValidatorForFHIRVersion(fhirVersion)
		// RunValidation
		validationObj := validator.RunValidation(capStat, val.httpResponse, val.mimeTypes, fhirVersion, val.tlsVersion, val.smartHttpResponse)

		// convert object to JSON?
		validationJSON, err := json.Marshal(validationObj)
		if err != nil {
			log.Warnf("Error marshalling object to JSON. Error: %s", err)
			continue
		}

		// Then add the object to the fhir_endpoint_info row
		_, err = updateFHIREndpointInfoHistoryStatement.QueryContext(ctx,
			validationJSON,
			val.updatedTime,
			ha.fhirURL)
		if err != nil {
			log.Warnf("Failed to add validation for URL %s. Error: %s", ha.fhirURL, err)
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
// on the given migration direction, either populating the validation table if
// it's an "up" migration or validation field in fhir_endpoints_info if it's a "down" migration
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
	numWorkers := 1
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
