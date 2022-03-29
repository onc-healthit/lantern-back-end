package chplendpointquerier

type EndpointList struct {
	Endpoints []LanternEntry `json:"Endpoints"`
}

type LanternEntry struct {
	URL              string `json:"URL"`
	OrganizationName string `json:"OrganizationName"`
	NPIID            string `json:"NPIID"`
}

var MedHostURL = "https://api.mhdi10xasayd.com/medhost-developer-composition/v1/fhir-base-urls.json"
var NextGenURL = "https://nextgen.com/api/practice-search"

func QueryCHPLEndpointList(chplURL string, fileToWriteTo string) {

	if chplURL == MedHostURL {
		MedHostQuerier(chplURL, fileToWriteTo)
	} else if chplURL == NextGenURL {
		CHPLwebscraper(chplURL, fileToWriteTo)
	}
}
