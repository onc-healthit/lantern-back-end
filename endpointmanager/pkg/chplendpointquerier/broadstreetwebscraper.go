package chplendpointquerier

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"strings"
)

func BroadStreetURLWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "article")
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("article").Each(func(index int, articleElems *goquery.Selection) {
		articleElems.Find("table").Each(func(index int, tableElems *goquery.Selection) {
			tableElems.Find("tbody").Each(func(index int, bodyElems *goquery.Selection) {
				bodyElems.Find("tr").Each(func(index int, trElems *goquery.Selection) {
					tdEntry := trElems.Find("td").First()
					if strings.Contains(tdEntry.Text(), "FHIR Production") {
						tdEntryNext := trElems.Find("td").Next()
						if tdEntryNext.Length() > 0 {
							var entry LanternEntry
							entry.URL = strings.TrimSpace(tdEntryNext.Text())
							lanternEntryList = append(lanternEntryList, entry)
						}
					}
				})
			})
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
