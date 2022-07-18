package endpointwebscraper

type EndpointList struct {
	Endpoints []LanternEntry `json:"Endpoints"`
}

type LanternEntry struct {
	URL              string `json:"URL"`
	OrganizationName string `json:"OrganizationName"`
	NPIID            string `json:"NPIID"`
}

var oneUpURL = "https://1up.health/fhir-endpoint-directory"
var careEvolutionURL = "https://fhir.docs.careevolution.com/overview/public_endpoints.html"
var athenaHealthURL = "https://mydata.athenahealth.com/aserver"
var techCareURL = "https://devportal.techcareehr.com/Serviceurls"
var carefluenceURL = "https://carefluence.com/carefluence-fhir-endpoints/"

func EndpointListWebscraper(vendorURL string, vendor string, fileToWriteTo string) {

	if vendorURL == oneUpURL || vendorURL == careEvolutionURL {
		HTMLtablewebscraper(vendorURL, vendor, fileToWriteTo)
	} else if vendorURL == athenaHealthURL {
		Athenawebscraper(vendorURL, fileToWriteTo)
	} else if vendorURL == techCareURL {
		Techcarewebscraper(vendorURL, fileToWriteTo)
	} else if vendorURL == carefluenceURL {
		Carefluenceebscraper(vendorURL, fileToWriteTo)
	}
}
