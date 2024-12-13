package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func ZoobooksystemsWebscraper(CHPLURL string, fileToWriteTo string) error {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList
	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".col-lg-6.text-secondary.fw-bold")
	if err != nil {
		return err
	}

	inProduction := false

	doc.Find(".col-lg-6").Each(func(index int, divhtml *goquery.Selection) {
		if divhtml.Text() == "PRODUCTION" {
			inProduction = true
		}

		if inProduction {
			divhtml.Find("a").Each(func(indextr int, ahtml *goquery.Selection) {
				if strings.Contains(ahtml.Text(), "https") && !strings.Contains(ahtml.Text(), "oauth") {
					var entry LanternEntry

					entryURL := strings.TrimSpace(ahtml.Text())
					entry.URL = entryURL

					lanternEntryList = append(lanternEntryList, entry)

					endpointEntryList.Endpoints = lanternEntryList

					err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
					if err != nil {
						log.Fatal(err)
					}

					return
				}
			})
		}
	})

	return nil
}
