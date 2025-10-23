package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func NextgenAPIWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList
	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".container")
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(".container").Each(func(index int, containerhtml *goquery.Selection) {
		containerhtml.Find("ul").Each(func(indextr int, ulhtml *goquery.Selection) {
			ulhtml.Find("li").Each(func(indextr int, lihtml *goquery.Selection) {
				lihtml.Find("ul").Each(func(indextr int, ul2html *goquery.Selection) {
					ul2html.Find("li").Each(func(indextr int, li2html *goquery.Selection) {
						li2html.Find("ul").Each(func(indextr int, ul3html *goquery.Selection) {
							ul3html.Find("li").Each(func(indextr int, li3html *goquery.Selection) {
								if strings.Contains(li3html.Text(), "R4 Version Base URL") {
									aElem := li3html.Find("a").First()
									hrefText, exists := aElem.Attr("href")
									if exists {
										var entry LanternEntry

										entryURL := strings.TrimSpace(hrefText)
										entry.URL = entryURL

										lanternEntryList = append(lanternEntryList, entry)
									}
								}
							})
						})
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
