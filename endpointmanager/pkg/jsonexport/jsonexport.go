package jsonexport

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/workers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type jsonEntry struct {
	URL               string      `json:"url"`
	OrganizationNames []string    `json:"api_information_source_name"`
	CreatedAt         time.Time   `json:"created_at"`
	ListSource        []string    `json:"list_source"`
	VendorName        string      `json:"certified_api_developer_name"`
	Operation         []Operation `json:"operation"`
}

// Operation is a subset of the FHIREndpointInfo and also includes FHIRVersion
type Operation struct {
	HTTPResponse           int                    `json:"http_response"`
	HTTPResponseTimeSecond float64                `json:"http_response_time_second"`
	Errors                 string                 `json:"errors"`
	FHIRVersion            string                 `json:"fhir_version"`
	TLSVersion             string                 `json:"tls_verison"`
	MIMETypes              []string               `json:"mime_types"`
	SupportedResources     []string               `json:"supported_resources"`
	SMARTHTTPResponse      int                    `json:"smart_http_response"`
	SMARTResponse          map[string]interface{} `json:"smart_response"`
	UpdatedAt              time.Time              `json:"updated"`
}

// Result is the value that is returned from getting the history data from the
// given URL
type Result struct {
	URL  string
	Rows []Operation
}

type historyArgs struct {
	fhirURL string
	store   *postgresql.Store
	result  chan Result
}

// CreateJSONExport formats the data from the fhir_endpoints_info and fhir_endpoints_info_history
// tables into a given specification
func CreateJSONExport(ctx context.Context, store *postgresql.Store, fileToWriteTo string) error {
	finalFormatJSON, err := createJSON(ctx, store)
	if err != nil {
		return err
	}
	// Write to the given file
	err = ioutil.WriteFile(fileToWriteTo, finalFormatJSON, 0644)
	return err
}

func createJSON(ctx context.Context, store *postgresql.Store) ([]byte, error) {
	// Get everything from the fhir_endpoints_info table
	sqlQuery := "SELECT DISTINCT url, endpoint_names, info_created, list_source, vendor_name FROM endpoint_export;"
	rows, err := store.DB.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("Make sure that the database is not empty. Error: %s", err)
	}

	// Put into an object
	var urls []string
	entryCheck := make(map[string]jsonEntry)
	defer rows.Close()
	for rows.Next() {
		var entry jsonEntry
		var vendorNameNullable sql.NullString
		var listSource string
		err = rows.Scan(
			&entry.URL,
			pq.Array(&entry.OrganizationNames),
			&entry.CreatedAt,
			&listSource,
			&vendorNameNullable)
		if err != nil {
			return nil, fmt.Errorf("Error scanning the row. Error: %s", err)
		}
		if !vendorNameNullable.Valid {
			entry.VendorName = ""
		}
		// If the URL already exists, include the new list source and organization names
		if val, ok := entryCheck[entry.URL]; ok {
			val.ListSource = append(val.ListSource, listSource)
			val.OrganizationNames = append(val.OrganizationNames, entry.OrganizationNames...)
			entryCheck[entry.URL] = val
		} else {
			entry.ListSource = []string{listSource}
			entryCheck[entry.URL] = entry
			urls = append(urls, entry.URL)
		}
	}

	var entries []jsonEntry
	for _, e := range entryCheck {
		entries = append(entries, e)
	}

	errs := make(chan error)
	numWorkers := viper.GetInt("export_numworkers")
	// If numWorkers not set, default to 10 workers
	if numWorkers == 0 {
		numWorkers = 10
	}
	allWorkers := workers.NewWorkers()

	// Start workers
	err = allWorkers.Start(ctx, numWorkers, errs)
	if err != nil {
		return nil, fmt.Errorf("Error from starting workers. Error: %s", err)
	}

	resultCh := make(chan Result)
	go createJobs(ctx, resultCh, urls, store, allWorkers)

	// Add the results from createJobs to mapURLHistory
	count := 0
	mapURLHistory := make(map[string][]Operation)
	for res := range resultCh {
		if res.URL != "unknown" {
			mapURLHistory[res.URL] = res.Rows
		}
		if count == len(urls)-1 {
			close(resultCh)
		}
		count++
	}

	// Add each array of rows to the Operation field in the entries
	for i, v := range entries {
		url := v.URL
		if val, ok := mapURLHistory[url]; ok {
			entries[i].Operation = val
		}
	}

	// Convert the object to JSON using proper tab formatting
	finalFormatJSON, err := json.MarshalIndent(entries, "", "\t")
	return finalFormatJSON, err
}

