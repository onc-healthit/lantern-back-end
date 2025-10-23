package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func SmileCdrWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "")
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("pre code.language-properties").Each(func(i int, s *goquery.Selection) {
		configText := s.Text()
		lines := strings.Split(configText, "\n")
		for _, line := range lines {
			if strings.Contains(line, "module.fhir_endpoint.config.base_url.fixed") {
				parts := strings.Split(line, "=")
				if len(parts) == 2 {
					url := strings.TrimSpace(parts[1])
					var entry LanternEntry

					entryURL := strings.TrimSpace(url)
					entry.URL = entryURL
					lanternEntryList = append(lanternEntryList, entry)
				}
			}
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
