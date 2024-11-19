package chplendpointquerier

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func DrcloudehrWebscraper(CHPLURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "")
	if err != nil {
		log.Fatal(err)
	}
	var entry LanternEntry
	doc.Find("td").Each(func(index int, tableElems *goquery.Selection) {
		fmt.Println(tableElems.Text())
		if strings.Contains(tableElems.Text(), "https") {
			fhirURL := strings.TrimSpace(tableElems.Text())
			entry.URL = fhirURL
			lanternEntryList = append(lanternEntryList, entry)
		} else {
			entry.OrganizationName = strings.TrimSpace(tableElems.Text())
		}
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
