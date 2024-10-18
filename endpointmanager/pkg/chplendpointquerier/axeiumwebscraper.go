package chplendpointquerier

import (
	"strings"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func AxeiumeWebscraper(vendorURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList
	doc, err := helpers.ChromedpQueryEndpointList(vendorURL, "")
	if err != nil {
		log.Fatal(err)
	}
	
	fhirEndpoint, exists := doc.Find("input#fhirEndpoint").Attr("value")
	if exists {
		var entry LanternEntry

		fhirURL := strings.TrimSpace(fhirEndpoint)
		entry.URL = fhirURL
		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
