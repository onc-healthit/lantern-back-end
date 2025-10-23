package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func IndianHealthWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL+"api-documentation/", ".container-fluid")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("#site_content").Each(func(index int, siteContentElems *goquery.Selection) {
		siteContentElems.Find(".mura-region-loose").Each(func(index int, muraRegionElems *goquery.Selection) {
			muraRegionElems.Find(".mura-region-local").Each(func(index int, muraRegionLocalElems *goquery.Selection) {
				muraRegionLocalElems.Find(".row").Each(func(index int, rowElems *goquery.Selection) {
					rowElems.Find(".col-md-4").Each(func(index int, colElems *goquery.Selection) {
						colElems.Find(".panel").Each(func(index int, panelElems *goquery.Selection) {
							panelElems.Find(".panel-body").Each(func(index int, panelBodyElems *goquery.Selection) {
								panelBodyElems.Find("ul").Each(func(index int, ulElems *goquery.Selection) {
									panelBodyElems.Find("li").Each(func(index int, liElems *goquery.Selection) {
										aElem := liElems.Find("a").First()
										if strings.Contains(aElem.Text(), "Single Patient EndPoint") {
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
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
