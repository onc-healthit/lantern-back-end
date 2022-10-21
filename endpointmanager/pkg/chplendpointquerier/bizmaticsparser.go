package chplendpointquerier

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func BizmaticsBundleParser(CHPLURL string, fileToWriteTo string) {

	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(CHPLURL)
	if err != nil {
		log.Fatal(err)
	}

	// convert bundle data to lantern format
	endpointEntryList.Endpoints = BundleToLanternFormat(respBody)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
