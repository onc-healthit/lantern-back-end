package chplendpointquerier

func QueryCHPLEndpointList(chplURL string, fileToWriteTo string) {

	if chplURL == "https://api.mhdi10xasayd.com/medhost-developer-composition/v1/fhir-base-urls.json" {
		MedHostQuerier(chplURL, fileToWriteTo)
	}
}
