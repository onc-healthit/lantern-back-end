package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func PraxisEMRWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".myClass")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("p").Each(func(index int, pElems *goquery.Selection) {
		if strings.HasPrefix(pElems.Text(), "Production root URL:") {
			endpointURLElem := pElems.Find("a")
			hrefText, exists := endpointURLElem.Attr("href")
			if exists {
				var entry LanternEntry

				entryURL := strings.TrimSpace(hrefText)
				entry.URL = entryURL

				lanternEntryList = append(lanternEntryList, entry)
			}
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
