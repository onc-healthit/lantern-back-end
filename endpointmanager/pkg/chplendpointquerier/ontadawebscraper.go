package chplendpointquerier

import (
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func OntadaWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList
	var entry LanternEntry

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, ".sc-dTSzeu.dfUAUz")
	if err != nil {
		log.Fatal(err)
	}

	divElem := doc.Find(".sc-dTSzeu.dfUAUz").First()
	spanElem := divElem.Find("span").First()

	entryURL := strings.TrimSpace(spanElem.Text())
	entry.URL = entryURL

	lanternEntryList = append(lanternEntryList, entry)

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
