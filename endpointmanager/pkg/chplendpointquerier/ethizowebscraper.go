package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func EthizoWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, "ul")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("li").Each(func(index int, liElem *goquery.Selection) {
		spanEntries := liElem.Find("span")
		if spanEntries.Length() > 0 {
			name := strings.TrimSpace(spanEntries.Eq(0).Text())
			if strings.Contains(name, "Production") {
				aElems := liElem.Find("a")
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
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
