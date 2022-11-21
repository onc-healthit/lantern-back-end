package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func CorrecTekWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	stu3EndpointList := "https://www.interopengine.com/2017/open-api-documentation.html"
	r4EndpointList := "https://www.interopengine.com/2021/open-api-documentation.html"
	count := 0

	fileToWriteTo = strings.TrimSuffix(fileToWriteTo, "EndpointSources.json")
	
	for count <= 1 {
		endpointListURL := stu3EndpointList
		if count == 1 {
			endpointListURL = r4EndpointList
		}
		
		doc, err := helpers.ChromedpQueryEndpointList(endpointListURL, "article")
		if err != nil {
			log.Fatal(err)
		}

		doc.Find("article").Each(func(index int, articleElem *goquery.Selection) {
			articleElem.Find("h4").Each(func(index int, h4Elem *goquery.Selection) {
				if strings.Contains(h4Elem.Text(), "General Concepts") {
					pElemURL:= h4Elem.Next().Next()
					aElems := pElemURL.Find("a")
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

		endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, lanternEntryList...)
		
		count++
	}

	err := WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
