package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func HMSfirstWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "table")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("tbody").Each(func(index int, rowhtml *goquery.Selection) {
		rowhtml.Find("tr").Each(func(indextr int, rowbodyhtml *goquery.Selection) {
			tableEntries := rowbodyhtml.Find("td")
			if tableEntries.Length() > 0 {
				entryURL := strings.TrimSpace(tableEntries.Eq(0).Text())
				var entry LanternEntry
				entry.URL = entryURL

				lanternEntryList = append(lanternEntryList, entry)
			}
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
