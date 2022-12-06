package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func MicroFourWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, "#serviceBaseUrls")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("#serviceBaseUrls").Each(func(index int, serviceBaseURLsElem *goquery.Selection) {
		serviceBaseURLsElem.Find("div").Each(func(index int, divElem *goquery.Selection) {
			divElem.Find("p").Each(func(indextr int, pElem *goquery.Selection) {
				if strings.Contains(pElem.Text(), "Service Base URL:") {
					var entry LanternEntry
					URL := strings.ReplaceAll(pElem.Text(), "Service Base URL:", "")

					fhirURL := strings.TrimSpace(URL)
					entry.URL = fhirURL
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
