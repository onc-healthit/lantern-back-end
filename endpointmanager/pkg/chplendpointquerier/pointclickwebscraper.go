package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func PointclickWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList
	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".container")
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(".container").Each(func(index int, containerhtml *goquery.Selection) {
		containerhtml.Find("section").Each(func(index int, sectionhtml *goquery.Selection) {
			sectionhtml.Find("ol").Each(func(indextr int, olhtml *goquery.Selection) {
				olhtml.Find("li").Each(func(indextr int, lihtml *goquery.Selection) {
					lihtml.Find("ol").Each(func(indextr int, ol2html *goquery.Selection) {
						ol2html.Find("li").Each(func(indextr int, li2html *goquery.Selection) {
							if strings.Contains(li2html.Text(), "Note: The base FHIR URL") {
								aElem := li2html.Find("a").First()
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

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
