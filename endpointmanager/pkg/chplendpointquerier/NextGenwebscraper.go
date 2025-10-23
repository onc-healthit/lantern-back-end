package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func NextGenwebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList
	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "#api-search-results")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("#api-search-results").Each(func(index int, div1html *goquery.Selection) {
		div2html := div1html.Find("div").First()
		ulhtml := div2html.Find("ul").First()
		ulhtml.Find("li").Each(func(indextr int, lihtml *goquery.Selection) {
			if strings.Contains(lihtml.Text(), "DSTU2") || strings.Contains(lihtml.Text(), "FHIR R4") {
				var litext = lihtml.Text()
				var URL = litext[strings.Index(litext, " https")+1 : len(litext)-1]
				var entry LanternEntry
				entry.URL = strings.TrimSpace(URL)

				lanternEntryList = append(lanternEntryList, entry)
			}
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
