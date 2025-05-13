package chplendpointquerier

import (
	"reflect"

	"github.com/PuerkitoBio/goquery"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
)

func entryExists(lanternEntryList []LanternEntry, lanternEntry LanternEntry) bool {
	for _, entry := range lanternEntryList {
		if reflect.DeepEqual(entry, lanternEntry) {
			return true
		}
	}
	return false
}

func GuidewellURLWebscraper(CHPLURL string, fileToWriteTo string) error {
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	doc, err := helpers.ChromedpQueryEndpointList(CHPLURL, "div.apiEndpointUrl")
	if err != nil {
		return err
	}
	doc.Find("div.apiEndpointUrl").Each(func(index int, urlElements *goquery.Selection) {

		var lanternEntry LanternEntry

		fhirURL := urlElements.Text()
		lanternEntry.URL = fhirURL
		if !entryExists(lanternEntryList, lanternEntry) {
			lanternEntryList = append(lanternEntryList, lanternEntry)
		}
	})

	endpointEntryList.Endpoints = lanternEntryList
	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		return err
	}

	return nil
}
