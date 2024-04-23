package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func UnitedHealthURLWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".cmp-text__paragraph")
	if err != nil {
		log.Fatal(err)
	}

	count := 1
	doc.Find("div").Each(func(index int, divhtml *goquery.Selection) {
		dataOpen, exists := divhtml.Attr("data-opensnewwindow")
		if exists && dataOpen != "" && count == 237 {
			pElem := divhtml.Find("p").First()
			if pElem.Length() > 0 && strings.Contains(pElem.Text(), ".fhir.") {
				var entry LanternEntry
				URL := strings.TrimSpace(pElem.Text())
				entry.URL = URL

				lanternEntryList = append(lanternEntryList, entry)
			}
		}
		count++
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
