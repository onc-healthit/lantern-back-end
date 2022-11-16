package chplendpointquerier

import (
	"io"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"os"

	log "github.com/sirupsen/logrus"
)

func CSVParser(CHPLURL string, fileToWriteTo string, csvFilePath string, numrecords int) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	csvReader, file, err := helpers.QueryAndOpenCSV(CHPLURL, csvFilePath)
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
		if records >= numrecords {
			break
		}

		var entry LanternEntry

		organizationName := ""
		URL := strings.TrimSpace(rec[1])

		entry.OrganizationName = organizationName
		entry.URL = URL

		lanternEntryList = append(lanternEntryList, entry)
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
