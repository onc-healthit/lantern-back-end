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
		h3Elem := boxElems.Find("h3")
		if h3Elem.Length() > 0 && h3Elem.Text() == "Request" {
			pEntries := boxElems.Find("p")

			if pEntries.Length() > 0 && strings.Contains(pEntries.Text(), "FHIR Base URL: ") {
				var entry LanternEntry

				aElem := pEntries.Find("a")

				entryURL := strings.TrimSpace(aElem.Text())
				entry.URL = entryURL

				lanternEntryList = append(lanternEntryList, entry)

				return
			}
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
