package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func MedicsCloudWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, "ul")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("li").Each(func(index int, liElem *goquery.Selection) {
		if strings.Contains(liElem.Text(), "Production Endpoint") {
			aElem := liElem.Find("a").First()
			hrefText, exists := aElem.Attr("href")
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
