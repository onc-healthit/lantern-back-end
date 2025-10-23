package chplendpointquerier

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func EhealthlineWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	ipPattern := `\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}(:\d{1,5})?\b`
	ipRegex := regexp.MustCompile(ipPattern)

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("td:contains('https')").Each(func(index int, item *goquery.Selection) {
		if !ipRegex.MatchString(item.Text()) {
			var entry LanternEntry
			entry.URL = strings.TrimSpace(item.Find("span").Text())
			lanternEntryList = append(lanternEntryList, entry)
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
