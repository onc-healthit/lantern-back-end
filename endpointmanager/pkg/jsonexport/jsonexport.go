package jsonexport

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/lib/pq"
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
	fhirURL    string
	store      *postgresql.Store
	result     chan Result
	exportType string
}

// CreateJSONExport formats the data from the fhir_endpoints_info and fhir_endpoints_info_history
// tables into a given specification
func CreateJSONExport(ctx context.Context, store *postgresql.Store, fileToWriteTo string, exportType string) error {
	finalFormatJSON, err := createJSON(ctx, store, exportType)
	if err != nil {
		return err
	}
	// Write to the given file
	err = ioutil.WriteFile(fileToWriteTo, finalFormatJSON, 0644)
	return err
}

func createJSON(ctx context.Context, store *postgresql.Store, exportType string) ([]byte, error) {
	// Get everything from the fhir_endpoints_info table
	sqlQuery := "SELECT DISTINCT url, endpoint_names, info_created, list_source, vendor_name FROM endpoint_export WHERE info_created IS NOT NULL;"
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
	go createJobs(ctx, resultCh, urls, store, allWorkers, exportType)

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

// Format the SMART Response into JSON
func getSMARTResponse(smartRsp []byte) map[string]interface{} {
	var defaultInt map[string]interface{}
	var smartInt map[string]interface{}
	if len(smartRsp) > 0 {
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
	var opResInt map[string][]string
	checkResource := make(map[string]bool)
	if len(opRes) > 0 {
		err := json.Unmarshal(opRes, &opResInt)
		if err != nil {
			return defaultInt
		}
		// convert operation and resource object to supported resources format
		// (list of every resource)
		for _, resArr := range opResInt {
			for _, resource := range resArr {
				if _, ok := checkResource[resource]; !ok {
					checkResource[resource] = true
					defaultInt = append(defaultInt, resource)
				}
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
	allWorkers *workers.Workers,
	exportType string) {
	for index := range urls {
		jobArgs := make(map[string]interface{})
		jobArgs["historyArgs"] = historyArgs{
			fhirURL:    urls[index],
			store:      store,
			result:     ch,
			exportType: exportType,
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

	exportType := ha.exportType

	// Get everything from the fhir_endpoints_info_history table for the given URL

	var selectHistory string
	if exportType == "month" {
		selectHistory = `
		SELECT fhir_endpoints_info_history.url, fhir_endpoints_metadata.http_response, fhir_endpoints_metadata.response_time_seconds, fhir_endpoints_metadata.errors,
		capability_statement, tls_version, mime_types, operation_resource,
		fhir_endpoints_metadata.smart_http_response, smart_response, fhir_endpoints_info_history.updated_at, capability_fhir_version
		FROM fhir_endpoints_info_history, fhir_endpoints_metadata
		WHERE fhir_endpoints_info_history.metadata_id = fhir_endpoints_metadata.id AND fhir_endpoints_info_history.url=$1 AND (date_trunc('month', fhir_endpoints_info_history.updated_at) = date_trunc('month', current_date - INTERVAL '1 month'))
		ORDER BY fhir_endpoints_info_history.updated_at DESC;`
	} else if exportType == "30days" {
		selectHistory = `
		SELECT fhir_endpoints_info_history.url, fhir_endpoints_metadata.http_response, fhir_endpoints_metadata.response_time_seconds, fhir_endpoints_metadata.errors,
		capability_statement, tls_version, mime_types, operation_resource,
		fhir_endpoints_metadata.smart_http_response, smart_response, fhir_endpoints_info_history.updated_at, capability_fhir_version
		FROM fhir_endpoints_info_history, fhir_endpoints_metadata
		WHERE fhir_endpoints_info_history.metadata_id = fhir_endpoints_metadata.id AND fhir_endpoints_info_history.url=$1 AND (date_trunc('day', fhir_endpoints_info_history.updated_at) >= date_trunc('day', current_date - INTERVAL '30 day'))
		ORDER BY fhir_endpoints_info_history.updated_at DESC;`
	} else if exportType == "all" {
		selectHistory = `
		SELECT fhir_endpoints_info_history.url, fhir_endpoints_metadata.http_response, fhir_endpoints_metadata.response_time_seconds, fhir_endpoints_metadata.errors,
		capability_statement, tls_version, mime_types, operation_resource,
		fhir_endpoints_metadata.smart_http_response, smart_response, fhir_endpoints_info_history.updated_at, capability_fhir_version
		FROM fhir_endpoints_info_history, fhir_endpoints_metadata
		WHERE fhir_endpoints_info_history.metadata_id = fhir_endpoints_metadata.id AND fhir_endpoints_info_history.url=$1
		ORDER BY fhir_endpoints_info_history.updated_at DESC;`
	}

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
			&op.UpdatedAt,
			&op.FHIRVersion)
		if err != nil {
			log.Warnf("Error while scanning the rows of the history table for URL %s. Error: %s", ha.fhirURL, err)
			result := Result{
				URL:  ha.fhirURL,
				Rows: resultRows,
			}
			ha.result <- result
			return nil
		}

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
