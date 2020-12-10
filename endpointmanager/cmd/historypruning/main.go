package main

import (
	"context"
	"encoding/json"
	"strconv"

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

	threshold := strconv.Itoa(viper.GetInt("pruning_threshold"))
	rows, err := store.DB.Query("SELECT url, capability_statement FROM fhir_endpoints_info_history WHERE operation='U' AND (date_trunc('minute', entered_at) < date_trunc('minute', current_date - interval '" + threshold + "' minute));")

	for rows.Next() {
		var fhirURL string
		var capStatJSON []byte
		err = rows.Scan(&fhirURL, &capStatJSON)
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

		capabilityhandler.HistoryPruningCheck(ctx, store, fhirEndpoint)

	}
}
