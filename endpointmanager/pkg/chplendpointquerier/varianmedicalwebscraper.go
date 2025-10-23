package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func VarianMedicalWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "#Api_Urls")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("#Api_Urls").Each(func(index int, apiURLsElem *goquery.Selection) {
		apiURLsElem.Find("div").Each(func(index int, divElems *goquery.Selection) {
			pElems := divElems.Find("p")
			if pElems.Length() > 0 {
				pElemURL := pElems.Eq(0)
				aElems := pElemURL.Find("a")
				if aElems.Length() > 0 {
					hrefText, exists := aElems.Eq(0).Attr("href")
					if exists {
						var entry LanternEntry

						fhirURL := strings.TrimSpace(hrefText)
						entry.URL = fhirURL
						lanternEntryList = append(lanternEntryList, entry)
					}
				}
			}
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
