package endpointwebscraper

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func Carefluenceebscraper(vendorURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(vendorURL, ".main-content-inner")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".main-content-inner").Each(func(index int, mainContent *goquery.Selection) {
		mainContent.Find("p").Each(func(indextr int, phtml *goquery.Selection) {
			// Only the first two entries are production server endpoints
			var entry LanternEntry

			fhirURL := strings.TrimSpace(phtml.Text())
			entry.URL = fhirURL
			lanternEntryList = append(lanternEntryList, entry)
		})
	})

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteEndpointListFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
