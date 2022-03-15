package chplendpointquerier

func queryCHPLEndpointList(chplURL string, fileToWriteTo string) {

	if chplURL == "https://www.mphrx.com/fhir-service-base-url-directory" {
		queryCHPLEndpointList(chplURL, fileToWriteTo)
	}
}
