package chplendpointquerier

import (
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

type HealthSamuraiBundle struct {
	Entries []HealthSamuraiBundleEntry `json:"entry"`
}

type HealthSamuraiBundleEntry struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

func HealthSamuraiWebscraper(CHPLURL string, fileToWriteTo string) {

	var entry LanternEntry
	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(CHPLURL)
	if err != nil {
		log.Fatal(err)
	}

	var healthSamuraiBundle HealthSamuraiBundle
	err = json.Unmarshal(respBody, &healthSamuraiBundle)
	if err != nil {
		log.Fatal(err)
	}

	for _, bundleEntry := range healthSamuraiBundle.Entries {
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
