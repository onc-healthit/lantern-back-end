package chplendpointquerier

import (
	"encoding/json"
	"fmt"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func FootholdURLQuerierParser(footholdURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(footholdURL)
	if err != nil {
		log.Fatal(err)
	}

	data := make(map[string]string)

	// Unmarshal the JSON data into the map
	err = json.Unmarshal(respBody, &data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for key, value := range data {
		var entry LanternEntry

		entry.URL = value
		entry.OrganizationName = key

		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList
	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
