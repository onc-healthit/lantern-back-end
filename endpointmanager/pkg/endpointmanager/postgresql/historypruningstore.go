package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
)

var pruningStatementQueryInterval *sql.Stmt
var pruningStatementNoQueryInterval *sql.Stmt
var pruningDeleteStatement *sql.Stmt

// PruneInfoHistory checks info table and prunes any repetitive entries
func (s *Store) PruneInfoHistory(ctx context.Context, threshold int, queryInterval int) {

	var rows *sql.Rows
	var err error

	thresholdString := strconv.Itoa(threshold)

	if queryInterval >= 0 {
		queryIntString := strconv.Itoa(threshold + (3 * queryInterval))
		rows, err = pruningStatementQueryInterval.QueryContext(ctx, thresholdString, queryIntString)
		helpers.FailOnError("", err)
	} else {
		rows, err = pruningStatementNoQueryInterval.QueryContext(ctx, thresholdString)
		helpers.FailOnError("", err)
	}

	if !rows.Next() {
		return
	}

	_, fhirURL1, _, capStat1, tlsVersion1, mimeTypes1, smartResponse1 := getRowInfo(rows)

	for rows.Next() {

		operation2, fhirURL2, entryDate2, capStat2, tlsVersion2, mimeTypes2, smartResponse2 := getRowInfo(rows)

		// If capstat is not null check if current entry that was passed in has capstat equal to capstat of old entry being checked from history table, otherwise check they are both null
		var capStatEqual bool
		var smartResponseEqual bool

		tlsVersionEqual := (tlsVersion1 == tlsVersion2)
		mimeTypesEqual := helpers.StringArraysEqual(mimeTypes1, mimeTypes2)

		if capStat1 != nil {
			capStatEqual = capStat1.EqualIgnore(capStat2)
		} else {
			capStatEqual = (capStat2 == nil)
		}

		if smartResponse1 != nil {
			smartResponseEqual = smartResponse1.EqualIgnore(smartResponse2)
		} else {
			smartResponseEqual = (smartResponse2 == nil)
		}

		equal := capStatEqual && tlsVersionEqual && mimeTypesEqual && smartResponseEqual

		if equal && operation2 != "I" {
			_, err := pruningDeleteStatement.ExecContext(ctx, fhirURL1, entryDate2)
			helpers.FailOnError("", err)
		} else {
			fhirURL1 = fhirURL2
			capStat1 = capStat2
			tlsVersion1 = tlsVersion2
			mimeTypes1 = mimeTypes2
			smartResponse1 = smartResponse2
			continue
		}
	}
}

func getRowInfo(rows *sql.Rows) (string, string, string, capabilityparser.CapabilityStatement, string, []string, capabilityparser.SMARTResponse) {
	var capInt map[string]interface{}
	var fhirURL string
	var operation string
	var capStatJSON []byte
	var entryDate string
	var tlsVersion string
	var mimeTypes []string
	var smartResponseJSON []byte
	var smartResponseInt map[string]interface{}

	err := rows.Scan(&operation, &fhirURL, &capStatJSON, &entryDate, &tlsVersion, pq.Array(&mimeTypes), &smartResponseJSON)
	helpers.FailOnError("", err)

	err = json.Unmarshal(capStatJSON, &capInt)
	helpers.FailOnError("", err)
	capStat, err := capabilityparser.NewCapabilityStatementFromInterface(capInt)
	helpers.FailOnError("", err)

	err = json.Unmarshal(smartResponseJSON, &smartResponseInt)
	helpers.FailOnError("", err)
	smartResponse := capabilityparser.NewSMARTRespFromInterface(smartResponseInt)

	return operation, fhirURL, entryDate, capStat, tlsVersion, mimeTypes, smartResponse
}

func prepareHistoryPruningStatements(s *Store) error {
	var err error
	pruningStatementQueryInterval, err = s.DB.Prepare(`
		SELECT operation, url, capability_statement, entered_at, tls_version, mime_types, smart_response FROM fhir_endpoints_info_history 
		WHERE (operation='U' OR operation='I') 
			AND ((date_trunc('minute', entered_at) <= date_trunc('minute', current_date - interval '" + $1 + "' minute)) 
			AND (date_trunc('minute', entered_at) >= date_trunc('minute', current_date - interval '" + $2 + "' minute))) 
		ORDER BY url, entered_at ASC;`)
	if err != nil {
		return err
	}
	pruningStatementNoQueryInterval, err = s.DB.Prepare(`
		SELECT operation, url, capability_statement, entered_at, tls_version, mime_types, smart_response FROM fhir_endpoints_info_history 
		WHERE (operation='U' OR operation='I') 
			AND (date_trunc('minute', entered_at) <= date_trunc('minute', current_date - interval '" + $1 + "' minute)) 
		ORDER BY url, entered_at ASC;")`)
	if err != nil {
		return err
	}
	pruningDeleteStatement, err = s.DB.Prepare(`
		DELETE FROM fhir_endpoints_info_history WHERE url=$1 AND operation='U' AND entered_at = $2;`)
	return nil
}