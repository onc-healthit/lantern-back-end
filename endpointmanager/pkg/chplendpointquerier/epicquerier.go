package chplendpointquerier

import (
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func EpicQuerier(epicURL string, fileToWriteTo string) {

	DSTU2URL := strings.Join(strings.Split(epicURL, "/")[:3], "/") + "/Endpoints/DSTU2"
	R4URL := strings.Join(strings.Split(epicURL, "/")[:3], "/") + "/Endpoints/R4"

	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(DSTU2URL)
	if err != nil {
		log.Fatal(err)
	}

	endpointEntryList.Endpoints = BundleToLanternFormat(respBody, epicURL)

	respBody, err = helpers.QueryEndpointList(R4URL)
	if err != nil {
		log.Fatal(err)
	}

	endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, BundleToLanternFormat(respBody, epicURL)...)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
