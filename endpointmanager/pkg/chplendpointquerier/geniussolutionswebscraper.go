package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func GeniusSolutionsWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".topicContent")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("table").Each(func(index int, tableElems *goquery.Selection) {
		tableElems.Find("tbody").Each(func(index int, tbodyElems *goquery.Selection) {
			tbodyElems.Find("tr").Each(func(index int, trElems *goquery.Selection) {
				trElems.Find("td").Each(func(index int, tdElems *goquery.Selection) {
					if tdElems.Has("b").Length() > 0 {
						bElem := tdElems.Find("b").First()
						if strings.HasPrefix(bElem.Text(), "Api Endpoint:") {
							endpointURLEntry := tdElems.Next()
							endpointURLElem := endpointURLEntry.Find("span").First()
							urlText := endpointURLElem.Text()
							var entry LanternEntry

							entryURL := strings.TrimSpace(urlText)
							entry.URL = entryURL

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
