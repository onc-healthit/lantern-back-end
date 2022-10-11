package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func Athenawebscraper(vendorURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(vendorURL, "table")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("app-api-servers").Each(func(index int, apiServers *goquery.Selection) {
		apiServers.Find("table").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("tr").Each(func(indextr int, rowbodyhtml *goquery.Selection) {
				var entry LanternEntry
				tableEntries := rowbodyhtml.Find("td")
				if tableEntries.Length() > 0 {
					organizationName := strings.TrimSpace(tableEntries.Eq(0).Text())
					fhirURL := strings.TrimSpace(tableEntries.Eq(1).Text())

					entry.OrganizationName = organizationName
					entry.URL = fhirURL

					lanternEntryList = append(lanternEntryList, entry)
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
