package archivefile

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

/**
Fields from fhir_endpoints:
"url":"",
"created_at":"",
"list_source":[],
"api_information_source_name":[],
*/
/**
"updated":{
	"first":"",
	"last":""
},
"number_of_updates":""
"operation":{
	"first":"",
	"last":""
},
"fhir_version":{
	"first":"",
	"last":""
},
"tls_version":{
	"first":"",
	"last":""
},
"mime_types":{
	"first":"",
	"last":""
},
*/
type totalSummary struct {
	URL               string                 `json:"url"`
	CreatedAt         time.Time              `json:"created_at"`
	ListSource        []string               `json:"list_source"`
	OrganizationNames []string               `json:"api_information_source_name"`
	Updated           map[string]interface{} `json:"updated_at"`
	NumberOfUpdates   int                    `json:"number_of_updates"`
	Operation         map[string]interface{} `json:"operation"`
	FHIRVersion       map[string]interface{} `json:"fhir_version"`
	TLSVersion        map[string]interface{} `json:"tls_version"`
	MIMETypes         firstLastStrArr        `json:"mime_types"`
	Vendor            map[string]interface{} `json:"certified_api_developer_name"`
}

// @TODO Remove json
type historyEntry struct {
	URL              string
	UpdatedAt        time.Time
	Operation        string
	FHIRVersion      string
	FHIRVersionError error
	TLSVersion       string
	MIMETypes        []string
}

type vendorEntry struct {
	URL        string
	VendorName string
}

type firstLastStrArr struct {
	First []string `json:"first"`
	Last  []string `json:"last"`
}

// CreateArchive gets all data from fhir_endpoints, fhir_endpoints_info and vendors between
// the given start and end date and summarizes the data
// @TODO Get rid of all print statements
func CreateArchive(ctx context.Context, store *postgresql.Store, dateStart string, dateEnd string) ([]totalSummary, error) {
	// Get the fhir_endpoints specific information
	sqlQuery := "SELECT DISTINCT url, organization_names, created_at, list_source from fhir_endpoints;"
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

	/**
	Fields from fhir_endpoints_info_history:
	"updated":{
		"first":"",
		"last":""
	},
	"number_of_updates":""
	"operation":{
		"first":"",
		"last":""
	},
	"fhir_version":{
		"first":"",
		"last":""
	},
	"tls_version":{
		"first":"",
		"last":""
	},
	"mime_types":{
		"first":"",
		"last":""
	},
	*/

	// Get all rows in the history table between given dates
	historyQuery := `SELECT url, updated_at, operation, capability_statement, tls_version, mime_types FROM fhir_endpoints_info_history
		WHERE updated_at between '` + dateStart + `' AND '` + dateEnd + `' ORDER BY updated_at`
	historyRows, err := store.DB.QueryContext(ctx, historyQuery)
	if err != nil {
		return nil, fmt.Errorf("ERROR getting data from fhir_endpoints_info_history: %s", err)
	}

	// Have to pull out the data properly
	// Have to handle a null vendor id?
	// @TODO Break this out into workers again?
	resultMap := make(map[string][]historyEntry)
	defer historyRows.Close()
	for historyRows.Next() {
		var e historyEntry
		var capStat []byte
		err = historyRows.Scan(
			&e.URL,
			&e.UpdatedAt,
			&e.Operation,
			&capStat,
			&e.TLSVersion,
			pq.Array(&e.MIMETypes))
		if err != nil {
			return nil, fmt.Errorf("Error while scanning the rows of the history table. Error: %s", err)
		}

		e.FHIRVersion, e.FHIRVersionError = getFHIRVersion(capStat)

		// If the URL already exists, currently just print out something
		if val, ok := resultMap[e.URL]; ok {
			resultMap[e.URL] = append(val, e)
		} else {
			resultMap[e.URL] = []historyEntry{e}
		}
	}

	// @TODO Might want to put this in the above loop later
	// Loop through the url list to get associated history data
	for _, url := range urls {
		u, ok := allData[url]
		if !ok {
			return nil, fmt.Errorf("The URL %s does not exist in the fhir_endpoints tables", url)
		}
		u.NumberOfUpdates = 0
		u.Updated = makeDefaultMap()
		u.Operation = makeDefaultMap()
		u.FHIRVersion = makeDefaultMap()
		u.TLSVersion = makeDefaultMap()
		if history, ok := resultMap[url]; ok {
			u.NumberOfUpdates = len(history)
			startElem := history[0]
			endElem := history[len(history)-1]

			u.Updated["first"] = startElem.UpdatedAt
			if startElem.UpdatedAt != endElem.UpdatedAt {
				u.Updated["last"] = endElem.UpdatedAt
			}

			u.Operation["first"] = startElem.Operation
			if startElem.Operation != endElem.Operation {
				u.Operation["last"] = endElem.Operation
			}

			if startElem.FHIRVersionError == nil {
				u.FHIRVersion["first"] = startElem.FHIRVersion
			}
			if (startElem.FHIRVersion != endElem.FHIRVersion) && endElem.FHIRVersionError != nil {
				u.FHIRVersion["last"] = endElem.FHIRVersion
			}

			u.TLSVersion["first"] = startElem.TLSVersion
			if startElem.TLSVersion != endElem.TLSVersion {
				u.TLSVersion["last"] = endElem.TLSVersion
			}

			u.MIMETypes.First = startElem.MIMETypes
			if !helpers.StringArraysEqual(startElem.MIMETypes, endElem.MIMETypes) {
				u.MIMETypes.Last = endElem.MIMETypes
			}
		} else {
			log.Infof("This url %s does not have an entry in the history table", url)
		}
		allData[url] = u
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
		} else {
			log.Infof("This url %s does not have an entry in the vendor table", url)
		}
		allData[url] = u
	}

	var entries []totalSummary
	for _, e := range allData {
		entries = append(entries, e)
	}

	/**
	Fields from the new metadata table:
	"response_time_second":"",
	"http_response":[
		{
			"http_response_code": 200,
			"http_response_count": 10
		},
		{
			"http_response_code": 404,
			"http_response_count": 10
		}...
	],
	"smart_http_response":[
		{
			"smart_http_response_code": 200,
			"smart_http_response_count": 10
		},
		{
			"smart_http_response_code": 404,
			"smart_http_response_count": 10
		}...
	],
	"errors":[
		{
			"error": "did not return",
			"error_count": 10
		},
		{
			"error": "something went wrong",
			"error_count": 10
		}...
	],
	*/

	return entries, nil
}

// Gets the FHIR Version from the capability statement
func getFHIRVersion(capStat []byte) (string, error) {
	if capStat != nil {
		formatCapStat, err := capabilityparser.NewCapabilityStatement(capStat)
		if err != nil {
			return "", err
		}
		if formatCapStat != nil {
			fhirVersion, err := formatCapStat.GetFHIRVersion()
			if err != nil {
				return "", err
			}
			return fhirVersion, nil
		}
	}
	return "", fmt.Errorf("no capability statement to retreive FHIR Version from")
}

// Creates a default first & last JSON object
func makeDefaultMap() map[string]interface{} {
	defaultMap := map[string]interface{}{
		"first": nil,
		"last":  nil,
	}
	return defaultMap
}
