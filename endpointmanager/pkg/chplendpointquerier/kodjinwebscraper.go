package chplendpointquerier

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"strings"
)

func KodjinWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".md-typeset__table")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("table").Each(func(index int, tableElems *goquery.Selection) {
		tableElems.Find("tbody").Each(func(index int, tbodyElems *goquery.Selection) {
			tbodyElems.Find("tr").Each(func(index int, trElems *goquery.Selection) {
				trElems.Find("td").Each(func(index int, tdElems *goquery.Selection) {
					if strings.HasPrefix(tdElems.Text(), "FHIR API Endpoint") {
						endpointURLEntry := tdElems.Next()
						urlText := endpointURLEntry.Text()
						var entry LanternEntry

						entryURL := strings.TrimSpace(urlText)
						entry.URL = entryURL

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
