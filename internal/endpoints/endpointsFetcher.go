package endpoints

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result ListOfEndpoints
	json.Unmarshal([]byte(byteValue), &result)

	return result
}

// Prometheus doesn't allow special characters in the namespace, strip them out
func NamespaceifyString(name string) string {
	var nameString = strings.Replace(name, " ", "", -1)
	nameString = strings.Replace(nameString, "-", "", -1)
	nameString = strings.Replace(nameString, "–", "", -1)
	nameString = strings.Replace(nameString, "_", "", -1)
	nameString = strings.Replace(nameString, "&", "", -1)
	nameString = strings.Replace(nameString, "(", "", -1)
	nameString = strings.Replace(nameString, ")", "", -1)
	nameString = strings.Replace(nameString, ".", "", -1)
	nameString = strings.Replace(nameString, ",", "", -1)
	nameString = strings.Replace(nameString, "/", "", -1)
	nameString = strings.Replace(nameString, "+", "", -1)
	nameString = strings.Replace(nameString, "'", "", -1)
	nameString = strings.Replace(nameString, "’", "", -1)
	return strings.Replace(nameString, "'", "", -1)
}
