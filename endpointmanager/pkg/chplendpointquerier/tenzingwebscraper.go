package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func TenzingURLWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, ".humanColumnApiDescription.markdown.formalTheme")
	if err != nil {
		log.Fatal(err)
	}

	fhirEndpointsHeaderElem := doc.Find(".humanColumnApiDescription.markdown.formalTheme")
	if fhirEndpointsHeaderElem.Length() > 0 {
		fhirEndpointsHeaderElem.Find("ul").Each(func(index int, ulElems *goquery.Selection) {
			ulElems.Find("li").Each(func(index int, liElems *goquery.Selection) {
				liElems.Find("p").Each(func(index int, pElems *goquery.Selection) {
					if strings.HasPrefix(pElems.Text(), "FHIR ") {
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
