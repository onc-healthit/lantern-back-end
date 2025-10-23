package chplendpointquerier

import (
	"encoding/json"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

type Entry struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	LogoURL string `json:"logoUrl,omitempty"`
	URL     string `json:"url"`
}

type Bundle struct {
	ResourceType string  `json:"resourceType"`
	Type         string  `json:"type"`
	Total        int     `json:"total"`
	Entry        []Entry `json:"entry"`
	ID           string  `json:"id"`
}

func WillowQuerierParser(willowURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(willowURL)
	if err != nil {
		log.Fatal(err)
	}
	var willowArr Bundle
	err = json.Unmarshal(respBody, &willowArr)
	if err != nil {
		log.Fatal(err)
	}

	for _, willowEntry := range willowArr.Entry {
		var entry LanternEntry

		entry.URL = willowEntry.URL
		entry.OrganizationName = willowEntry.Name

		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList
	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
