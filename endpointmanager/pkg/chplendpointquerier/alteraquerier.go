package chplendpointquerier

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func AlteraQuerier(chplURL string, fileToWriteTo string) {
	DSTU2URL :=  chplURL + "/download/DSTU2"
	R4URL := chplURL + "/download/R4"

	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(DSTU2URL)
	if err != nil {
		log.Fatal(err)
	}

	endpointEntryList.Endpoints = BundleToLanternFormat(respBody)

	respBody, err = helpers.QueryEndpointList(R4URL)
	if err != nil {
		log.Fatal(err)
	}

	endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, BundleToLanternFormat(respBody)...)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
