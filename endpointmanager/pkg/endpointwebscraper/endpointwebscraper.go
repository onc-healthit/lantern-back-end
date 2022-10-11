package endpointwebscraper

import (
	"encoding/json"
	"io/ioutil"
)

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
var techCareURL = "https://devportal.techcareehr.com/Serviceurls"
var carefluenceURL = "https://carefluence.com/carefluence-fhir-endpoints/"

func EndpointListWebscraper(vendorURL string, vendor string, fileToWriteTo string) {

	if vendorURL == oneUpURL || vendorURL == careEvolutionURL {
		HTMLtablewebscraper(vendorURL, vendor, fileToWriteTo)
	} else if vendorURL == techCareURL {
		Techcarewebscraper(vendorURL, fileToWriteTo)
	} else if vendorURL == carefluenceURL {
		Carefluenceebscraper(vendorURL, fileToWriteTo)
	}
}

// WriteEndpointListFile writes the given endpointEntryList to a json file and stores it in the prod resources directory
func WriteEndpointListFile(endpointEntryList EndpointList, fileToWriteTo string) error {
	finalFormatJSON, err := json.MarshalIndent(endpointEntryList, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("../../../resources/prod_resources/"+fileToWriteTo, finalFormatJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}
