package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func IntegraConnectWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tbody").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("tr").Each(func(indextr int, rowbodyhtml *goquery.Selection) {
				var entry LanternEntry
				tableEntries := rowbodyhtml.Find("td")
				if tableEntries.Length() > 0 {
					organizationName := strings.TrimSpace(tableEntries.Eq(0).Text())
					
					aElemURL := tableEntries.Eq(1).Find("a")
					hrefText, exists := aElemURL.Eq(0).Attr("href")
					if exists {
						URL := strings.TrimSpace(hrefText)
						entry.URL = URL
						entry.OrganizationName = organizationName
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
