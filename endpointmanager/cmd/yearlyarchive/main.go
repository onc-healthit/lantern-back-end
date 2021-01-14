package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

/**
Fields from fhir_endpoints:
"url":"",
"created_at":"",
"list_source":[],
"api_information_source_name":[],
*/
type totalSummary struct {
	URL               string    `json:"url"`
	CreatedAt         time.Time `json:"created_at"`
	ListSource        []string  `json:"list_source"`
	OrganizationNames []string  `json:"api_information_source_name"`
	Updated           map[string]interface{}
	NumberOfUpdates   int
	Operation         map[string]interface{}
	FHIRVersion       map[string]interface{}
	TLSVersion        map[string]interface{}
	MIMETypes         firstLastStrArr
	Vendor            map[string]interface{} `json:"certified_api_developer_name""`
}

// @TODO Remove json
type historyEntry struct {
	URL              string    `json:"url"`
	UpdatedAt        time.Time `json:"updated_at"`
	Operation        string    `json:"operation"`
	FHIRVersion      string    `json:"fhir_version"`
	FHIRVersionError error     `json:"fhir_version_error"`
	TLSVersion       string    `json:"tls_version"`
	MIMETypes        []string  `json:"mime_types`
}

type vendorEntry struct {
	VendorID   int
	UpdatedAt  time.Time
	URL        string
	ID         int
	VendorName string
}

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

type firstLastStrArr struct {
	First []string `json:"first"`
	Last  []string `json:"last"`
}

// type historySummary struct {
// 	Updated         map[string]interface{}
// 	NumberOfUpdates int
// 	Operation       map[string]interface{}
// 	FHIRVersion     map[string]interface{}
// 	TLSVersion      map[string]interface{}
// 	MIMETypes       firstLastStrArr
// }

// type vendorSummary struct {
// 	Vendor map[string]interface{} `json:"certified_api_developer_name""`
// }

var defaultMapInterface = map[string]interface{}{
	"first": nil,
	"last":  nil,
}

