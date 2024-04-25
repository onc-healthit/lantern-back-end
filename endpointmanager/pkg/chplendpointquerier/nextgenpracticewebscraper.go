package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func NextgenPracticeWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList
	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".col-md-12.margin-top25")
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(".col-md-12.margin-top25").Each(func(index int, divhtml *goquery.Selection) {
		divhtml.Find("ul").Each(func(indextr int, ulhtml *goquery.Selection) {
			ulhtml.Find("li").Each(func(indextr int, lihtml *goquery.Selection) {
				if strings.Contains(lihtml.Text(), "Patient Access Endpoint FHIR R4") {
					parts := strings.Split(lihtml.Text(), " - ")
					var entry LanternEntry

					entryURL := strings.TrimSpace(parts[1])
					entry.URL = entryURL

					lanternEntryList = append(lanternEntryList, entry)
				}
			})
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
