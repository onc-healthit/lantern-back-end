package chplendpointquerier

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"strings"
)

func MyheloURLWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList
	doc, err := helpers.ChromedpQueryEndpointList(chplURL, "")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("pre").Each(func(index int, preElem *goquery.Selection) {
		if strings.HasPrefix(preElem.Text(), "https:") {
			var entry LanternEntry

			entryURL := strings.TrimSpace(preElem.Text())
			entry.URL = entryURL

			lanternEntryList = append(lanternEntryList, entry)
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
