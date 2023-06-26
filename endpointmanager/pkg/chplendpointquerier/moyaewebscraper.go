package chplendpointquerier

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"strings"
)

func MoyaeURLWebscraper(techcareURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(techcareURL, ".sc-fzpans")
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	doc.Find("p").Each(func(indextr int, phtml *goquery.Selection) {
		fhirURL := false
		phtml.Find("strong").Each(func(indextr int, stronghtml *goquery.Selection) {
			spanHtml := phtml.Find("span")
			if spanHtml != nil {
				if strings.HasPrefix(spanHtml.Text(), "FHIR Base URL:") {
					fhirURL = true
				}
			}
		})

		if fhirURL {
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
			fhirURL = false
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