// @TODO Get rid of all print statements
func main() {
	var dateStart string
	var dateEnd string

	if len(os.Args) >= 3 {
		dateStart = os.Args[1]
		dateEnd = os.Args[2]
	} else {
		log.Fatalf("ERROR: Missing date-range command-line arguments")
	}

	err := config.SetupConfig()
	helpers.FailOnError("", err)

	layout := "2006-01-02"
	formatStart, err := time.Parse(layout, dateStart)

	if err != nil {
		log.Fatalf("ERROR: Start date not in correct format, %s", err)
	}
	fmt.Println(formatStart)

	formatEnd, err := time.Parse(layout, dateEnd)

	if err != nil {
		log.Fatalf("ERROR: End date not in correct format, %s", err)
	}
	fmt.Println(formatEnd)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	log.Info("Successfully connected to DB!")

	ctx := context.Background()

	// Get the fhir_endpoints specific information
	sqlQuery := "SELECT DISTINCT url, organization_names, created_at, list_source from fhir_endpoints;"
	rows, err := store.DB.QueryContext(ctx, sqlQuery)
	if err != nil {
		// return nil, fmt.Errorf("Make sure that the database is not empty. Error: %s", err)
		log.Fatalf("ERROR getting data from fhir_endpoints: %s", err)
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
			log.Fatalf("ERROR getting row from fhir_endpoints: %s", err)
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
		// return nil, fmt.Errorf("Make sure that the database is not empty. Error: %s", err)
		log.Fatalf("ERROR getting data from fhir_endpoints_info_history: %s", err)
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
			// log.Warnf("Error while scanning the rows of the history table. Error: %s", err)
			log.Fatalf("Error while scanning the rows of the history table. Error: %s", err)
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
	for k, u := range allData {
		if history, ok := resultMap[k]; ok {
			u.NumberOfUpdates = len(history)

			startElem := history[0]
			endElem := history[len(history)-1]

			// summary.Updated = make(map[string]interface{})
			// @TODO Make sure to have defaults for each if for some reason an entry doesn't have a row in the history table?
			u.Updated = defaultMapInterface
			u.Updated["first"] = startElem.UpdatedAt
			u.Updated["last"] = nil
			if startElem.UpdatedAt != endElem.UpdatedAt {
				u.Updated["last"] = endElem.UpdatedAt
			}

			u.Operation = make(map[string]interface{})
			u.Operation["first"] = startElem.Operation
			u.Operation["last"] = nil
			if startElem.Operation != endElem.Operation {
				u.Operation["last"] = endElem.Operation
			}

			u.FHIRVersion = make(map[string]interface{})
			u.FHIRVersion["first"] = nil
			if startElem.FHIRVersionError == nil {
				u.FHIRVersion["first"] = startElem.FHIRVersion
			}
			u.FHIRVersion["last"] = nil
			if (startElem.FHIRVersion != endElem.FHIRVersion) && endElem.FHIRVersionError != nil {
				u.FHIRVersion["last"] = endElem.FHIRVersion
			}

			u.TLSVersion = make(map[string]interface{})
			u.TLSVersion["first"] = startElem.TLSVersion
			u.TLSVersion["last"] = nil
			if startElem.TLSVersion != endElem.TLSVersion {
				u.TLSVersion["last"] = endElem.TLSVersion
			}

			u.MIMETypes.First = startElem.MIMETypes
			if !helpers.StringArraysEqual(startElem.MIMETypes, endElem.MIMETypes) {
				u.MIMETypes.Last = endElem.MIMETypes
			}

			allData[k] = u
		} else {
			log.Infof("This url %s does not have an entry in the history table", u)
		}
	}

	/**
	Want to deal with vendor stuff separately
	"certified_api_developer_name":{
		"first":"",
		"last":""
	},
	*/

	// Get vendor information
	vendorQuery := `SELECT f.vendor_id, f.updated_at, f.url, v.id, v.name FROM fhir_endpoints_info_history f, vendors v
		WHERE f.updated_at between '` + dateStart + `' AND '` + dateEnd + `' AND f.vendor_id = v.id ORDER BY f.updated_at`
	vendorRows, err := store.DB.QueryContext(ctx, vendorQuery)
	if err != nil {
		// return nil, fmt.Errorf("Make sure that the database is not empty. Error: %s", err)
		log.Fatalf("ERROR getting data from fhir_endpoints_info_history and vendors: %s", err)
	}

	vendorResults := make(map[string][]vendorEntry)
	defer vendorRows.Close()
	for vendorRows.Next() {
		var v vendorEntry
		var vendorIDNullable sql.NullInt64
		err = vendorRows.Scan(
			&vendorIDNullable,
			&v.UpdatedAt,
			&v.URL,
			&v.ID,
			&v.VendorName)
		if err != nil {
			// log.Warnf("Error while scanning the rows of the history table. Error: %s", err)
			log.Fatalf("Error while scanning the rows of the history and vendor table. Error: %s", err)
		}

		if !vendorIDNullable.Valid {
			v.VendorID = 0
		} else {
			v.VendorID = int(vendorIDNullable.Int64)
		}

		// @TODO Do we have to worry about repeat urls?
		// If the URL already exists, currently just print out something
		if val, ok := vendorResults[v.URL]; ok {
			vendorResults[v.URL] = append(val, v)
		} else {
			vendorResults[v.URL] = []vendorEntry{v}
		}
	}

	for k, u := range allData {
		if vResult, ok := vendorResults[k]; ok {
			startElem := vResult[0]
			endElem := vResult[len(vResult)-1]

			u.Vendor = make(map[string]interface{})
			u.Vendor["first"] = startElem.VendorName
			u.Vendor["last"] = nil
			if startElem.VendorName != endElem.VendorName {
				u.Vendor["last"] = endElem.VendorName
			}
		} else {
			log.Infof("This url %s does not have an entry in the vendor table", u)
		}
	}

	var entries []totalSummary
	for _, e := range allData {
		entries = append(entries, e)
	}

	/**
	{
		"url":"",
		"created_at":"",
		"list_source":"",
		"api_information_source_name":{
			"first":"",
			"last":""
			},
		"updated":{
			"first":"",
			"last":""
			},
		"number_of_updates":""
		"operation":{
			"first":"",
			"last":""
			},
		"certified_api_developer_name":{
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
		"response_time_second":"",
		"http_response":{
			"http_response_code":[""],
			"http_response_count":[""]
			},
		"smart_http_response":{
			"smart_http_response_code":[""],
			"smart_http_response_count":[""]
			},
		"errors":{
			"error":[""],
			"error_count":[""]
			}
		}
	*/

	/**
	Fields from fhir_endpoints:
	"url":"",
	"created_at":"",
	"list_source":"",
	"api_information_source_name":{
		"first":"",
		"last":""
	},
	*/

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

	/** Questions
	1. We don't have the history of the api information source name (organization name) since there is no fhir_endpoints history table, just one for the info
	*/

	// Format as JSON
	finalFormatJSON, err := json.MarshalIndent(entries, "", "\t")
	// finalFormatJSON, err := json.MarshalIndent(resultMap, "", "\t")
	// finalFormatJSON, err := json.MarshalIndent(allHistory, "", "\t")
	// finalFormatJSON, err := json.MarshalIndent(vendorHistory, "", "\t")
	fmt.Printf("JSON: %s", string(finalFormatJSON))

}

// Get the FHIR Version from the capability statement
func getFHIRVersion(capStat []byte) (string, error) {
	// Get the FHIR Version from the capability statement
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
