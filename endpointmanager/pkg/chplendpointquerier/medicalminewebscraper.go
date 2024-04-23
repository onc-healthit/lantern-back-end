package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func MedicalMineWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".content")
	if err != nil {
		log.Fatal(err)
	}

	found := false
	doc.Find(".content").Each(func(index int, contentElem *goquery.Selection) {
		contentElem.Find("p").Each(func(indextr int, pElem *goquery.Selection) {
			if found {
				return
			}
			codeElems := contentElem.Find("code").First()
			if codeElems.Length() > 0 {
				var entry LanternEntry

				URL := strings.TrimSpace(codeElems.Eq(0).Text())

				entry.URL = URL

				lanternEntryList = append(lanternEntryList, entry)
				found = true
			}
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
