package chplendpointquerier

import (
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func MedHostQuerier(medhostURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	respBody := helpers.QueryEndpointList(medhostURL)

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
			entry.URL = serviceBaseURL
		}

		developerName, ok := medhostEntry["facilityName"].(string)
		if ok {
			entry.OrganizationName = developerName
		}

		npiID, ok := medhostEntry["npi"].(string)
		if ok {
			entry.NPIID = npiID
		}

		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList
	WriteCHPLFile(endpointEntryList, fileToWriteTo)

}
