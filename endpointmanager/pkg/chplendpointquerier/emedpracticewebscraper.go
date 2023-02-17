package chplendpointquerier

import (
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func eMedPracticeWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "#fhir-api-documentation")
	if err != nil {
		log.Fatal(err)
	}

	apiDocElem := doc.Find("#fhir-api-documentation")
	if apiDocElem.Length() > 0 {
		pElemURL := apiDocElem.Eq(0).Next().Next()
		codeElemURL := pElemURL.Find("code")
		codeElemText := codeElemURL.Eq(0).Text()

		var entry LanternEntry

		entryURL := strings.TrimSpace(codeElemText)
		entry.URL = entryURL

		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
