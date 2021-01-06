package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lib/pq"
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
type fhirEndpointsEntry struct {
	URL               string    `json:"url"`
	CreatedAt         time.Time `json:"created_at"`
	ListSource        []string  `json:"list_source"`
	OrganizationNames []string  `json:"api_information_source_name"`
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
	sqlQuery := "SELECT DISTINCT url, organization_names, created_at, list_source from fhir_endpoints where created_at between '" + dateStart + "' AND '" + dateEnd + "';"
	rows, err := store.DB.QueryContext(ctx, sqlQuery)
	if err != nil {
		// return nil, fmt.Errorf("Make sure that the database is not empty. Error: %s", err)
		log.Fatalf("ERROR getting data from fhir_endpoints: %s", err)
	}

	var urls []string
	entryCheck := make(map[string]fhirEndpointsEntry)
	defer rows.Close()
	for rows.Next() {
		var entry fhirEndpointsEntry
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

	var entries []fhirEndpointsEntry
	for _, e := range entryCheck {
		entries = append(entries, e)
	}

	// Format as JSON
	finalFormatJSON, err := json.MarshalIndent(entries, "", "\t")
	fmt.Printf("JSON: %s", string(finalFormatJSON))

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
	*/

	// 1. For each URL in the fhir_endpoints_info_history table, we want to get the first row
	// that is after the start date and the last row that is before the end date
	// 2. Want to compare those two values

	// dateTimeStart := dateStart + "T00:00:00.001Z"
	// formatStart, err := time.Parse(time.RFC3339, dateTimeStart)

	// if err != nil {
	// 	log.Fatalf("ERROR: Start date not in correct format, %s", err)
	// }
	// fmt.Println(formatStart)

	// dateTimeEnd := dateEnd + "T00:00:00.001Z"
	// formatEnd, err := time.Parse(time.RFC3339, dateTimeEnd)

	// if err != nil {
	// 	log.Fatalf("ERROR: End date not in correct format, %s", err)
	// }
	// fmt.Println(formatEnd)

	// startSec := formatStart.Unix()
	// fmt.Println(startSec)

	// endSec := formatEnd.Unix()
	// fmt.Println(endSec)

	// ctx := context.Background()
	// "SELECT date.datetime AS time, date.url AS fhirURL,
	//                 FROM (SELECT extract(epoch from fhir_endpoints_info_history.entered_at) AS datetime, url as fhirURL FROM fhir_endpoints_info_history) as date,
	//                 WHERE date.datetime between " + startSec + "AND " + endSec
	// sqlQuery := "SELECT DISTINCT url, endpoint_names, info_created, list_source, vendor_name FROM endpoint_export;"
	// rows, err := store.DB.QueryContext(ctx, sqlQuery)
	// if err != nil {
	// 	return nil, fmt.Errorf("Make sure that the database is not empty. Error: %s", err)
	// }

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

}
