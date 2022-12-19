package chplendpointquerier

import (
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/PuerkitoBio/goquery"
)

func eMedPracticeWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "#OverView")
	if err != nil {
		log.Fatal(err)
	}


	doc.Find("#OverView").Each(func(index int, overviewElems *goquery.Selection) {
		overviewElems.Find("pre").Each(func(preIndex int, preElems *goquery.Selection) {
			if preIndex == 0 {
				preElems.Find("code").Each(func(index int, codeElems *goquery.Selection) {
					codeElemText := codeElems.Eq(0).Text()
					codeElemText = strings.TrimPrefix(codeElemText, "GET")
					urlText := strings.Split(codeElemText, "/Patient")

					var entry LanternEntry

					entryURL := strings.TrimSpace(urlText[0])
					entry.URL = entryURL

					lanternEntryList = append(lanternEntryList, entry)
				})
			}
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
