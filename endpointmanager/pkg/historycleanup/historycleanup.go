package historycleanup

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/smartparser"
	log "github.com/sirupsen/logrus"
)

// GetInfoHistoryDuplicateData checks info table and stores the identifiers of any repetitive entries in CSV files
func GetInfoHistoryDuplicateData(ctx context.Context, store *postgresql.Store, queryInterval bool) {

	historyRowCount := 1
	historyDuplicateRowCount := 1

	var rows *sql.Rows
	var distinctURLrows *sql.Rows
	var err error
	var existingDistinctURLs []string
	var URLCaptured bool

	// Get distinct URLs from the history table
	distinctURLrows, err = store.GetDistinctURLsFromHistory(ctx)
	helpers.FailOnError("", err)

	// Open (or create if not present) csv files (in APPEND mode) to store list of distinct URLs and pruning data identifiers
	// NOTE: This will create CSV files in the /home directory of the lantern-back-end-endpoint_manager-1 container
	distinctURLfile, err := os.OpenFile("/home/distinctURLsFromHistory.csv", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Failed to create file: %s", err)
	}
	defer distinctURLfile.Close()

	// Read the distinctURLsFromHistory file to check whether URLs are already added to it
	csvReader := csv.NewReader(distinctURLfile)
	csvData, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV data: %v\n", err)
	}

	// Ignore the URLs already added during the pruning data capture operation
	if len(csvData) > 0 {
		log.Info("Existing distinctURLsFromHistory file detected. URLs already present in this file will be ignored.")
		existingDistinctURLs = flatten2D(csvData)
	}

	duplicateInfoHistoryFile, err := os.OpenFile("/home/duplicateInfoHistoryIds.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to create file: %s", err)
	}
	defer duplicateInfoHistoryFile.Close()

	// Create CSV writers
	distinctURLWriter := csv.NewWriter(distinctURLfile)
	duplicateInfoHistoryDataWriter := csv.NewWriter(duplicateInfoHistoryFile)

	log.Info("Starting the duplicate info history data capture.")

	for distinctURLrows.Next() {

		url := getDistinctRowInfo(distinctURLrows)
		URLCaptured = false

		// Check whether duplicate data is already captured for the given URL
		for idx, val := range existingDistinctURLs {
			if url == val {
				log.Info("Duplicate info history data already captured. Ignoring URL: ", url)

				// Set the flag
				URLCaptured = true

				// Remove the URL from the list of existing URLs
				existingDistinctURLs = append(existingDistinctURLs[:idx], existingDistinctURLs[idx+1:]...)
				break
			}
		}

		// Skip the current iteration if duplicate data is already captured
		if URLCaptured {
			continue
		}

		rows, err = store.PruningGetInfoHistoryUsingURL(ctx, queryInterval, url)
		helpers.FailOnError("", err)

		if !rows.Next() {
			return
		}

		var pruningData [][]string
		_, fhirURL1, _, capStat1, tlsVersion1, mimeTypes1, smartResponse1, _, requestedFhirVersion1 := getRowInfo(rows)

		for rows.Next() {
			log.Info("Info History Row Count: ", historyRowCount)
			historyRowCount++
			operation2, fhirURL2, entryDate2, capStat2, tlsVersion2, mimeTypes2, smartResponse2, valResID2, requestedFhirVersion2 := getRowInfo(rows)

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
				log.Info("Duplicate Info History Row Count: ", historyDuplicateRowCount)
				historyDuplicateRowCount++
				log.Infof("Duplicate Data Captured :: URL: %s, Entered At: %s, Requested FHIR Version: %s, Validation Result ID: %s", fhirURL2, entryDate2, requestedFhirVersion2, strconv.Itoa(valResID2))
				pruningData = append(pruningData, []string{fhirURL2, entryDate2, requestedFhirVersion2, strconv.Itoa(valResID2)})
			} else {
				fhirURL1 = fhirURL2
				capStat1 = capStat2
				tlsVersion1 = tlsVersion2
				mimeTypes1 = mimeTypes2
				smartResponse1 = smartResponse2
				requestedFhirVersion1 = requestedFhirVersion2
				continue
			}
		}

		err = duplicateInfoHistoryDataWriter.WriteAll(pruningData)
		if err != nil {
			log.Fatal("Error writing to duplicateInfoHistoryDataWriter:", err)
		}

		duplicateInfoHistoryDataWriter.Flush()
		if err := duplicateInfoHistoryDataWriter.Error(); err != nil {
			log.Fatal("Error flushing duplicateInfoHistoryDataWriter:", err)
		}

		err = distinctURLWriter.Write([]string{url})
		if err != nil {
			log.Fatal("Error writing to distinctURLWriter:", err)
		}

		distinctURLWriter.Flush()
		if err := distinctURLWriter.Error(); err != nil {
			log.Fatal("Error flushing distinctURLWriter:", err)
		}
	}
}

// flatten2D converts a 2D slice to a 1D slice
func flatten2D(data2D [][]string) []string {
	var data1D []string
	for _, row := range data2D {
		data1D = append(data1D, row...)
	}
	return data1D
}

func getDistinctRowInfo(rows *sql.Rows) string {
	var url string

	err := rows.Scan(&url)
	helpers.FailOnError("", err)

	return url
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
