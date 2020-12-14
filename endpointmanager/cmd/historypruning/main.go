package main

import (
	"context"
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/capabilityreceiver/pkg/capabilityhandler"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"

	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

func main() {
	err := config.SetupConfig()
	helpers.FailOnError("", err)

	ctx := context.Background()
	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	helpers.FailOnError("", err)

	var count int
	ctStatement, err := store.DB.Prepare(`SELECT count(*) FROM fhir_endpoints_info_history WHERE url = $1 AND entered_at = $2;`)
	helpers.FailOnError("", err)

	rows, err := store.DB.Query("SELECT url, capability_statement, entered_at FROM fhir_endpoints_info_history WHERE operation='U' OR operation='I' ORDER BY url, entered_at DESC;")
	helpers.FailOnError("", err)

	for rows.Next() {

		var fhirURL string
		var capStatJSON []byte
		var entryDate string
		err = rows.Scan(&fhirURL, &capStatJSON, &entryDate)
		helpers.FailOnError("", err)

		var capInt map[string]interface{}
		err = json.Unmarshal(capStatJSON, &capInt)
		helpers.FailOnError("", err)
		capStat, err := capabilityparser.NewCapabilityStatementFromInterface(capInt)
		helpers.FailOnError("", err)

		fhirEndpoint := endpointmanager.FHIREndpointInfo{
			URL:                 fhirURL,
			CapabilityStatement: capStat,
		}

		// Check to make sure the entry has not already been deleted, and if not call history pruning function
		err = ctStatement.QueryRow(fhirURL, entryDate).Scan(&count)
		helpers.FailOnError("", err)
		if count != 0 {
			capabilityhandler.HistoryPruningCheck(ctx, store, &fhirEndpoint, entryDate)
		}

	}
}
