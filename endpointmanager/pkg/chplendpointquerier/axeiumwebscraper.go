package chplendpointquerier

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

func AxeiumeWebscraper(vendorURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	var entry LanternEntry

	fhirURL := strings.TrimSpace(vendorURL) + "r4"
	entry.URL = fhirURL
	lanternEntryList = append(lanternEntryList, entry)

	endpointEntryList.Endpoints = lanternEntryList

	err := WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
