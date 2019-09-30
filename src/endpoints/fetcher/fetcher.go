package fetcher

import (
	"encoding/json"
	"fmt"
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

func GetListOfEndpoints(filePath string) ListOfEndpoints {
	jsonFile, err := os.Open(filePath)
	// If we os.Open returns an error then handle it
	if err != nil {
		// TODO: Use a logging solution instead of println
		fmt.Println(err.Error())
	}
	// Defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result ListOfEndpoints
	err = json.Unmarshal([]byte(byteValue), &result)
	if err != nil {
		// TODO: Use a logging solution instead of println
		println("Endpoint List Parsing Error: ", err.Error())
	}

	return result
}
