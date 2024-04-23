package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func CignaURLWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".markdown-content-container")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("p").Each(func(index int, phtml *goquery.Selection) {
		if strings.Contains(phtml.Text(), "The base url for each endpoint is: ") {
			aElems := phtml.Find("a")
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

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
