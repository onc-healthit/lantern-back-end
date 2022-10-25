package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func IntelichartWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "#auth-exception-table")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("#auth-exception-table").Each(func(index int, tableElems *goquery.Selection) {
		tableElems.Find("tbody").Each(func(index int, tbodyElems *goquery.Selection) {
			tbodyElems.Find("tr").Each(func(index int, trElems *goquery.Selection) {
				trElems.Find("td").Each(func(index int, tdElems *goquery.Selection) {
					if strings.HasPrefix(tdElems.Text(), "Production API") {
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
