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
	log "github.com/sirupsen/logrus"
)

// PruneInfoHistory checks info table and prunes any repetitive entries
func PruneInfoHistory(ctx context.Context, store *postgresql.Store, queryInterval bool) {

	var distinctURLrows *sql.Rows
	var pruningMetadataCountRows *sql.Rows
	var lastPruneRows *sql.Rows
	var pruningMetadataId int
	var rows *sql.Rows
	var err error

	var lastPruneSuccessful bool
	var lastPruneQueryIntStartDate string
	var lastPruneQueryIntEndDate string

	// LANTERN-724: Check whether data is present in the info history pruning metadata table
	pruningMetadataCountRows, err = store.GetPruningMetadataCount(ctx)
	helpers.FailOnError("", err)

	if !pruningMetadataCountRows.Next() {
		log.Fatal("Error fetching the count of info history pruning metadata table")
	}

	pruningMetadataCount := getPruningMetadataCountRowInfo(pruningMetadataCountRows)

	if pruningMetadataCount > 0 {
		// LANTERN-724: Fetch the last pruned row's entered_at date from the latest entry in the info history pruning metadata table
		lastPruneRows, err = store.GetLastPruneEntryDate(ctx)
		helpers.FailOnError("", err)

		if !lastPruneRows.Next() {
			log.Fatal("Error fetching latest entry from the pruning metadata table")
		}

		lastPruneSuccessful, lastPruneQueryIntStartDate, lastPruneQueryIntEndDate = getLastPruneRowInfo(lastPruneRows)
	}

	// LANTERN-724: Insert an entry in the info history pruning metadata table
	pruningMetadataId, err = store.AddPruningMetadata(ctx, queryInterval, lastPruneSuccessful, lastPruneQueryIntStartDate, lastPruneQueryIntEndDate)
	helpers.FailOnError("", err)

	// LANTERN-724: Initialize counters to determine the number of rows processed and number of rows pruned during history pruning
	numRowsProcessed := 0
	numRowsPruned := 0

	// LANTERN-724: Get distinct URLs from the info history table
	distinctURLrows, err = store.GetDistinctURLs(ctx, queryInterval, lastPruneSuccessful, lastPruneQueryIntStartDate, lastPruneQueryIntEndDate)

	// LANTERN-724: Update the pruning metadata entry anytime an error is thrown
	if err != nil {
		Update(ctx, store, queryInterval, pruningMetadataId, false, numRowsProcessed, numRowsPruned)
		helpers.FailOnError("", err)
	}

	for distinctURLrows.Next() {

		url := getDistinctRowInfo(distinctURLrows)

		rows, err = store.PruningGetInfoHistory(ctx, queryInterval, url, lastPruneSuccessful, lastPruneQueryIntStartDate, lastPruneQueryIntEndDate)

		if err != nil {
			Update(ctx, store, queryInterval, pruningMetadataId, false, numRowsProcessed, numRowsPruned)
			helpers.FailOnError("", err)
		}

		if !rows.Next() {
			return
		}

		_, fhirURL1, _, capStat1, tlsVersion1, mimeTypes1, smartResponse1, _, requestedFhirVersion1 := getRowInfo(rows)
		numRowsProcessed++

		for rows.Next() {

			operation2, fhirURL2, entryDate2, capStat2, tlsVersion2, mimeTypes2, smartResponse2, _, requestedFhirVersion2 := getRowInfo(rows)

			equalFhirEntries := fhirURL1 == fhirURL2

			if equalFhirEntries {
				equalFhirEntries = (requestedFhirVersion1 == requestedFhirVersion2)

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
			}

			if equalFhirEntries && operation2 == "U" {
				err := store.PruningDeleteInfoHistory(ctx, fhirURL1, entryDate2, requestedFhirVersion1)

				if err != nil {
					Update(ctx, store, queryInterval, pruningMetadataId, false, numRowsProcessed, numRowsPruned)
					helpers.FailOnError("", err)
				}

				// valResIDExists, err := store.CheckIfValidationResultIDExists(ctx, valResID2)
				// if err != nil {
				// 	Update(ctx, store, queryInterval, pruningMetadataId, false, numRowsProcessed, numRowsPruned)
				// 	helpers.FailOnError("", err)
				// }

				// // Only delete validations data if it does not exist in fhir_endpoints_info
				// if !valResIDExists {
				// 	// Delete the validation table entries for the history table row
				// 	err = store.PruningDeleteValidationTable(ctx, valResID2)

				// 	if err != nil {
				// 		Update(ctx, store, queryInterval, pruningMetadataId, false, numRowsProcessed, numRowsPruned)
				// 		helpers.FailOnError("", err)
				// 	}

				// 	err = store.PruningDeleteValidationResultEntry(ctx, valResID2)

				// 	if err != nil {
				// 		Update(ctx, store, queryInterval, pruningMetadataId, false, numRowsProcessed, numRowsPruned)
				// 		helpers.FailOnError("", err)
				// 	}
				// }

				numRowsPruned++
			} else {
				fhirURL1 = fhirURL2
				capStat1 = capStat2
				tlsVersion1 = tlsVersion2
				mimeTypes1 = mimeTypes2
				smartResponse1 = smartResponse2
				requestedFhirVersion1 = requestedFhirVersion2
				continue
			}
			numRowsProcessed++
		}
	}

	Update(ctx, store, queryInterval, pruningMetadataId, true, numRowsProcessed, numRowsPruned)
}

