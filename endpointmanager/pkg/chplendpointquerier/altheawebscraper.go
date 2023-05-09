package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func AltheaWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "div")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("div").Each(func(index int, divElems *goquery.Selection) {
		divElems.Find("div").Each(func(index int, sub1divElems *goquery.Selection) {
			sub1divElems.Find("div").Each(func(index int, sub2divElems *goquery.Selection) {
				sub2divElems.Find("div").Each(func(index int, sub3divElems *goquery.Selection) {
					h2Entries := sub3divElems.Find("h2").First()
					if strings.Contains(h2Entries.Text(), "Production Endpoint") {
						sub3divElems.Find("p").Each(func(index int, pElems *goquery.Selection) {
							bEntries := pElems.Find("b").First()
							if bEntries.Length() > 0 {
								if strings.HasPrefix(bEntries.Text(), " FHIR :") {
									spanEntries := pElems.Find("span")
									if spanEntries.Length() > 0 {
										var entry LanternEntry
										entryURL := strings.TrimSpace(spanEntries.Text())
										entry.URL = entryURL
										lanternEntryList = append(lanternEntryList, entry)
									}
								}
							}

						})
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
