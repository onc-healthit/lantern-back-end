package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func ElationWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".rdmd-table-inner")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tbody").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("tr").Each(func(indextr int, rowbodyhtml *goquery.Selection) {
				tableEntries := rowbodyhtml.Find("td")
				if tableEntries.Length() > 0 {
					if strings.Contains(tableEntries.Eq(0).Text(), "FHIR API Endpoint") {
						entryURL := strings.TrimSpace(tableEntries.Eq(1).Text())
						if !strings.Contains(entryURL, "sandbox") {
							var entry LanternEntry
							entry.URL = entryURL

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
