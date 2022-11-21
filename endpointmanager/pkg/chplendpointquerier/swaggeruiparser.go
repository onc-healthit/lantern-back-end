package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"regexp"
)

func SwaggerUIWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "pre")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("pre").Each(func(index int, baseURL *goquery.Selection) {
		var entry LanternEntry

		urlInfoText := baseURL.Text()

		if strings.Contains(urlInfoText, "Base URL:") {

			re, err := regexp.Compile(`(\[|\]|Base URL:)`)
			if err != nil {
				log.Fatal(err)
			}

			url := re.ReplaceAllString(urlInfoText, "")

			entryURL := strings.TrimSpace(url)
			entry.URL = entryURL

			lanternEntryList = append(lanternEntryList, entry)
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
