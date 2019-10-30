package fetcher

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Schema for data pulled out of EndpointSources file
type ListOfEndpoints struct {
	Entries []struct {
		OrganizationName     string `json:"OrganizationName"`
		FHIRPatientFacingURI string `json:"FHIRPatientFacingURI"`
		Type                 string `json:"Type"`
		Keywords             []struct {
			Kind  string `json:"Kind"`
			Value string `json:"Value"`
		} `json:"Keywords"`
	} `json:"Entries"`
}

// GetListOfEndpoints parsers a list of endpoints out of the file at the provided path
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
