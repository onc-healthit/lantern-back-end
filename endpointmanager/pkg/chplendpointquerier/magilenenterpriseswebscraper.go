package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func MagilenEnterprisesWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "table")
	if err != nil {
		log.Fatal(err)
	}

	tableElems := doc.Find("table")
	if tableElems.Length() > 0 {
		tableElem := tableElems.Eq(0)

		tableElem.Find("tbody").Each(func(index int, tbodyElem *goquery.Selection) {
			tbodyElem.Find("tr").Each(func(trIndex int, trElem *goquery.Selection) {
				if trIndex == 1 {
					trElem.Find("td").Each(func(tdIndex int, tdElem *goquery.Selection) {
						if trIndex == 1 {
							pElem := tdElem.Find("p")
							if pElem.Length() > 1 {
								firstHalfURLSpan := pElem.Eq(0).Find("span")
								secondHalfURLSpan := pElem.Eq(1).Find("span")

								firstHalfURLAElem := firstHalfURLSpan.Find("a")
								if firstHalfURLAElem.Length() > 0 {
									fhirstHalfURLText, exists := firstHalfURLAElem.Eq(0).Attr("href")
									if exists {

										secondHalfURLText := secondHalfURLSpan.Text()

										var entry LanternEntry

										fhirURL := strings.TrimSpace(fhirstHalfURLText + secondHalfURLText)
										entry.URL = fhirURL
										lanternEntryList = append(lanternEntryList, entry)
									}
								}
							}
						}
					})
				}
			})
		})
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
