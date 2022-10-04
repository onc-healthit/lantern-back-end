package endpointwebscraper

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func Techcarewebscraper(vendorURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc := helpers.ChromedpQueryEndpointList(vendorURL, ".WordSection1")

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

	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
