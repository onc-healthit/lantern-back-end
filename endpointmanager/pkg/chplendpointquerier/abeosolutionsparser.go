package chplendpointquerier

import (
	"io"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"os"

	log "github.com/sirupsen/logrus"
)

func AbeoSolutionsCSVParser(CHPLURL string, fileToWriteTo string) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	csvFilePath := "./FHIRServiceURLs.csv"

	csvReader, file, err := helpers.QueryAndOpenCSV(CHPLURL, csvFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		var entry LanternEntry

		organizationName := ""
		URL := strings.TrimSpace(rec[1])

		entry.OrganizationName = organizationName
		entry.URL = URL

		lanternEntryList = append(lanternEntryList, entry)
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
