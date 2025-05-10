package chplendpointquerier

import (
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func eClinicalWorksBundleParser(CHPLURL string, fileToWriteTo string) {
	bundleFilePath := "practiceList.json"

	var endpointEntryList EndpointList

	respBodyJSON, err := helpers.QueryAndReadFile(CHPLURL, bundleFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// convert bundle data to lantern format
	bundleLanternFormat := BundleToLanternFormat(respBodyJSON, CHPLURL)

	endpointEntryList.Endpoints = bundleLanternFormat

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove(bundleFilePath)
	if err != nil {
		log.Fatal(err)
	}
}
