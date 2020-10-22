package main

import (
	"context"
	"fmt"
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
	HTTPResponse           int                            `json:"http_response"`
	HTTPResponseTimeSecond float64                        `json:"http_response_time_second"`
	Errors                 string                         `json:"errors"`
	FHIRVersion            string                         `json:"fhir_version"`
	TLSVersion             string                         `json:"tls_verison"`
	MIMETypes              []string                       `json:"mime_types"`
	SupportedResources     []string                       `json:"supported_resources"`
	SMARTHTTPResponse      int                            `json:"smart_http_response"`
	SMARTResponse          capabilityparser.SMARTResponse `json:"smart_response"`
	UpdatedAt              time.Time                      `json:"updated"`
}

func main() {
	err := config.SetupConfig()
	helpers.FailOnError("", err)

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)
	ctx := context.Background()
	log.Info("Successfully connected to DB!")

	// Copy entire contents of endpoint_export view into a csv which will be written to /tmp
	sqlQuery := "SELECT url, organization_names, created_at, list_source, vendor_name FROM endpoint_export;"
	rows, err := store.DB.QueryContext(ctx, sqlQuery)
	helpers.FailOnError("Error querying endpoint_export", err)

	var entries []*jsonEntry
	defer rows.Close()
	for rows.Next() {
		var entry jsonEntry
		err = rows.Scan(
			&entry.URL,
			pq.Array(&entry.OrganizationNames),
			&entry.CreatedAt,
			&entry.ListSource,
			&entry.VendorName)
		helpers.FailOnError("Error saving endpoint_export data", err)
		entries = append(entries, &entry)
	}

	fmt.Printf("ENTRIES IN THE DATABASE: %+v", entries)

	// Get everything from the fhir_endpoints_info table
	// Put into an object
	// Get everything from the fhir_endpoints_info_history table
	// Put it all into that object

	// @TODO Figure out how to get fhirversion
}
