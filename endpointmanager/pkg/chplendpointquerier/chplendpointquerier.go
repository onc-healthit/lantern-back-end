package chplendpointquerier

type EndpointList struct {
	Endpoints []LanternEntry `json:"Endpoints"`
}

type LanternEntry struct {
	URL              string `json:"URL"`
	OrganizationName string `json:"OrganizationName"`
	NPIID            string `json:"NPIID"`
}

func QueryCHPLEndpointList(chplURL string, fileToWriteTo string) {

	if chplURL == "https://api.mhdi10xasayd.com/medhost-developer-composition/v1/fhir-base-urls.json" {
		MedHostQuerier(chplURL, fileToWriteTo)
	} else if chplURL == "https://nextgen.com/api/practice-search" {
		CHPLwebscraper(chplURL, fileToWriteTo)
	}
}
