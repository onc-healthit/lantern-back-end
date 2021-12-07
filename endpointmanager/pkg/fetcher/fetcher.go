package fetcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/pkg/errors"
)

// OrgKeyword is a struct for each keyword
type OrgKeyword struct {
	Kind  string
	Value string
}

// EndpointEntry is a struct for each entry of data pulled out of the EndpointSources file
type EndpointEntry struct {
	OrganizationNames    []string
	NPIIDs               []string
	FHIRPatientFacingURI string
	ListSource           string
}

// ListOfEndpoints is a structure for the whole EndpointSources file
type ListOfEndpoints struct {
	Entries []EndpointEntry
}

// Source is a slice of the known endpoint source lists
var sources = []string{"Cerner", "Epic", "Lantern", "CareEvolution", "1Up", "FHIR"}

// Endpoints is an interface that every endpoint list can implement to parse their list into
// the universal format ListOfEndpoints
type Endpoints interface {
	GetEndpoints(map[string]interface{}, string) ListOfEndpoints
}

// GetEndpointsFromFilepath parses a list of endpoints out of the file at the provided path
func GetEndpointsFromFilepath(filePath string, source string, listURL string) (ListOfEndpoints, error) {
	jsonFile, err := os.Open(filePath)
	// If we os.Open returns an error then handle it
	if err != nil {
		return ListOfEndpoints{}, err
	}
	// Defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	if len(byteValue) == 0 {
		return ListOfEndpoints{}, nil
	}

	validSource := helpers.StringArrayContains(sources, source)
	if validSource {
		return GetListOfEndpointsKnownSource([]byte(byteValue), source, listURL)
	}
	return GetListOfEndpoints([]byte(byteValue), source, listURL)
}

// GetListOfEndpointsKnownSource parses a list of endpoints out of a given byte array
func GetListOfEndpointsKnownSource(rawendpts []byte, source string, listURL string) (ListOfEndpoints, error) {
	var result ListOfEndpoints
	var initialList map[string]interface{}

	err := json.Unmarshal(rawendpts, &initialList)

	if err != nil {
		return result, err
	}

	// return nil if null or {} was passed in as the rawendpts byte array
	if len(initialList) == 0 {
		return result, nil
	}

	if source == "Cerner" {
		cernerList, err := convertInterfaceToList(initialList, "endpoints")
		if err != nil {
			return result, fmt.Errorf("cerner list not given in Cerner format: %s", err)
		}
		result = CernerList{}.GetEndpoints(cernerList, listURL)
	} else if source == "Epic" {
		epicList, err := convertInterfaceToList(initialList, "Entries")
		if err != nil {
			return result, fmt.Errorf("epic list not given in EPIC format: %s", err)
		}
		result = EpicList{}.GetEndpoints(epicList, listURL)
	} else if source == "Lantern" {
		lanternList, err := convertInterfaceToList(initialList, "Endpoints")
		if err != nil {
			return result, fmt.Errorf("lantern list not given in Lantern format: %s", err)
		}
		result = LanternList{}.GetEndpoints(lanternList, listURL)
	} else if source == "CareEvolution" {
		careEvolutionList, err := convertInterfaceToList(initialList, "Endpoints")
		if err != nil {
			return result, fmt.Errorf("CareEvolution list not given in CareEvolution format: %s", err)
		}
		result = CareEvolutionList{}.GetEndpoints(careEvolutionList, listURL)
	} else if source == "1Up" {
		oneUpList, err := convertInterfaceToList(initialList, "Endpoints")
		if err != nil {
			return result, fmt.Errorf("1Up list not given in 1Up format: %s", err)
		}
		result = OneUpList{}.GetEndpoints(oneUpList, listURL)
	} else if source == "FHIR" {
		// based on: https://www.hl7.org/fhir/endpoint-examples-general-template.json.html
		fhirList, err := convertInterfaceToList(initialList, "entry")
		if err != nil {
			return result, fmt.Errorf("fhir list not given in FHIR format: %s", err)
		}
		result = FHIRList{}.GetEndpoints(fhirList, listURL)
	} else {
		return result, fmt.Errorf("no endpoint list parser implemented for the given source")
	}

	return result, err
}

// GetListOfEndpoints parses a list of endpoints out of a given byte array
func GetListOfEndpoints(rawendpts []byte, source string, listURL string) (ListOfEndpoints, error) {
	var result ListOfEndpoints
	var initialList map[string][]map[string]interface{}

	err := json.Unmarshal(rawendpts, &initialList)
	if err != nil {
		return result, errors.Wrap(err,
			`provided endpoint list was not formatted as expected.
			See 'Expected Endpoint Source Formatting' in the Endpoint Manager README.`)
	}

	// return nil if null or {} was passed in as the rawendpts byte array
	if len(initialList) == 0 {
		return result, nil
	}

	defaultList, ok := initialList["Entries"]
	if !ok {
		return result, fmt.Errorf(`the given endpoint list is not formatted in the default format,
			see 'Expected Endpoint Source Formatting' in the Endpoint Manager README`)
	}
	result = getDefaultEndpoints(defaultList, source, listURL)

	return result, err
}

func convertInterfaceToList(list map[string]interface{}, ref string) ([]map[string]interface{}, error) {
	var formatList []map[string]interface{}

	endptList := list[ref]
	if endptList == nil {
		return formatList, fmt.Errorf("incorrect reference value")
	}

	intList, ok := endptList.([]interface{})
	if !ok {
		return formatList, fmt.Errorf("endpoint list is not an array")
	}

	for i := range intList {
		elem, ok := intList[i].(map[string]interface{})
		if !ok {
			return formatList, fmt.Errorf("list element is not map[string]interface{}")
		}
		formatList = append(formatList, elem)
	}
	return formatList, nil
}
