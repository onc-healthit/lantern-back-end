package chplendpointquerier

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func AspMDeWebscraper(vendorURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList
	found := false
	doc, err := helpers.ChromedpQueryEndpointList(vendorURL, "p")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("h1").Each(func(i int, s *goquery.Selection) {
		sectionTitle := s.Text()

		if sectionTitle == "Public Endpoint" {
			s.Parent().Find("a").Each(func(j int, link *goquery.Selection) {
				if found {
					return
				}
				href, exists := link.Attr("href")
				if exists {
					var entry LanternEntry
					entry.URL = href
					lanternEntryList = append(lanternEntryList, entry)
					found = true
				}
			})
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
