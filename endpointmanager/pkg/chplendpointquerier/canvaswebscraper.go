package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func Canvaswebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		if index == 1 {
			tablehtml.Find("tbody").Each(func(indextr int, rowhtml *goquery.Selection) {
				rowhtml.Find("tr").Each(func(indextr int, rowbodyhtml *goquery.Selection) {
					var entry LanternEntry
					tableEntries := rowbodyhtml.Find("td")
					if tableEntries.Length() > 0 {
						organizationName := strings.TrimSpace(tableEntries.Eq(0).Text())
						URL := strings.TrimSpace(tableEntries.Eq(1).Text())

						entry.OrganizationName = organizationName
						entry.URL = URL

						lanternEntryList = append(lanternEntryList, entry)
					}
				})
			})
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
