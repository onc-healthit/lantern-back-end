package chplendpointquerier

import (
	"encoding/csv"
	"io"
	"os"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

// CustomCSVParser reads a CSV from a URL or file and writes processed data to an output file.
// Parameters:
// - inputSource: URL or file path of the input CSV.
// - fileToWriteTo: File path to write the processed data.
// - csvFilePath: Temporary file path for storing the downloaded CSV (if applicable).
// - numrecords: Number of records to process (-1 for all records).
// - startrecord: Starting index of records to process.
// - header: Boolean indicating if the CSV has a header to skip.
// - urlIndex: Column index where the URL is located.
// - organizationIndex: Column index where the organization name is located.
func CustomCSVParser(inputSource string, fileToWriteTo string, csvFilePath string, numrecords int, startrecord int, header bool, urlIndex int, organizationIndex int) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	var csvReader *csv.Reader
	var file *os.File
	var err error
	if strings.HasPrefix(inputSource, "http://") || strings.HasPrefix(inputSource, "https://") {
		csvReader, file, err = helpers.QueryAndOpenCSV(inputSource, csvFilePath, header)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	} else {
		file, err = os.Open(inputSource)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		csvReader = csv.NewReader(file)
		if header {
			_, err := csvReader.Read()
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	records := 0
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if numrecords >= 0 && records >= numrecords+startrecord {
			break
		}
		if records >= startrecord {
			var entry LanternEntry

			organizationName := ""
			if organizationIndex >= 0 {
				organizationName = strings.TrimSpace(rec[organizationIndex])
			}

			if !strings.Contains(strings.ToLower(organizationName), "auth") {

				URL := strings.TrimSpace(rec[urlIndex])
				URL = strings.Replace(URL, "/metadata", "", 1)

				entry.OrganizationName = organizationName
				entry.URL = URL

				lanternEntryList = append(lanternEntryList, entry)

			}
		}

		records++
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(csvFilePath); err == nil {
		err = os.Remove(csvFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}
}
