package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/archivefile"
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

	entries, err := archivefile.CreateArchive(ctx, store, dateStart, dateEnd)
	helpers.FailOnError("", err)

	// Format as JSON
	finalFormatJSON, err := json.MarshalIndent(entries, "", "\t")
	helpers.FailOnError("", err)
	fmt.Printf("JSON: %s", string(finalFormatJSON))
}
