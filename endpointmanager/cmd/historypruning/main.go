package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
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

	thresholdInt := viper.GetInt("pruning_threshold")
	threshold := strconv.Itoa(thresholdInt)
	// queryInterval := strconv.Itoa(thresholdInt + (2 * viper.GetInt("capquery_qryintvl")))
	queryInterval := ""

	historyPruningCheck(ctx, store, threshold, queryInterval)
}

func historyPruningCheck(ctx context.Context, store *postgresql.Store, threshold string, queryInterval string) {

	var rows *sql.Rows
	var err error

	if len(queryInterval) != 0 {
		rows, err = store.DB.Query("SELECT operation, url, capability_statement, entered_at FROM fhir_endpoints_info_history WHERE (operation='U' OR operation='I') AND ((date_trunc('minute', entered_at) <= date_trunc('minute', current_date - interval '" + threshold + "' minute)) AND (date_trunc('minute', entered_at) >= date_trunc('minute', current_date - interval '" + queryInterval + "' minute))) ORDER BY url, entered_at DESC;")
		helpers.FailOnError("", err)
	} else {
		rows, err = store.DB.Query("SELECT operation, url, capability_statement, entered_at FROM fhir_endpoints_info_history WHERE (operation='U' OR operation='I') AND (date_trunc('minute', entered_at) <= date_trunc('minute', current_date - interval '" + threshold + "' minute)) ORDER BY url, entered_at DESC;")
		helpers.FailOnError("", err)
	}

	if !rows.Next() {
		return
	}

	_, fhirURL1, entryDate1, capStat1 := getRowInfo(rows)

	for rows.Next() {

		operation2, fhirURL2, entryDate2, capStat2 := getRowInfo(rows)

		// If capstat is not null check if current entry that was passed in has capstat equal to capstat of old entry being checked from history table, otherwise check they are both null
		var equal bool
		if capStat1 != nil {
			equal = capStat1.EqualIgnore(capStat2)
		} else {
			equal = (capStat2 == nil)
		}

		if equal {
			if operation2 == "I" {
				_, err := store.DB.Exec("DELETE FROM fhir_endpoints_info_history WHERE url=$1 AND operation='U' AND entered_at = $2;", fhirURL1, entryDate1)
				helpers.FailOnError("", err)
				if !rows.Next() {
					return
				}
				_, fhirURL1, entryDate1, capStat1 = getRowInfo(rows)
			} else {
				_, err := store.DB.Exec("DELETE FROM fhir_endpoints_info_history WHERE url=$1 AND operation='U' AND entered_at = $2;", fhirURL1, entryDate2)
				helpers.FailOnError("", err)
			}
		} else {
			fhirURL1 = fhirURL2
			entryDate1 = entryDate2
			capStat1 = capStat2
			continue
		}
	}
}

func getRowInfo(rows *sql.Rows) (string, string, string, capabilityparser.CapabilityStatement) {
	var capInt map[string]interface{}
	var operation string
	var fhirURL string
	var capStatJSON []byte
	var entryDate string

	err := rows.Scan(&operation, &fhirURL, &capStatJSON, &entryDate)
	helpers.FailOnError("", err)

	err = json.Unmarshal(capStatJSON, &capInt)
	helpers.FailOnError("", err)
	capStat, err := capabilityparser.NewCapabilityStatementFromInterface(capInt)
	helpers.FailOnError("", err)
	return operation, fhirURL, entryDate, capStat
}
