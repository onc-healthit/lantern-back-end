package fetcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

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

// Source is an enum of the known endpoint source list urls
type Source string

// Cerner, Epic, Lantern & FHIR are fields in the Source enum
const (
	Cerner  Source = "CernerSource"
	Epic    Source = "EpicSource"
	Lantern Source = "LanternEndpointSourcesJson"
	FHIR    Source = "FHIRSource"
)

// Converts the string version of the endpoint source to the fetcher.Source enum
// This will eventually become unnecessary once we're pulling the data directly from the
// endpoint lists.
func checkSource(source string) Source {
	switch source {
	case "Cerner":
		return Cerner
	case "Epic":
		return Epic
	case "Lantern":
		return Lantern
	}
	return ""
}

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

	validSource := checkSource(source)
	if validSource != "" {
		return GetListOfEndpointsKnownSource([]byte(byteValue), validSource, listURL)
	}
	return GetListOfEndpoints([]byte(byteValue), source, listURL)
}

// GetListOfEndpointsKnownSource parses a list of endpoints out of a given byte array
func GetListOfEndpointsKnownSource(rawendpts []byte, source Source, listURL string) (ListOfEndpoints, error) {
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

	if source == Cerner {
		cernerList := initialList["endpoints"]
		if cernerList == nil {
			return result, fmt.Errorf("cerner list not given in Cerner format")
		}
		result = CernerList{}.GetEndpoints(cernerList, listURL)
	} else if source == Epic {
		epicList := initialList["Entries"]
		if epicList == nil {
			return result, fmt.Errorf("epic list not given in Epic format")
		}
		result = EpicList{}.GetEndpoints(epicList, listURL)
	} else if source == Lantern {
		lanternList := initialList["Endpoints"]
		if lanternList == nil {
			return result, fmt.Errorf("lantern list not given in Lantern format")
		}
		result = LanternList{}.GetEndpoints(lanternList, listURL)
		}
		result = LanternList{}.GetEndpoints(lanternList)
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
