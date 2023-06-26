package chplendpointquerier

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func ImedemrWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "table")
	if err != nil {
		log.Fatal(err)
	}

	tableElems := doc.Find("table")
	if tableElems.Length() > 0 {
		tableElem := tableElems.Eq(1)

		tableElem.Find("tbody").Each(func(index int, tbodyElem *goquery.Selection) {
			tbodyElem.Find("tr").Each(func(trIndex int, trElem *goquery.Selection) {
				if trIndex >= 1 {
					tdElem := trElem.Find("td")
					org := tdElem.Eq(0)
					URL := tdElem.Eq(1)
					var entry LanternEntry
					entry.OrganizationName = org.Text()
					entry.URL = URL.Text()
					lanternEntryList = append(lanternEntryList, entry)

				}
			})
		})
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
