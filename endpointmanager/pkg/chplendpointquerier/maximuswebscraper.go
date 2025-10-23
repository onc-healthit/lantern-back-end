package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func MaximusURLWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, ".width50padd")
	if err != nil {
		log.Fatal(err)
	}

	fhirEndpointsHeaderElem := doc.Find(".width50padd")
	if fhirEndpointsHeaderElem.Length() > 0 {
		fhirEndpointsHeaderElem.Find("ul").Each(func(index int, ulElems *goquery.Selection) {
			ulElems.Find("li").Each(func(index int, liElems *goquery.Selection) {
				if strings.Contains(liElems.Text(), "FHIR Base URL:") {
					divElems := liElems.Find("div")
					if divElems.Length() > 0 {
						preElems := liElems.Find("pre")
						if preElems.Length() > 0 {
							var entry LanternEntry

							fhirURL := strings.TrimSpace(preElems.Text())
							entry.URL = fhirURL
							lanternEntryList = append(lanternEntryList, entry)
						}
					}
				}
			})
		})
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
