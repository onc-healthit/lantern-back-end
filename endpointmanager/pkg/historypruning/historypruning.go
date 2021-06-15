package historypruning

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/smartparser"
)

// PruneInfoHistory checks info table and prunes any repetitive entries
func PruneInfoHistory(ctx context.Context, store *postgresql.Store, queryInterval bool) {
	var rows *sql.Rows
	var err error

	rows, err = store.PruningGetInfoHistory(ctx, queryInterval)
	helpers.FailOnError("", err)

	if !rows.Next() {
		return
	}

	_, fhirURL1, _, capStat1, tlsVersion1, mimeTypes1, smartResponse1, _ := getRowInfo(rows)

	for rows.Next() {

		operation2, fhirURL2, entryDate2, capStat2, tlsVersion2, mimeTypes2, smartResponse2, valResID2 := getRowInfo(rows)

		equalFhirEntries := fhirURL1 == fhirURL2

		if equalFhirEntries {
			equalFhirEntries = (tlsVersion1 == tlsVersion2)

			if equalFhirEntries {
				equalFhirEntries = helpers.StringArraysEqual(mimeTypes1, mimeTypes2)

				if equalFhirEntries {
					// If capstat is not null check if current entry that was passed in has capstat equal to capstat of old entry being checked from history table, otherwise check they are both null
					if capStat1 != nil {
						equalFhirEntries = capStat1.EqualIgnore(capStat2)
					} else {
						equalFhirEntries = (capStat2 == nil)
					}

					if equalFhirEntries {
						// If smartresponse is not null check if current entry that was passed in has smartresponse equal to smartresponse of old entry being checked from history table, otherwise check they are both null
						if smartResponse1 != nil {
							ignoredFields := []string{}
							equalFhirEntries = smartResponse1.EqualIgnore(smartResponse2, ignoredFields)
						} else {
							equalFhirEntries = (smartResponse2 == nil)
						}
					}
				}
			}
		}

		if equalFhirEntries && operation2 == "U" {
			err := store.PruningDeleteInfoHistory(ctx, fhirURL1, entryDate2)
			helpers.FailOnError("", err)
			// Delete the validation table entries for the history table row
			err = store.PruningDeleteValidationTable(ctx, valResID2)
			helpers.FailOnError("", err)
			err = store.PruningDeleteValidationResultEntry(ctx, valResID2)
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

func getRowInfo(rows *sql.Rows) (string, string, string, capabilityparser.CapabilityStatement, string, []string, smartparser.SMARTResponse, int) {
	var capInt map[string]interface{}
	var fhirURL string
	var operation string
	var capStatJSON []byte
	var entryDate string
	var tlsVersion string
	var mimeTypes []string
	var smartResponseJSON []byte
	var smartResponseInt map[string]interface{}
	var valResID int

	err := rows.Scan(&operation, &fhirURL, &capStatJSON, &entryDate, &tlsVersion, pq.Array(&mimeTypes), &smartResponseJSON, &valResID)
	helpers.FailOnError("", err)

	err = json.Unmarshal(capStatJSON, &capInt)
	helpers.FailOnError("", err)
	capStat, err := capabilityparser.NewCapabilityStatementFromInterface(capInt)
	helpers.FailOnError("", err)

	err = json.Unmarshal(smartResponseJSON, &smartResponseInt)
	helpers.FailOnError("", err)
	smartResponse := smartparser.NewSMARTRespFromInterface(smartResponseInt)

	return operation, fhirURL, entryDate, capStat, tlsVersion, mimeTypes, smartResponse, valResID
}
