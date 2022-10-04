package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func TriMedTechV8Webscraper(trimedtechURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(trimedtechURL, "get-smartconfiguration")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("h4").Each(func(index int, h4Elems *goquery.Selection) {
		if strings.Contains(h4Elems.Text(), "main service base endpoint") {
			h4Elems.Find("a").Each(func(index int, aElems *goquery.Selection) {
				if aElems.Length() > 0 {
					hrefText, exists := aElems.Attr("href")
					if exists && !strings.Contains(hrefText, "#") {
						var entry LanternEntry

						entryURL := strings.TrimSpace(hrefText)
						entry.URL = entryURL

						lanternEntryList = append(lanternEntryList, entry)

						return
					}
				}
			})
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
