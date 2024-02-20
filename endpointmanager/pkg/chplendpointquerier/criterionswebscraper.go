package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func CriterionsWebscraper(CHPLURL string, fileToWriteTo string) {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".et_pb_text_inner")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".et_pb_text_inner").Each(func(index int, divhtml *goquery.Selection) {
		divhtml.Find("h3").Each(func(indextr int, h3html *goquery.Selection) {
			parts := strings.SplitN(h3html.Text(), "https://", 2)
			organization := parts[0]
			url := "https://" + parts[1]

			if strings.HasSuffix(organization, "-") {
				organization = strings.TrimSuffix(organization, "-")
			}

			var entry LanternEntry

			entry.URL = strings.TrimSpace(url)
			entry.OrganizationName = strings.TrimSpace(organization)
			lanternEntryList = append(lanternEntryList, entry)
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
