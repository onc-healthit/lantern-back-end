package chplendpointquerier

import (
	"strings"
	"fmt"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/PuerkitoBio/goquery"

)

func eMedPracticeWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "#bulkDataExport")
	if err != nil {
		log.Fatal(err)
	}

	urlFound := false

	doc.Find("ul").Each(func(index int, ulElems *goquery.Selection) {
		ulElems.Find("li").Each(func(index int, liElems *goquery.Selection) {
			if strings.Contains(liElems.Text(), "https:") && !urlFound {
				divElems := liElems.Find("div")
				if divElems.Length() > 0 {
					preElems := liElems.Find("pre")
					if preElems.Length() > 0 {
						var entry LanternEntry
						
						urlStart := strings.Index(preElems.Text(), "https://")
						if urlStart == -1 {
							fmt.Println("URL not found")
							return
						}

						urlEnd := strings.Index(preElems.Text()[urlStart:], "8443/")
						if urlEnd == -1 {
							fmt.Println("End of base URL not found")
							return
						}

						fhirURL := preElems.Text()[urlStart : urlStart+urlEnd+len("8443")]
						entry.URL = fhirURL
						lanternEntryList = append(lanternEntryList, entry)
						
						urlFound = true
						
						return
					}
				}
			}
		})

		if urlFound {
			return
		}
	})

	endpointEntryList.Endpoints = lanternEntryList
	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
