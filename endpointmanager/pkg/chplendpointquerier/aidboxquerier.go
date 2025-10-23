package chplendpointquerier

import (
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func AidboxQuerierParser(aidboxURL string, fileToWriteTo string) {

	var lanternEntryList []LanternEntry
	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(aidboxURL)
	if err != nil {
		log.Fatal(err)
	}
	var aidboxArr Bundle
	err = json.Unmarshal(respBody, &aidboxArr)
	if err != nil {
		log.Fatal(err)
	}

	for _, aidboxEntry := range aidboxArr.Entry {
		var entry LanternEntry

		entry.URL = aidboxEntry.URL
		entry.OrganizationName = aidboxEntry.Name

		lanternEntryList = append(lanternEntryList, entry)
	}

	endpointEntryList.Endpoints = lanternEntryList
	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
