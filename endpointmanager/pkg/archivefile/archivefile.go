package archivefile

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/workers"
	log "github.com/sirupsen/logrus"
)

// totalSummary is the format of a given URL's JSON object for the archive file
type totalSummary struct {
	URL                string                 `json:"url"`
	CreatedAt          time.Time              `json:"created_at"`
	ListSource         []string               `json:"list_source"`
	OrganizationNames  []string               `json:"api_information_source_name"`
	Updated            map[string]interface{} `json:"updated_at"`
	NumberOfUpdates    int                    `json:"number_of_updates"`
	Operation          map[string]interface{} `json:"operation"`
	FHIRVersion        map[string]interface{} `json:"fhir_version"`
	TLSVersion         map[string]interface{} `json:"tls_version"`
	MIMETypes          firstLastStrArr        `json:"mime_types"`
	Vendor             map[string]interface{} `json:"certified_api_developer_name"`
	ResponseTimeSecond interface{}            `json:"median_response_time"`
	HTTPResponse       []httpResponse         `json:"http_response"`
	SmartHTTPResponse  []smartHTTPResponse    `json:"smart_http_response"`
	Errors             []responseErrors       `json:"errors"`
}

// formats for specific fields in the above totalSummary struct
type firstLastStrArr struct {
	First []string `json:"first"`
	Last  []string `json:"last"`
}
type httpResponse struct {
	ResponseCode  int `json:"http_response_code"`
	ResponseCount int `json:"http_response_count"`
}
type smartHTTPResponse struct {
	ResponseCode  int `json:"smart_http_response_code"`
	ResponseCount int `json:"smart_http_response_count"`
}
type responseErrors struct {
	Error      string `json:"error"`
	ErrorCount int    `json:"error_count"`
}

// Result is the value that is returned from getting the history data from the
// given URL
type Result struct {
	URL     string
	Summary totalSummary
}

// historyArgs is the format for the data passed to getHistory from a worker
type historyArgs struct {
	fhirURL   string
	dateStart string
	dateEnd   string
	store     *postgresql.Store
	result    chan Result
}

// historyEntry is the format of the data received from the history table for the given URL
type historyEntry struct {
	URL              string
	UpdatedAt        time.Time
	Operation        string
	FHIRVersion      string
	FHIRVersionError error
	TLSVersion       string
	MIMETypes        []string
}

// vendorEntry is the format of the data received from the vendor table for the given URL
type vendorEntry struct {
	URL        string
	VendorName string
}

// metadataEntry is the format of the data received from the fhir_endpoints_metadata for the
// given URL
type metadataEntry struct {
	URL                 string
	ResponseTimeSeconds float64
	HTTPResponse        int
	SMARTHTTPResponse   int
	Errors              string
}

