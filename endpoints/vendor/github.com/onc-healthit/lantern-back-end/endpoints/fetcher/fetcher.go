package fetcher

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Keyword is a struct for each keyword
type OrgKeyword struct {
	Kind  string `json:"Kind"`
	Value string `json:"Value"`
}

// EndpointEntry is a struct for each entry of data pulled out of the EndpointSources file
type EndpointEntry struct {
	OrganizationName     string       `json:"OrganizationName"`
	FHIRPatientFacingURI string       `json:"FHIRPatientFacingURI"`
	Type                 string       `json:"Type"`
	Keywords             []OrgKeyword `json:"Keywords"`
}

// ListOfEndpoints is a structure for the whole EndpointSources file
type ListOfEndpoints struct {
	Entries []EndpointEntry `json:"Entries"`
}

// GetListOfEndpoints parses a list of endpoints out of the file at the provided path
func GetListOfEndpoints(filePath string) (ListOfEndpoints, error) {
	var result ListOfEndpoints

	jsonFile, err := os.Open(filePath)
	// If we os.Open returns an error then handle it
	if err != nil {
		return result, err
	}
	// Defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal([]byte(byteValue), &result)

	return result, err
}
