package chplendpointquerier

import (
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"strings"
)

func MedHostQuerier(medhostURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(medhostURL)
	if err != nil {
		log.Fatal(err)
	}

	var medhostArr []map[string]interface{}
	err = json.Unmarshal(respBody, &medhostArr)
	if err != nil {
		log.Fatal(err)
	}

	for _, medhostEntry := range medhostArr {
		var entry LanternEntry

		serviceBaseURL, ok := medhostEntry["serviceBaseUrl"].(string)
		if !ok {
			log.Fatal("Error converting serviceBaseUrl to type string")
		} else {
			entry.URL = strings.TrimSpace(serviceBaseURL)
		}

		developerName, ok := medhostEntry["facilityName"].(string)
		if ok {
			entry.OrganizationName = strings.TrimSpace(developerName)
		}

		npiID, ok := medhostEntry["npi"].(string)
		if ok {
			entry.NPIID = strings.TrimSpace(npiID)
		}

		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList
	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
