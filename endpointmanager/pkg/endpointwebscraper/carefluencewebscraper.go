package endpointwebscraper

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func Carefluenceebscraper(vendorURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc := helpers.ChromedpQueryEndpointList(vendorURL, ".main-content-inner")

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

	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
