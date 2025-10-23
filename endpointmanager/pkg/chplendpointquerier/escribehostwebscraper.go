package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func EscribeHOSTWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".banner2")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("div").Each(func(index int, divElems *goquery.Selection) {
		subDivEntries := divElems.Find("div")
		if subDivEntries.Length() > 1 {
			label := strings.TrimSpace(subDivEntries.Eq(0).Text())

			if strings.Contains(label, "FHIR Server Base URL") {
				urlDiv := subDivEntries.Eq(1)
				aElem := urlDiv.Find("a").First()
				hrefText, exists := aElem.Attr("href")
				if exists {
					var entry LanternEntry

					entryURL := strings.TrimSpace(hrefText)
					entry.URL = entryURL

					lanternEntryList = append(lanternEntryList, entry)
				}
			}
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
