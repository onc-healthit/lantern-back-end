package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func oneMedicalWebscraper(oneMedicalURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(oneMedicalURL, "#root-url")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("main").Each(func(index int, mainElem *goquery.Selection) {
		mainElem.Find(".gdoc-page").Each(func(index int, gdocPage *goquery.Selection) {
			gdocPage.Find("article").Each(func(index int, articleElem *goquery.Selection) {
				articleElem.Find("p").Each(func(index int, pElem *goquery.Selection) {
					var entry LanternEntry

					if pElem.Length() > 0 {
						if strings.Contains(pElem.Text(), "Production root URL:") {
							aElems := pElem.Find("a")

							if aElems.Length() > 0 {

								entryURL, exists := aElems.Eq(0).Attr("href")

								if exists {
									entryURL = strings.TrimSpace(entryURL)
									entry.URL = entryURL

									lanternEntryList = append(lanternEntryList, entry)
									return
								}
							}
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
