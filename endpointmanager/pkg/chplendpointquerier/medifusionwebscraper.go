package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func MedifusionWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".width50padd")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("ul").Each(func(index int, ulhtml *goquery.Selection) {
		ulhtml.Find("li").Each(func(indextr int, lihtml *goquery.Selection) {
			if strings.Contains(lihtml.Text(), "FHIR Base URL:") {
				divEntries := lihtml.Find("div")
				if divEntries.Length() > 0 {
					preEntries := divEntries.Find("pre")
					var entry LanternEntry
					entryURL := strings.TrimSpace(preEntries.Eq(0).Text())
					entry.URL = entryURL

					lanternEntryList = append(lanternEntryList, entry)

				}
			}
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
