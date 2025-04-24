package chplendpointquerier

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func ModernizingMedicineQuerier(chplURL string, fileToWriteTo string) {
	emaURL := "https://fhir.mmi.prod.fhir.ema-api.com/fhir/r4/Endpoint?connection-type=hl7-fhir-rest"
	gastroURL := "https://fhir.gastro.prod.fhir.ema-api.com/fhir/r4/Endpoint?connection-type=hl7-fhir-rest"
	exscribeURL := "https://ehrapi-exscribe-prod-fhir.ema-api.com/api/Endpoint"
	//traknetURL := "https://ehrapi-traknet-prod-fhir.ema-api.com/api/Endpoint"
	//sammyURL := "https://ehrapi-sammyehr-prod-fhir.ema-api.com/api/Endpoint"

	var endpointEntryList EndpointList
	respBody, err := helpers.QueryEndpointList(emaURL)
	if err != nil {
		log.Fatal(err)
	}
	endpointEntryList.Endpoints = BundleToLanternFormat(respBody, chplURL)

	respBody, err = helpers.QueryEndpointList(gastroURL)
	if err != nil {
		log.Fatal(err)
	}
	endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, BundleToLanternFormat(respBody, chplURL)...)

	respBody, err = helpers.QueryEndpointList(exscribeURL)
	if err != nil {
		log.Fatal(err)
	}
	endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, BundleToLanternFormat(respBody, chplURL)...)

	//respBody, err = helpers.QueryEndpointList(traknetURL)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, BundleToLanternFormat(respBody)...)

	// respBody, err = helpers.QueryEndpointList(sammyURL)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// endpointEntryList.Endpoints = append(endpointEntryList.Endpoints, BundleToLanternFormat(respBody)...)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Fatal(err)
	}
}
