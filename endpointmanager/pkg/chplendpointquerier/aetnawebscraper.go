package chplendpointquerier

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func AetnaURLWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, ".base-url")
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("hgroup").Each(func(index int, hElems *goquery.Selection) {
		doc.Find("pre").Each(func(index int, pElems *goquery.Selection) {
			if strings.Contains(pElems.Text(), "Base URL:") {
				re := regexp.MustCompile(`(?m)Base URL: (\S+)`)
				// Find the submatches
				matches := re.FindStringSubmatch(pElems.Text())
				// Check if any matches were found
				if len(matches) >= 2 {
					url := matches[1]
					url = "https://" + url + "/patientaccess"
					var entry LanternEntry
					entryURL := url
					entry.URL = entryURL
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
