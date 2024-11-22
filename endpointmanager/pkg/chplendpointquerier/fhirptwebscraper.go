package chplendpointquerier

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

func FhirptWebScraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList
	var entry LanternEntry

	entry.URL = strings.TrimSpace(CHPLURL)

	lanternEntryList = append(lanternEntryList, entry)

	endpointEntryList.Endpoints = lanternEntryList

	err := WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
