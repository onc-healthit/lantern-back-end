package chplendpointquerier

import (
	"io"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"os"

	log "github.com/sirupsen/logrus"
)

func NetsmartCSVParser(CHPLURL string, fileToWriteTo string) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	csvFilePath := "./service-base-urls-20240905.csv"

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
		zipcode := strings.TrimSpace(rec[4])
		URL := strings.TrimSpace(rec[5])

		entry.OrganizationName = orgName
		entry.OrganizationZipCode = zipcode
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