// CreateArchive gets all data from fhir_endpoints, fhir_endpoints_info and vendors between
// the given start and end date and summarizes the data
func CreateArchive(ctx context.Context,
	store *postgresql.Store,
	dateStart string,
	dateEnd string,
	numWorkers int,
	workerDur int) ([]totalSummary, error) {
	// Get the fhir_endpoints specific information
	sqlQuery := "SELECT DISTINCT url, organization_names, created_at, list_source FROM fhir_endpoints;"
	rows, err := store.DB.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("ERROR getting data from fhir_endpoints: %s", err)
	}

	var urls []string
	allData := make(map[string]totalSummary)
	defer rows.Close()
	for rows.Next() {
		var entry totalSummary
		var listSource string
		err = rows.Scan(
			&entry.URL,
			pq.Array(&entry.OrganizationNames),
			&entry.CreatedAt,
			&listSource)
		if err != nil {
			return nil, fmt.Errorf("ERROR getting row from fhir_endpoints: %s", err)
		}

		// If the URL already exists, include the new list source and organization names
		if val, ok := allData[entry.URL]; ok {
			val.ListSource = append(val.ListSource, listSource)
			val.OrganizationNames = append(val.OrganizationNames, entry.OrganizationNames...)
			allData[entry.URL] = val
		} else {
			entry.ListSource = []string{listSource}
			allData[entry.URL] = entry
			urls = append(urls, entry.URL)
		}
	}

	// Start workers
	errs := make(chan error)
	allWorkers := workers.NewWorkers()
	err = allWorkers.Start(ctx, numWorkers, errs)
	if err != nil {
		return nil, fmt.Errorf("Error from starting workers. Error: %s", err)
	}

	// Get history data using workers
	resultCh := make(chan Result)
	go createJobs(ctx, resultCh, urls, dateStart, dateEnd, "history", workerDur, store, allWorkers)

	// Add the results from createJobs to allData
	count := 0
	for res := range resultCh {
		u, ok := allData[res.URL]
		if !ok {
			return nil, fmt.Errorf("The URL %s does not exist in the fhir_endpoints tables", res.URL)
		}
		u.NumberOfUpdates = res.Summary.NumberOfUpdates
		u.Updated = res.Summary.Updated
		u.Operation = res.Summary.Operation
		u.FHIRVersion = res.Summary.FHIRVersion
		u.TLSVersion = res.Summary.TLSVersion
		u.MIMETypes = res.Summary.MIMETypes
		allData[res.URL] = u
		if count >= len(urls)-1 {
			close(resultCh)
		}
		count++
	}

	// Get vendor information separately so the endpoints that don't have vendor information aren't
	// removed from the other history request
	vendorQuery := `SELECT f.url, v.name FROM fhir_endpoints_info_history f, vendors v
		WHERE f.updated_at between '` + dateStart + `' AND '` + dateEnd + `' AND f.vendor_id = v.id ORDER BY f.updated_at`
	vendorRows, err := store.DB.QueryContext(ctx, vendorQuery)
	if err != nil {
		return nil, fmt.Errorf("ERROR getting data from fhir_endpoints_info_history and vendors: %s", err)
	}

	vendorResults := make(map[string][]vendorEntry)
	defer vendorRows.Close()
	for vendorRows.Next() {
		var v vendorEntry
		err = vendorRows.Scan(
			&v.URL,
			&v.VendorName)
		if err != nil {
			return nil, fmt.Errorf("Error while scanning the rows of the history and vendor table. Error: %s", err)
		}

		if val, ok := vendorResults[v.URL]; ok {
			vendorResults[v.URL] = append(val, v)
		} else {
			vendorResults[v.URL] = []vendorEntry{v}
		}
	}

	for _, url := range urls {
		u, ok := allData[url]
		if !ok {
			return nil, fmt.Errorf("The URL %s does not exist in the fhir_endpoints tables", url)
		}
		u.Vendor = makeDefaultMap()
		if vResult, ok := vendorResults[url]; ok {
			startElem := vResult[0]
			endElem := vResult[len(vResult)-1]

			u.Vendor["first"] = startElem.VendorName
			if startElem.VendorName != endElem.VendorName {
				u.Vendor["last"] = endElem.VendorName
			}
		}
		allData[url] = u
	}

	// Get history data using workers
	metaResultCh := make(chan Result)
	go createJobs(ctx, metaResultCh, urls, dateStart, dateEnd, "metadata", workerDur, store, allWorkers)

	// Add the results from metadata to allData
	count = 0
	for res := range metaResultCh {
		u, ok := allData[res.URL]
		if !ok {
			return nil, fmt.Errorf("The URL %s does not exist in the fhir_endpoints tables", res.URL)
		}
		u.ResponseTimeSecond = res.Summary.ResponseTimeSecond
		u.HTTPResponse = res.Summary.HTTPResponse
		u.SmartHTTPResponse = res.Summary.SmartHTTPResponse
		u.Errors = res.Summary.Errors
		allData[res.URL] = u
		if count == len(urls)-1 {
			close(metaResultCh)
		}
		count++
	}

	var entries []totalSummary
	for _, e := range allData {
		entries = append(entries, e)
	}

	return entries, nil
}

// Creates a default first & last JSON object, using map[string]interface{} so that an
// empty field is "null" instead of defining it with strings or another type where the default
// would be "" or 0, etc.
func makeDefaultMap() map[string]interface{} {
	defaultMap := map[string]interface{}{
		"first": nil,
		"last":  nil,
	}
	return defaultMap
}

