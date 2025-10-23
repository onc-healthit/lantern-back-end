package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func SmarteMRWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".markdown-body")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("p").Each(func(index int, pElems *goquery.Selection) {
		if strings.Contains(pElems.Text(), "Production URL:") {
			endpointURLElems := pElems.Find("a")
			if endpointURLElems.Length() > 1 {
				hrefText, exists := endpointURLElems.Eq(1).Attr("href")
				if exists {
					var entry LanternEntry

					entryURL := strings.TrimSpace(hrefText)
					entry.URL = entryURL

					lanternEntryList = append(lanternEntryList, entry)
				}
			}
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
