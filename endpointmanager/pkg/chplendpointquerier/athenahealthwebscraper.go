package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
)

func Athenawebscraper(vendorURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc := helpers.ChromedpQueryEndpointList(vendorURL, "table")

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

	WriteCHPLFile(endpointEntryList, fileToWriteTo)

}