// creates jobs for the workers so that each worker gets the history data
// for a specified url
func createJobs(ctx context.Context,
	ch chan Result,
	urls []string,
	dateStart string,
	dateEnd string,
	jobType string,
	workerDur int,
	store *postgresql.Store,
	allWorkers *workers.Workers) {
	for index := range urls {
		jobArgs := make(map[string]interface{})
		jobArgs["historyArgs"] = historyArgs{
			fhirURL:   urls[index],
			dateStart: dateStart,
			dateEnd:   dateEnd,
			store:     store,
			result:    ch,
		}

		job := workers.Job{
			Context:     ctx,
			Duration:    time.Duration(workerDur) * time.Second,
			HandlerArgs: &jobArgs,
		}

		if jobType == "history" {
			job.Handler = getHistory
		} else {
			job.Handler = getMetadata
		}

		err := allWorkers.Add(&job)
		if err != nil {
			log.Warnf("Error while adding job for getting history for URL %s, %s", urls[index], err)
		}
	}
}

// getHistory retrieves the data from the history table for a specific URL and formats it
// as a totalSummary object before sending it back over the channel
func getHistory(ctx context.Context, args *map[string]interface{}) error {
	returnResult := totalSummary{
		NumberOfUpdates: 0,
		Updated:         makeDefaultMap(),
		Operation:       makeDefaultMap(),
		FHIRVersion:     makeDefaultMap(),
		TLSVersion:      makeDefaultMap(),
	}
	var history []historyEntry

	ha, ok := (*args)["historyArgs"].(historyArgs)
	if !ok {
		return fmt.Errorf("unable to cast arguments to type historyArgs")
	}

	// Get all rows in the history table between given dates
	historyQuery := `SELECT url, updated_at, operation, capability_fhir_version, tls_version, mime_types FROM fhir_endpoints_info_history
		WHERE updated_at between '` + ha.dateStart + `' AND '` + ha.dateEnd + `' AND url=$1 ORDER BY updated_at`
	historyRows, err := ha.store.DB.QueryContext(ctx, historyQuery, ha.fhirURL)
	if err != nil {
		log.Warnf("Failed getting the history rows for URL %s. Error: %s", ha.fhirURL, err)
		result := Result{
			URL:     ha.fhirURL,
			Summary: returnResult,
		}
		ha.result <- result
		return nil
	}

	defer historyRows.Close()
	for historyRows.Next() {
		var e historyEntry
		var fhirVersion string
		var err = historyRows.Scan(
			&e.URL,
			&e.UpdatedAt,
			&e.Operation,
			&fhirVersion,
			&e.TLSVersion,
			pq.Array(&e.MIMETypes))
		if err != nil {
			log.Warnf("Error while scanning the rows of the history table for URL %s. Error: %s", ha.fhirURL, err)
			result := Result{
				URL:     ha.fhirURL,
				Summary: returnResult,
			}
			ha.result <- result
			return nil
		}

		if fhirVersion == "" {
			e.FHIRVersion = fhirVersion
			e.FHIRVersionError = fmt.Errorf("received NULL FHIR version")
		} else {
			e.FHIRVersion = fhirVersion
			e.FHIRVersionError = nil
		}

		history = append(history, e)
	}

	if len(history) > 0 {
		returnResult.NumberOfUpdates = len(history)
		startElem := history[0]
		endElem := history[len(history)-1]

		returnResult.Updated["first"] = startElem.UpdatedAt
		if startElem.UpdatedAt != endElem.UpdatedAt {
			returnResult.Updated["last"] = endElem.UpdatedAt
		}
		returnResult.Operation["first"] = startElem.Operation
		if startElem.Operation != endElem.Operation {
			returnResult.Operation["last"] = endElem.Operation
		}
		if startElem.FHIRVersionError == nil {
			returnResult.FHIRVersion["first"] = startElem.FHIRVersion
		}
		if (startElem.FHIRVersion != endElem.FHIRVersion) && (endElem.FHIRVersionError == nil) {
			returnResult.FHIRVersion["last"] = endElem.FHIRVersion
		}
		returnResult.TLSVersion["first"] = startElem.TLSVersion
		if startElem.TLSVersion != endElem.TLSVersion {
			returnResult.TLSVersion["last"] = endElem.TLSVersion
		}
		returnResult.MIMETypes.First = startElem.MIMETypes
		if !helpers.StringArraysEqual(startElem.MIMETypes, endElem.MIMETypes) {
			returnResult.MIMETypes.Last = endElem.MIMETypes
		}
	}

	result := Result{
		URL:     ha.fhirURL,
		Summary: returnResult,
	}
	ha.result <- result
	return nil
}

