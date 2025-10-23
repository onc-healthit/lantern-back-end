package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func CenteneURLWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".apiDetails_item__1bp0n.MuiBox-root.css-1xdhyk6")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("div").Each(func(index int, divhtml *goquery.Selection) {
		dElem := divhtml.Find("div")
		if dElem.Length() > 1 {

			if strings.Contains(dElem.Eq(0).Text(), "Production") && strings.Contains(dElem.Eq(1).Text(), "production") {
				var entry LanternEntry
				URL := strings.TrimSpace(dElem.Eq(1).Text())
				entry.URL = URL
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