// Get the FHIR Version from the capability statement
func getFHIRVersion(capStat []byte) string {
	// Get the FHIR Version from the capability statement
	if capStat != nil {
		formatCapStat, err := capabilityparser.NewCapabilityStatement(capStat)
		if err != nil {
			return ""
		}
		if formatCapStat != nil {
			fhirVersion, err := formatCapStat.GetFHIRVersion()
			if err != nil {
				return ""
			}
			return fhirVersion
		}
	}
	return ""
}

// Format the SMART Response into JSON
func getSMARTResponse(smartRsp []byte) map[string]interface{} {
	var defaultInt map[string]interface{}
	var smartInt map[string]interface{}
	if smartRsp != nil && len(smartRsp) > 0 {
		err := json.Unmarshal(smartRsp, &smartInt)
		if err != nil {
			return defaultInt
		}
		return smartInt
	}
	return defaultInt
}

// Format the Operation Resource Object into the Supported Resources format
func getSupportedResources(opRes []byte) []string {
	var defaultInt []string
	var opResInt []endpointmanager.OperationAndResource
	checkResource := make(map[string]bool)
	if opRes != nil && len(opRes) > 0 {
		err := json.Unmarshal(opRes, &opResInt)
		if err != nil {
			return defaultInt
		}
		// convert operation and resource object to supported resources format
		// (list of every resource)
		for _, obj := range opResInt {
			if _, ok := checkResource[obj.Resource]; !ok {
				checkResource[obj.Resource] = true
				defaultInt = append(defaultInt, obj.Resource)
			}
		}
	}
	return defaultInt
}

// creates jobs for the workers so that each worker gets the history data
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
		workerDur := viper.GetInt("export_duration")
		// If duration not set, default to 120 seconds
		if workerDur == 0 {
			workerDur = 120
		}

		job := workers.Job{
			Context:     ctx,
			Duration:    time.Duration(workerDur) * time.Second,
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
	var resultRows []Operation

	ha, ok := (*args)["historyArgs"].(historyArgs)
	if !ok {
		return fmt.Errorf("unable to cast arguments to type historyArgs")
	}

	// Get everything from the fhir_endpoints_info_history table for the given URL
	selectHistory := `
		SELECT fhir_endpoints_info_history.url, fhir_endpoints_metadata.http_response, fhir_endpoints_metadata.response_time_seconds, fhir_endpoints_metadata.errors,
		capability_statement, tls_version, mime_types, operation_resource,
		fhir_endpoints_metadata.smart_http_response, smart_response, fhir_endpoints_info_history.updated_at
		FROM fhir_endpoints_info_history, fhir_endpoints_metadata
		WHERE fhir_endpoints_info_history.metadata_id = fhir_endpoints_metadata.id AND fhir_endpoints_info_history.url=$1;`
	historyRows, err := ha.store.DB.QueryContext(ctx, selectHistory, ha.fhirURL)
	if err != nil {
		log.Warnf("Failed getting the history rows for URL %s. Error: %s", ha.fhirURL, err)
		result := Result{
			URL:  ha.fhirURL,
			Rows: resultRows,
		}
		ha.result <- result
		return nil
	}

	// Puts the rows in an array and sends it back on the channel to be processed
	defer historyRows.Close()
	for historyRows.Next() {
		var op Operation
		var url string
		var capStat []byte
		var smartRsp []byte
		var opRes []byte
		err = historyRows.Scan(
			&url,
			&op.HTTPResponse,
			&op.HTTPResponseTimeSecond,
			&op.Errors,
			&capStat,
			&op.TLSVersion,
			pq.Array(&op.MIMETypes),
			&opRes,
			&op.SMARTHTTPResponse,
			&smartRsp,
			&op.UpdatedAt)
		if err != nil {
			log.Warnf("Error while scanning the rows of the history table for URL %s. Error: %s", ha.fhirURL, err)
			result := Result{
				URL:  ha.fhirURL,
				Rows: resultRows,
			}
			ha.result <- result
			return nil
		}

		op.FHIRVersion = getFHIRVersion(capStat)
		op.SMARTResponse = getSMARTResponse(smartRsp)
		op.SupportedResources = getSupportedResources(opRes)

		resultRows = append(resultRows, op)
	}
	result := Result{
		URL:  ha.fhirURL,
		Rows: resultRows,
	}
	ha.result <- result
	return nil
}
