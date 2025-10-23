package chplendpointquerier

import (
	"encoding/json"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"strings"
)

func AzureWebsitesURLWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	jsonResponse, err := helpers.QueryEndpointList(chplURL)
	if err != nil {
		log.Fatal(err)
	}

	type response struct {
		BaseUrls []string `json:"baseUrls"`
	}

	azureJSON := response{}
	err = json.Unmarshal(jsonResponse, &azureJSON)
	if err != nil {
		log.Fatal(err)
	}

	endpoint := azureJSON.BaseUrls[0]
	var entry LanternEntry
	if endpoint != "" {
		entryURL := strings.TrimSpace(endpoint)
		entry.URL = entryURL

		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
