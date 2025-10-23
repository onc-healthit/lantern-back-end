package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func PracticeSuiteWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "#page-header")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("article").Each(func(index int, articleElems *goquery.Selection) {
		articleElems.Find(".entry-content").Each(func(index int, entryElems *goquery.Selection) {
			entryElems.Find("h3").Each(func(index int, h3Elems *goquery.Selection) {
				if strings.Contains(h3Elems.Text(), "Production Endpoint") {
					endpointList := h3Elems.Next()
					endpointList.Find("li").Each(func(index int, liElems *goquery.Selection) {
						if strings.Contains(liElems.Text(), "FHIR") {
							liElems.Find("a").Each(func(index int, aElems *goquery.Selection) {
								hrefText, exists := aElems.Attr("href")
								if exists {
									var entry LanternEntry

									entryURL := strings.TrimSpace(hrefText)
									entry.URL = entryURL

									lanternEntryList = append(lanternEntryList, entry)
								}
							})
						}
					})
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
