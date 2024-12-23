package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func CarefluenceWebscraper(vendorURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(vendorURL, "main")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("main").Each(func(index int, mainContent *goquery.Selection) {
		mainContent.Find("div").Each(func(index int, div1Content *goquery.Selection) {
			div1Content.Find("div").Each(func(index int, div2Content *goquery.Selection) {
				div2Content.Find("p").Each(func(indextr int, phtml *goquery.Selection) {
					var entry LanternEntry

					fhirURL := strings.TrimSpace(phtml.Text())
					entry.URL = fhirURL
					lanternEntryList = append(lanternEntryList, entry)
				})
			})
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
