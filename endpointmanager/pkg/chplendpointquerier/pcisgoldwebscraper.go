package chplendpointquerier

import (
	"encoding/json"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"strings"
)

func PCISgoldURLWebscraper(chplURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	jsonResponse, err := helpers.QueryEndpointList(chplURL)
	if err != nil {
		log.Fatal(err)
	}

	var qualifactsJSON []map[string]string
	err = json.Unmarshal(jsonResponse, &qualifactsJSON)
	if err != nil {
		log.Fatal(err)
	}

	for _, endpointEntry := range qualifactsJSON {

		var entry LanternEntry

		orgName, ok := endpointEntry["Name"]
		if !ok {
			log.Fatal(err)
		}

		URL, ok := endpointEntry["FhirUrl"]
		if !ok {
			log.Fatal(err)
		}

		if URL != "" {
			entryURL := strings.TrimSpace(URL)
			entry.URL = entryURL
			entry.OrganizationName = orgName

			lanternEntryList = append(lanternEntryList, entry)
		}
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
