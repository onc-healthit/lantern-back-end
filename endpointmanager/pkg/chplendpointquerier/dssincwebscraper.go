package chplendpointquerier

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"strings"
)

func DssIncWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "#Api_Urls")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("#Api_Urls").Each(func(index int, tableElems *goquery.Selection) {
		tableElems.Find("div").Each(func(index1 int, divElems *goquery.Selection) {
			divElems.Find("p").Each(func(indextr int, phtml *goquery.Selection) {
				if strings.Contains(phtml.Text(), "API Base URL") {
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
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
