package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func Techcarewebscraper(techcareURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(techcareURL, ".WordSection1")
	if err != nil {
		log.Fatal(err)
	}

	count := 0

	doc.Find(".WordSection1").Each(func(index int, wordSectionElem *goquery.Selection) {
		wordSectionElem.Find("p").Each(func(indextr int, phtml *goquery.Selection) {
			// Only the first one entry is a production server FHIR endpoint
			if count < 1 {
				var entry LanternEntry
				fhirURLLink := phtml.Find("a")
				if fhirURLLink.Length() > 0 {

					fhirURL, ok := fhirURLLink.Attr("href")
					if ok {
						fhirURL = strings.TrimSpace(fhirURL)
						entry.URL = fhirURL
						lanternEntryList = append(lanternEntryList, entry)
					}

					count++
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
