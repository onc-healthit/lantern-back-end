package chplendpointquerier

import (
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

type CustomBundle struct {
	Entries []CustomBundleEntry `json:"entry"`
}

type CustomBundleEntry struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

func CustomBundleQuerierParser(CHPLURL string, fileToWriteTo string) {

	var entry LanternEntry
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(CHPLURL)
	if err != nil {
		log.Fatal(err)
	}

	var customBundle CustomBundle
	err = json.Unmarshal(respBody, &customBundle)
	if err != nil {
		log.Fatal(err)
	}

	for _, bundleEntry := range customBundle.Entries {
		entry.URL = bundleEntry.Url
		entry.OrganizationName = bundleEntry.Name

		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
