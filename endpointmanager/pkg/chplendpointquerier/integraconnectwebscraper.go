package chplendpointquerier

import (
	"encoding/json"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func IntegraConnectWebscraper(CHPLURL string, fileToWriteTo string) {

	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(CHPLURL)
	if err != nil {
		log.Fatal(err)
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(respBody), &data)
	if err != nil {
		log.Println("Error unmarshaling JSON:", err)
		return
	}

	entries := data["entry"].([]interface{})

	var filteredEntries []map[string]interface{}

	// Filter the entries based on the resourceType set to Organization or Endpoint
	for _, entry := range entries {
		resource := entry.(map[string]interface{})["resource"].(map[string]interface{})
		if strings.EqualFold(strings.TrimSpace(resource["resourceType"].(string)), "Organization") ||
			strings.EqualFold(strings.TrimSpace(resource["resourceType"].(string)), "Endpoint") {
			filteredEntries = append(filteredEntries, entry.(map[string]interface{}))
		}
	}

	// Update the data with the filtered entries
	data["entry"] = filteredEntries

	// Marshal the data into json file
	newJsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}

	// convert bundle data to lantern format
	endpointEntryList.Endpoints = BundleToLanternFormat(newJsonData, CHPLURL)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