func Update(ctx context.Context, store *postgresql.Store, queryInterval bool, pruningMetadataId int, successful bool, numRowsProcessed int, numRowsPruned int) {
	err := store.UpdatePruningMetadata(ctx, pruningMetadataId, successful, numRowsProcessed, numRowsPruned)
	helpers.FailOnError("", err)
}

func getDistinctRowInfo(rows *sql.Rows) string {
	var url string

	err := rows.Scan(&url)
	helpers.FailOnError("", err)

	return url
}

func getPruningMetadataCountRowInfo(rows *sql.Rows) int {
	var count int

	err := rows.Scan(&count)
	helpers.FailOnError("", err)

	return count
}

func getLastPruneRowInfo(rows *sql.Rows) (bool, string, string) {
	var successful bool
	var lastPruneEntryDate string
	var queryIntEndDate string

	err := rows.Scan(&successful, &lastPruneEntryDate, &queryIntEndDate)
	helpers.FailOnError("", err)

	return successful, lastPruneEntryDate, queryIntEndDate
}

func getRowInfo(rows *sql.Rows) (string, string, string, capabilityparser.CapabilityStatement, string, []string, smartparser.SMARTResponse, int, string) {
	var capInt map[string]interface{}
	var fhirURL string
	var operation string
	var capStatJSON []byte
	var entryDate string
	var tlsVersion string
	var mimeTypes []string
	var smartResponseJSON []byte
	var smartResponseInt map[string]interface{}
	var valResIDNullable sql.NullInt64
	var valResID int
	var requestedFhirVersion string

	err := rows.Scan(&operation, &fhirURL, &capStatJSON, &entryDate, &tlsVersion, pq.Array(&mimeTypes), &smartResponseJSON, &valResIDNullable, &requestedFhirVersion)
	helpers.FailOnError("", err)

	if !valResIDNullable.Valid {
		valResID = 0
	} else {
		valResID = int(valResIDNullable.Int64)
	}

	err = json.Unmarshal(capStatJSON, &capInt)
	helpers.FailOnError("", err)
	capStat, err := capabilityparser.NewCapabilityStatementFromInterface(capInt)
	helpers.FailOnError("", err)

	err = json.Unmarshal(smartResponseJSON, &smartResponseInt)
	helpers.FailOnError("", err)
	smartResponse := smartparser.NewSMARTRespFromInterface(smartResponseInt)

	return operation, fhirURL, entryDate, capStat, tlsVersion, mimeTypes, smartResponse, valResID, requestedFhirVersion
}
