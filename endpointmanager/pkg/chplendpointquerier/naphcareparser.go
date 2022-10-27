package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func NaphCareWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".container")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".container").Each(func(index int, bodyElems *goquery.Selection) {
		bodyElems.Find(".WordSection1").Each(func(index int, wordSectionElems *goquery.Selection) {
			bodyTextElems := wordSectionElems.Find(".BodyText")
			productionEndpointElem := bodyTextElems.First()
			endpointURLElem := productionEndpointElem.Find("a")
			hrefText, exists := endpointURLElem.Attr("href")
			if exists {
				var entry LanternEntry

				entryURL := strings.TrimSpace(hrefText)
				entry.URL = entryURL

				lanternEntryList = append(lanternEntryList, entry)
			}
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
