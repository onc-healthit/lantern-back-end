package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func PCESystemsWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, "table")
	if err != nil {
		log.Fatal(err)
	}

	mainContentTable := doc.Find("table")
	if mainContentTable.Length() > 1 {
		fhirEndpointTable := mainContentTable.Eq(1)
		fhirEndpointTable.Find("tbody").Each(func(index int, tbodyElems *goquery.Selection) {
			tbodyElems.Find("tr").Each(func(index int, trElems *goquery.Selection) {
				trElems.Find("td").Each(func(index int, tdElems *goquery.Selection) {
					if strings.Contains(tdElems.Text(), "Production") {
						productionURLElem := tdElems.Next()
						var entry LanternEntry

						fhirURL := strings.TrimSpace(productionURLElem.Text())
						entry.URL = fhirURL
						lanternEntryList = append(lanternEntryList, entry)
					}
				})
			})
		})
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
