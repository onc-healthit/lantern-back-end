package chplendpointquerier

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func BundleQuerierParser(CHPLURL string, fileToWriteTo string) error {

	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(CHPLURL)
	if err != nil {
		log.Info(err)
		return err
	}

	// convert bundle data to lantern format
	endpointEntryList.Endpoints = BundleToLanternFormat(respBody)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Info(err)
		return err
	}

	return nil
}
