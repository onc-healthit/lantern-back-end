package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func InteropxWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".api-main-main")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("table").Each(func(index int, tableElem *goquery.Selection) {
		tableElem.Find("tbody").Each(func(index int, bodyElem *goquery.Selection) {
			bodyElem.Find("tr").Each(func(index int, trElem *goquery.Selection) {
				trElem.Find("td").Each(func(columnIndex int, tdElem *goquery.Selection) {
					if columnIndex == 1 {
						var entry LanternEntry

						entryURL := strings.TrimSpace(tdElem.Text())
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
