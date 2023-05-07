package chplendpointquerier

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"io"
	"strings"

	"os"

	log "github.com/sirupsen/logrus"
)

func NetsmartCSVParser(CHPLURL string, fileToWriteTo string) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	csvFilePath := "./netsmart-fhir-base-urls.csv"

	csvReader, file, err := helpers.QueryAndOpenCSV(CHPLURL, csvFilePath, true)
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
		orgName := strings.TrimSpace(rec[0])
		URL := strings.TrimSpace(rec[1])

		entry.OrganizationName = orgName
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
