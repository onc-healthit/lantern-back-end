package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func HealthSamuraiWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".container")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".container").Each(func(index int, containterElems *goquery.Selection) {
		containterElems.Find(".row").Each(func(index int, rowElems *goquery.Selection) {
			rowElems.Find(".col-12").Each(func(index int, colElems *goquery.Selection) {
				colElems.Find("ul").Each(func(index int, ulElems *goquery.Selection) {
					ulElems.Find("li").Each(func(index int, liElems *goquery.Selection) {
						var entry LanternEntry

						entryURL := strings.TrimSpace(liElems.Text())
						entry.URL = entryURL

						lanternEntryList = append(lanternEntryList, entry)
					})
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