// getMetadata retrieves the data from the metadata table for a specific URL and formats it
// as a totalSummary object before sending it back over the channel
func getMetadata(ctx context.Context, args *map[string]interface{}) error {
	var returnResult totalSummary
	var history []metadataEntry

	ha, ok := (*args)["historyArgs"].(historyArgs)
	if !ok {
		return fmt.Errorf("unable to cast arguments to type historyArgs")
	}

	// Get all rows in the history table between given dates
	metadataQuery := `SELECT url, response_time_seconds, http_response, smart_http_response, errors FROM fhir_endpoints_metadata
		WHERE updated_at between '` + ha.dateStart + `' AND '` + ha.dateEnd + `' AND url=$1 ORDER BY updated_at`
	metadataRows, err := ha.store.DB.QueryContext(ctx, metadataQuery, ha.fhirURL)
	if err != nil {
		log.Warnf("Failed getting the metadata rows for URL %s. Error: %s", ha.fhirURL, err)
		result := Result{
			URL:     ha.fhirURL,
			Summary: returnResult,
		}
		ha.result <- result
		return nil
	}

	defer metadataRows.Close()
	for metadataRows.Next() {
		var e metadataEntry
		err = metadataRows.Scan(
			&e.URL,
			&e.ResponseTimeSeconds,
			&e.HTTPResponse,
			&e.SMARTHTTPResponse,
			&e.Errors)
		if err != nil {
			log.Warnf("Error while scanning the rows of the metadata table for URL %s. Error: %s", ha.fhirURL, err)
			result := Result{
				URL:     ha.fhirURL,
				Summary: returnResult,
			}
			ha.result <- result
			return nil
		}

		history = append(history, e)
	}

	if len(history) > 0 {
		var respTime []float64
		httpResponseMap := make(map[int]int)
		smartHTTPRespMap := make(map[int]int)
		errorsMap := make(map[string]int)
		// Keep track of each unique http response, smart http response, and error value
		// and how many of each unique value there is
		for _, elem := range history {
			respTime = append(respTime, elem.ResponseTimeSeconds)
			if val, ok := httpResponseMap[elem.HTTPResponse]; ok {
				httpResponseMap[elem.HTTPResponse] = val + 1
			} else {
				httpResponseMap[elem.HTTPResponse] = 1
			}
			if val, ok := smartHTTPRespMap[elem.SMARTHTTPResponse]; ok {
				smartHTTPRespMap[elem.SMARTHTTPResponse] = val + 1
			} else {
				smartHTTPRespMap[elem.SMARTHTTPResponse] = 1
			}
			if val, ok := errorsMap[elem.Errors]; ok {
				errorsMap[elem.Errors] = val + 1
			} else {
				errorsMap[elem.Errors] = 1
			}
		}
		// Calculate median of given response times
		sort.Slice(respTime, func(i, j int) bool {
			return respTime[i] < respTime[j]
		})
		var median float64
		if len(respTime)%2 == 0 {
			idx := (len(respTime) - 1) / 2
			floatMedian := (respTime[idx] + respTime[idx+1]) / 2
			// Round to 4 decimal places
			median = math.Round(floatMedian*10000) / 10000
		} else {
			idx := len(respTime) / 2
			median = respTime[idx]
		}

		// Loop through each map and for each element create an array
		// item for it's respective totalSummary field
		var httpRespArr []httpResponse
		for resp, total := range httpResponseMap {
			httpResp := httpResponse{
				ResponseCode:  resp,
				ResponseCount: total,
			}
			httpRespArr = append(httpRespArr, httpResp)
		}
		var smartHTTPRespArr []smartHTTPResponse
		for resp, total := range smartHTTPRespMap {
			smartResp := smartHTTPResponse{
				ResponseCode:  resp,
				ResponseCount: total,
			}
			smartHTTPRespArr = append(smartHTTPRespArr, smartResp)
		}
		var errorArray []responseErrors
		for resp, total := range errorsMap {
			errorResp := responseErrors{
				Error:      resp,
				ErrorCount: total,
			}
			errorArray = append(errorArray, errorResp)
		}
		returnResult.ResponseTimeSecond = median
		returnResult.HTTPResponse = httpRespArr
		returnResult.SmartHTTPResponse = smartHTTPRespArr
		returnResult.Errors = errorArray
	}

	result := Result{
		URL:     ha.fhirURL,
		Summary: returnResult,
	}
	ha.result <- result
	return nil
}
