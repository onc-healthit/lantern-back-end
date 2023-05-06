package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func WebchartNowWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".table")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("tbody").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			tableEntries := rowhtml.Find("td")
			if tableEntries.Length() > 1 {
				if strings.Contains(tableEntries.Eq(0).Text(), "Maui Medical Group") {

					var entry LanternEntry
					entry.OrganizationName = tableEntries.Eq(0).Text()
					entryURL := strings.TrimSpace(tableEntries.Eq(1).Text())
					entry.URL = entryURL

					lanternEntryList = append(lanternEntryList, entry)
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
