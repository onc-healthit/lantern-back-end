package chplendpointquerier

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"strings"
)

func AspMDeWebscraper(chplURL string, fileToWriteTo string) {
	found := false

	baseURL := strings.TrimSuffix(chplURL, "/fhir_aspmd.asp#apiendpoints")
	baseURL = strings.TrimSuffix(baseURL, "/fhir_aspmd.asp")

	doc, err := helpers.ChromedpQueryEndpointList(chplURL, "p")
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
				if exists && strings.Contains(href, "endpoints.asp") {
					bundleURL := baseURL + "/" + href
					found = true
					BundleQuerierParser(bundleURL, fileToWriteTo)
				}
			})
		}
	})
}
