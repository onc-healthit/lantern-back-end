package chplendpointquerier

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func HumanaURLWebscraper(CHPLURL string, fileToWriteTo string) {


	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "p")
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("p").Each(func(index int, pElement *goquery.Selection) {
		pElement.Find("label").Each(func(index int, labelElement *goquery.Selection){
			if (strings.Contains(labelElement.Text(), "https")){
				url := labelElement.Text()
				url = strings.TrimSuffix(url, "/Patient")
				var entry LanternEntry
						entry.URL = url
						lanternEntryList = append(lanternEntryList, entry)
			}
		})
	})
	
	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
