package endpointwebscraper

import (
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"

	"github.com/PuerkitoBio/goquery"
)

func HTMLtablewebscraper(vendorURL string, vendor string, fileToWriteTo string) {

	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(vendorURL, "")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			var entry LanternEntry
			tableEntries := rowhtml.Find("td")
			if tableEntries.Length() > 0 {
				if vendor == "CareEvolution" {
					if indextr != 1 {
						entry.OrganizationName = strings.TrimSpace(tableEntries.Eq(0).Text())
						entry.URL = strings.TrimSpace(tableEntries.Eq(1).Text())
						endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, entry)
					}
				}
			}
		})
	})

	err = WriteEndpointListFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
