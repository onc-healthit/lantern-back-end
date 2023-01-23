package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func MeridianWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, ".fhirurl-conatainer")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".fhirurl-conatainer").Each(func(index int, containerElem *goquery.Selection) {
		containerElem.Find(".contact").Each(func(index int, contactElem *goquery.Selection) {
			contactElem.Find(".fhirurl-para").Each(func(index int, pElem *goquery.Selection) {
				pElem.Find("a").Each(func(indextr int, aElems *goquery.Selection) {
					hrefText, exists := aElems.Eq(0).Attr("href")
					if exists {
						var entry LanternEntry

						fhirURL := strings.TrimSpace(hrefText)
						entry.URL = fhirURL
						lanternEntryList = append(lanternEntryList, entry)
					}
				})
			})
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
