package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func UnifyWebscraper(unifyURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(unifyURL, ".main-container")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".box").Each(func(index int, boxElems *goquery.Selection) {
		boxElems.Find(".tbl-container").Each(func(index int, tblElems *goquery.Selection) {
			table := tblElems.Find("table")
			tbody := table.Find("tbody")
			tr := tbody.Find("tr").First()
			td1 := tr.Find("td").First()

			if td1.Length() > 0 && strings.Contains(td1.Text(), "Sandbox Base URL") {
				tr.Find("td").Each(func(index int, tdElems *goquery.Selection) {
					if tdElems.Length() > 0 && strings.Contains(tdElems.Text(), "https") {
						var entry LanternEntry

						entryURL := strings.TrimSpace(tdElems.Text())
						entry.URL = entryURL

						lanternEntryList = append(lanternEntryList, entry)

						return
					}
				})
			}
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
