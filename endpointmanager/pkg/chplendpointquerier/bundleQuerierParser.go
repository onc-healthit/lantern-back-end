package chplendpointquerier

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func BundleQuerierParser(CHPLURL string, fileToWriteTo string) {

	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(CHPLURL)
	if err != nil {
		log.Info("Error for the URL: ", CHPLURL)
		log.Fatal(err)
	}

	// convert bundle data to lantern format
	endpointEntryList.Endpoints = BundleToLanternFormat(respBody, CHPLURL)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Info("Error for the URL: ", CHPLURL)
		log.Fatal(err)
	}
}
