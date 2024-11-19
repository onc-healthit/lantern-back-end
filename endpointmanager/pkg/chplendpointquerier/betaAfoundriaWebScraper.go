package chplendpointquerier

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func BetaAfoundriaWebScraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".container")
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(".container h3").Each(func(index int, header *goquery.Selection) {
		nextAnchor := header.NextFiltered("a")
		href, exists := nextAnchor.Attr("href")
		if exists {
			var entry LanternEntry
			entry.URL = href
			lanternEntryList = append(lanternEntryList, entry)
		}
	})

	endpointEntryList.Endpoints = lanternEntryList
	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
