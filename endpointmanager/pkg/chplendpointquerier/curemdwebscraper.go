package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func CuremdWebscraper(CHPLURL string, fileToWriteTo string) {

	var entry LanternEntry
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("label").Each(func(index int, item *goquery.Selection) {
		labelText := strings.TrimSpace(item.Text())
		if labelText == "Capability Statement:" {
			pTag := item.NextFiltered("p")
			if pTag.Length() > 0 {
				pTag.Find("a").Each(func(i int, link *goquery.Selection) {
					url, exists := link.Attr("href")
					url = strings.TrimSuffix(url, "/metadata")
					if exists {
						entry.URL = url
						lanternEntryList = append(lanternEntryList, entry)
					}
				})
			}
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
