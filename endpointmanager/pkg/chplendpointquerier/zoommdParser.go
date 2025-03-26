package chplendpointquerier

import (
	"io"
	"strings"

	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	log "github.com/sirupsen/logrus"
)

func ZoomMDCSVParser(CHPLURL string, fileToWriteTo string) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	csvFilePath := "./zoommd.csv"

	csvReader, file, err := helpers.QueryAndOpenCSV(CHPLURL, csvFilePath, true)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Warnf("error closing file: %v", err)
		}
	}()

	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		var entry LanternEntry

		organizationName := strings.TrimSpace(rec[0])
		if organizationName == "Production FHIR Server Endpoint" {
			URL := strings.TrimSpace(rec[1])

			entry.OrganizationName = organizationName
			entry.URL = URL

			lanternEntryList = append(lanternEntryList, entry)
		}
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
