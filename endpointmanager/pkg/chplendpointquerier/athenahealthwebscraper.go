package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func NextGenwebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(vendorURL, "table")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tbody").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("tr").Each(func(indextr int, rowbodyhtml *goquery.Selection) {
				var entryDSTU2 LanternEntry
				var entryR4 LanternEntry
				tableEntries := rowbodyhtml.Find("td")
				if tableEntries.Length() > 0 {
					organizationName := strings.TrimSpace(tableEntries.Eq(1).Text())
					DSTU2URL := strings.TrimSpace(tableEntries.Eq(6).Text())
					R4URL := strings.TrimSpace(tableEntries.Eq(7).Text())

					entryDSTU2.OrganizationName = organizationName
					entryDSTU2.URL = DSTU2URL

					entryR4.OrganizationName = organizationName
					entryR4.URL = R4URL

					lanternEntryList = append(lanternEntryList, entryDSTU2, entryR4)
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
