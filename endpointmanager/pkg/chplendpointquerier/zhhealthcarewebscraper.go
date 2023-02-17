package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func ZHHealthcareWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, "table")
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("table").Each(func(index int, tableElem *goquery.Selection) {
		tableElem.Find("tbody").Each(func(index int, tbodyElem *goquery.Selection) {
			tbodyElem.Find("tr").Each(func(indextr int, trElem *goquery.Selection) {
				if indextr == 1 {
					var entry LanternEntry
					tableEntries := trElem.Find("td")
					if tableEntries.Length() > 0 {
						fhirURL := strings.TrimSpace(tableEntries.Eq(1).Text())

						entry.URL = fhirURL

						lanternEntryList = append(lanternEntryList, entry)
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
