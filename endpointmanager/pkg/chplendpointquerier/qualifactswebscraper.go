package chplendpointquerier

import (
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

func QualifactsWebscraper(chplURL string, fileToWriteTo string) {

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

		endpointType, ok := endpointEntry["organization"]
		if !ok {
			log.Fatal(err)
		}
		
		if strings.Contains(endpointType, "Production Organizations") {
			URL, ok := endpointEntry["baseurl"]
			if !ok {
				log.Fatal(err)
			}
			
			if URL != "" {
				entryURL := strings.TrimSpace(URL)
				entry.URL = entryURL

				lanternEntryList = append(lanternEntryList, entry)
			}
		}
	}

	endpointEntryList.Endpoints = lanternEntryList

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}