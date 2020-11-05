package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type jsonEntry struct {
	URL               string      `json:"url"`
	OrganizationNames []string    `json:"api_information_source_name"`
	CreatedAt         time.Time   `json:"created_at"`
	ListSource        string      `json:"list_source"`
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

func main() {
	err := config.SetupConfig()
	helpers.FailOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	ctx := context.Background()
	log.Info("Successfully connected to DB!")

	// Get everything from the fhir_endpoints_info table
	sqlQuery := "SELECT url, endpoint_names, info_created, list_source, vendor_name FROM endpoint_export;"
	rows, err := store.DB.QueryContext(ctx, sqlQuery)
	helpers.FailOnError("Error querying endpoint_export", err)

	var vendorNameNullable sql.NullString

	// Put into an object
	var entries []*jsonEntry
	defer rows.Close()
	for rows.Next() {
		var entry jsonEntry
		err = rows.Scan(
			&entry.URL,
			pq.Array(&entry.OrganizationNames),
			&entry.CreatedAt,
			&entry.ListSource,
			&vendorNameNullable)
		helpers.FailOnError("Error saving endpoint_export data", err)

		if !vendorNameNullable.Valid {
			entry.VendorName = ""
		}

		entries = append(entries, &entry)
	}

	// Get everything from the fhir_endpoints_info_history table
	ctx = context.Background()
	selectHistory := `
		SELECT url, http_response, response_time_seconds, errors,
		capability_statement, tls_version, mime_types, supported_resources,
		smart_http_response, smart_response, updated_at
		FROM fhir_endpoints_info_history;`
	historyRows, err := store.DB.QueryContext(ctx, selectHistory)
	helpers.FailOnError("Error querying endpoint_export", err)
	log.Info("Successfully got everything from fhir_endpoints_info_history table")

	// Put it all into that object
	mapURLHistory := make(map[string][]Operation)
	defer historyRows.Close()
	for historyRows.Next() {
		var op Operation
		var url string
		var capStat []byte
		var smartRsp []byte
		err = historyRows.Scan(
			&url,
			&op.HTTPResponse,
			&op.HTTPResponseTimeSecond,
			&op.Errors,
			&capStat,
			&op.TLSVersion,
			pq.Array(&op.MIMETypes),
			pq.Array(&op.SupportedResources),
			&op.SMARTHTTPResponse,
			&smartRsp,
			&op.UpdatedAt)
		helpers.FailOnError("Error saving fhir_endpoints_info_history data", err)

		// Get fhirVersion
		if capStat != nil {
			formatCapStat, err := capabilityparser.NewCapabilityStatement(capStat)
			helpers.FailOnError("Error converting cap stat to CapabilityStatement", err)
			if formatCapStat != nil {
				fhirVersion, err := formatCapStat.GetFHIRVersion()
				helpers.FailOnError("Error getting FHIR Version", err)
				op.FHIRVersion = fhirVersion
			}
		}

		if smartRsp != nil {
			var smartInt map[string]interface{}
			if len(smartRsp) > 0 {
				err = json.Unmarshal(smartRsp, &smartInt)
				helpers.FailOnError("Error converting smart resp to map[string]interface{}", err)
				op.SMARTResponse = smartInt
			}
		}

		if val, ok := mapURLHistory[url]; ok {
			mapURLHistory[url] = append(val, op)
		} else {
			mapURLHistory[url] = []Operation{op}
		}
	}

	// Put the map into the array
	for i, v := range entries {
		url := v.URL
		if val, ok := mapURLHistory[url]; ok {
			entries[i].Operation = val
		}
	}

	// Convert to JSON
	// finalJSON, err := json.Marshal(entries[0])
	// helpers.FailOnError("Error converting interface to JSON", err)

	// @TODO Figure out how to write it to a file?
	finalFormatJSON, err := json.MarshalIndent(entries, "", "\t")
	helpers.FailOnError("Error converting interface to formatted JSON", err)
	err = ioutil.WriteFile("../../../shinydashboard/lantern/fhir_endpoints_fields.json", finalFormatJSON, 0644)
	helpers.FailOnError("Writing to file failed", err)

}
