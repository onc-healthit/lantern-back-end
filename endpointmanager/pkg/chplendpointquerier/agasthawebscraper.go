package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func AgasthaWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, ".col-12")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("p").Each(func(index int, pElem *goquery.Selection) {
		var hrefText = pElem.Text()
		var entry LanternEntry

		entryURL := strings.TrimSpace(hrefText)
		entry.URL = entryURL

		lanternEntryList = append(lanternEntryList, entry)
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
