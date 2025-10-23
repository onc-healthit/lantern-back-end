package chplendpointquerier

import (
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"os"

	log "github.com/sirupsen/logrus"
)

func HealthCare2000SVParser(CHPLURL string, fileToWriteTo string) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	csvFilePath := "./MDVitaFHIRUrls.csv"

	csvReader, file, err := helpers.QueryAndOpenCSV(CHPLURL, csvFilePath, true)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Only want the first URL in the csv file
	rec, err := csvReader.Read()
	if err != nil {
		log.Fatal(err)
	}

	var entry LanternEntry

	URL := strings.TrimSpace(rec[1])

	entry.URL = URL

	lanternEntryList = append(lanternEntryList, entry)

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
