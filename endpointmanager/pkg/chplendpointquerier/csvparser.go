package chplendpointquerier

import (
	"io"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"os"

	log "github.com/sirupsen/logrus"
)

func CSVParser(CHPLURL string, fileToWriteTo string, csvFilePath string, numrecords int, startrecord int, header bool, urlIndex int, organizationIndex int) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	csvReader, file, err := helpers.QueryAndOpenCSV(CHPLURL, csvFilePath, header)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

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

			URL := strings.TrimSpace(rec[urlIndex])

			entry.OrganizationName = organizationName
			entry.URL = URL

			lanternEntryList = append(lanternEntryList, entry)
		}

		records++
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove(csvFilePath)
	if err != nil {
		log.Fatal(err)
	}
}
