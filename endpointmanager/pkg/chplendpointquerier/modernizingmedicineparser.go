package chplendpointquerier

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func ModernizingMedicineQuerier(chplURL string, fileToWriteTo string) {

	emaURL := "https://fhir.m2qa." + chplURL
	gastroURL := "https://fhir.gastro." + chplURL

	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(emaURL)
	if err != nil {
		log.Fatal(err)
	}

	endpointEntryList.Endpoints = BundleToLanternFormat(respBody)

	respBody, err = helpers.QueryEndpointList(gastroURL)
	if err != nil {
		log.Fatal(err)
	}

	endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, BundleToLanternFormat(respBody)...)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}

}
