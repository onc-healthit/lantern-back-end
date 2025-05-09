package chplendpointquerier

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func CapellaEHRBundleParser(CHPLURL string, fileToWriteTo string) {
	var endpointEntryList EndpointList

	// Use the TLS-skipping option specifically for this domain
	respBody, err := helpers.QueryEndpointListWithTLSOption(CHPLURL, true)
	if err != nil {
		log.Info("Error for the URL even with TLS verification disabled: ", CHPLURL)
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
