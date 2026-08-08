package chplendpointquerier

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func CernerBundleParser(CHPLURL string, fileToWriteTo string) {

	milleniumR4URL := "https://raw.githubusercontent.com/cerner/ignite-endpoints/main/millennium_patient_r4_endpoints.json"
	milleniumDSTU2URL := "https://raw.githubusercontent.com/cerner/ignite-endpoints/main/millennium_patient_dstu2_endpoints.json"

	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(milleniumR4URL)
	if err != nil {
		log.Fatal(err)
	}

	// convert bundle data to lantern format
	milleniumR4Entries, _, _ := BundleToLanternFormat(respBody, CHPLURL)
	endpointEntryList.Endpoints = milleniumR4Entries

	respBody, err = helpers.QueryEndpointList(milleniumDSTU2URL)
	if err != nil {
		log.Fatal(err)
	}

	milleniumDSTU2Entries, _, _ := BundleToLanternFormat(respBody, CHPLURL)
	endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, milleniumDSTU2Entries...)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
