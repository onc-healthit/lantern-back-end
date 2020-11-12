package jsonexport

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
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
	sqlQuery := "SELECT url, endpoint_names, info_created, list_source, vendor_name FROM endpoint_export;"
	rows, err := store.DB.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}

	// Put into an object
	var entries []*jsonEntry
	defer rows.Close()
	for rows.Next() {
		var entry jsonEntry
		var vendorNameNullable sql.NullString
		err = rows.Scan(
			&entry.URL,
			pq.Array(&entry.OrganizationNames),
			&entry.CreatedAt,
			&entry.ListSource,
			&vendorNameNullable)
		if err != nil {
			return nil, err
		}

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
	if err != nil {
		return nil, err
	}

	// Group the rows by URL, to create a map from URLs
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
		if err != nil {
			return nil, err
		}

		op.FHIRVersion = getFHIRVersion(capStat)
		op.SMARTResponse = getSMARTResponse(smartRsp)

		if val, ok := mapURLHistory[url]; ok {
			mapURLHistory[url] = append(val, op)
		} else {
			mapURLHistory[url] = []Operation{op}
		}
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
	if smartRsp != nil {
		if len(smartRsp) > 0 {
			err := json.Unmarshal(smartRsp, &smartInt)
			if err != nil {
				return defaultInt
			}
			return smartInt
		}
	}
	return defaultInt
}
