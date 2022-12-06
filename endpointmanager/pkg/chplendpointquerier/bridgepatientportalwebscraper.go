package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func BridgePatientPortalWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, "#introduction/fhir-bridge-patient-portal/fhir-endpoints")
	if err != nil {
		log.Fatal(err)
	}

	fhirEndpointsHeaderElem := doc.Find("#introduction/fhir-bridge-patient-portal/fhir-endpoints")
	if fhirEndpointsHeaderElem.Length() > 0 {
		spanElem := fhirEndpointsHeaderElem.Eq(0).Next()
		spanElem.Find("ul").Each(func(index int, ulElems *goquery.Selection) {
			ulElems.Find("li").Each(func(index int, liElems *goquery.Selection) {
				liElems.Find("p").Each(func(index int, pElems *goquery.Selection) {
					if strings.Contains(pElems.Text(), "FHIR") {
						aElems := pElems.Find("a")
						if aElems.Length() > 0 {
							hrefText, exists := aElems.Eq(0).Attr("href")
							if exists {
								var entry LanternEntry

								fhirURL := strings.TrimSpace(hrefText)
								entry.URL = fhirURL
								lanternEntryList = append(lanternEntryList, entry)
							}
						}
					}
				})
			})
		})
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
