package chplendpointquerier

import (
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
)

func EpicQuerier(epicURL string, fileToWriteTo string) {

	DSTU2URL := strings.Join(strings.Split(epicURL, "/")[:3], "/") + "/Endpoints/DSTU2"
	R4URL := strings.Join(strings.Split(epicURL, "/")[:3], "/") + "/Endpoints/R4"

	var endpointEntryList EndpointList

	respBody := helpers.QueryEndpointList(DSTU2URL)

	endpointEntryList.Endpoints = BundleToLanternFormat(respBody)

	respBody = helpers.QueryEndpointList(R4URL)

	endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, BundleToLanternFormat(respBody)...)

	WriteCHPLFile(endpointEntryList, fileToWriteTo)

}
