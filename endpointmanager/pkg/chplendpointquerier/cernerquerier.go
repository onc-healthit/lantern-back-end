package chplendpointquerier

import (
	"strings"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func CernerQuerier(chplURL string, fileToWriteTo string) {

	chplURL = strings.ReplaceAll(chplURL, "github.com", "raw.githubusercontent.com")
	chplURL = strings.Replace(chplURL, "/blob", "", 1)
	
	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(chplURL)
	if err != nil {
		log.Fatal(err)
	}

	endpointEntryList.Endpoints = BundleToLanternFormat(respBody)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
