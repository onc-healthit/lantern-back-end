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

	doc.Find(".content").Each(func(index int, contentElem *goquery.Selection) {
		contentElem.Find("p").Each(func(indextr int, pElem *goquery.Selection) {
			if strings.Contains(pElem.Text(), "The base URL for accessing all the ChARM FHIR APIs is below") {
				codeContainingPElem := pElem.Next()
				codeElems := codeContainingPElem.Find("code")
				if codeElems.Length() > 0 {
					var entry LanternEntry

					URL := strings.TrimSpace(codeElems.Eq(0).Text())

					entry.URL = URL

					lanternEntryList = append(lanternEntryList, entry)
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
