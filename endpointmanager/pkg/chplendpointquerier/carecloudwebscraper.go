package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func CareCloudWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, ".contact")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("p.fhirurl-para").Each(func(index int, pElem *goquery.Selection) {
		aElem := pElem.Find("a").First()
		hrefText, exists := aElem.Attr("href")
		if exists {
			var entry LanternEntry

			entryURL := strings.TrimSpace(hrefText)
			entry.URL = entryURL

			lanternEntryList = append(lanternEntryList, entry)
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
