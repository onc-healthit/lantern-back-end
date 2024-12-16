package chplendpointquerier

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func ChntechsolutionsWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList
	var entry LanternEntry

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("p").Each(func(index int, phtml *goquery.Selection) {
		phtml.Find("code").Each(func(index int, phtml *goquery.Selection) {
			urlString := strings.ReplaceAll(phtml.Text(), "\n", " ")
			pattern := `https[^\s]*metadata`
			re := regexp.MustCompile(pattern)
			match := re.FindString(urlString)
			if len(match) != 0 {
				match = strings.TrimSpace(match)
				entry.URL = strings.TrimSuffix(match, "/metadata")
				lanternEntryList = append(lanternEntryList, entry)
				endpointEntryList.Endpoints = lanternEntryList
			}

		})
	})

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
