package chplendpointquerier

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"strings"
)

func OfficePracticumURLWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".MuiTableCell-root.MuiTableCell-body.MuiTableCell-sizeMedium.css-q34dxg")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("tbody.MuiTableBody-root.css-1xnox0e").Each(func(i int, bodyEle *goquery.Selection) {
		bodyEle.Find("tr").Each(func(index int, trEle *goquery.Selection) {
			organizationName := trEle.Find("td").First().Text()
			URL := trEle.Find("td").Next().Text()
			var entry LanternEntry
			entry.OrganizationName = strings.TrimSpace(organizationName)
			entry.URL = strings.TrimSpace(URL)
			lanternEntryList = append(lanternEntryList, entry)
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
