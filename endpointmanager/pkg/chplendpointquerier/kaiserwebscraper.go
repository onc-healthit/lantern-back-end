package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func KaiserURLWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".language-json")
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(".language-json").Each(func(index int, codehtml *goquery.Selection) {
		found := false
		processed := false
		codehtml.Find("span").Each(func(index int, spanhtml *goquery.Selection) {
			if strings.Contains(spanhtml.Text(), "CapabilityStatement") {
				found = true
			}
			if found {
				if strings.Contains(spanhtml.Text(), "/FHIR/api") {
					var entry LanternEntry
					URL := strings.TrimSpace(spanhtml.Text())
					entry.URL = URL
					processed = true
					lanternEntryList = append(lanternEntryList, entry)

				}
			}
			if processed {
				return
			}

		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
