package fetcher

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// OrgKeyword is a struct for each keyword
type OrgKeyword struct {
	Kind  string
	Value string
}

// EndpointEntry is a struct for each entry of data pulled out of the EndpointSources file
type EndpointEntry struct {
	OrganizationName     string
	FHIRPatientFacingURI string
	ListSource           string
}

// ListOfEndpoints is a structure for the whole EndpointSources file
type ListOfEndpoints struct {
	Entries []EndpointEntry
}

// Source is an enum of the known endpoint source list urls
type Source string

// Cerner is a field in the Source enum for the cerner endpoint url
const (
	Cerner Source = "https://github.com/cerner/ignite-endpoints"
	Epic          = "https://open.epic.com/MyApps/EndpointsJson"
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
	}
	return ""
}

// Endpoints is an interface that every endpoint list can implement to parse their list into
// the universal format ListOfEndpoints
type Endpoints interface {
	GetEndpoints(map[string]interface{}) ListOfEndpoints
}

// GetListOfEndpoints parses a list of endpoints out of the file at the provided path
func GetListOfEndpoints(filePath string) (ListOfEndpoints, error) {
	var result ListOfEndpoints
	var initialList map[string][]map[string]interface{}

	jsonFile, err := os.Open(filePath)
	// If we os.Open returns an error then handle it
	if err != nil {
		return result, err
	}
	// Defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal([]byte(byteValue), &initialList)

	// return nil if an empty endpoint list was passed in
	if initialList == nil {
		return result, nil
	}

	if err != nil {
		return result, err
	}

	finalResult, err := formatList(initialList)

	return finalResult, err
}

func formatList(initialList map[string][]map[string]interface{}) (ListOfEndpoints, error) {
	var endptList ListOfEndpoints
	var errs error

	// Cerner's top-level JSON field is "endpoints"
	cernerList, ok := initialList["endpoints"]
	if ok {
		endptList, errs = CernerList{}.GetEndpoints(cernerList)
	}

	// Everything else should have a top level JSON field of "Entries"
	defaultList, ok := initialList["Entries"]
	if ok {
		endptList, errs = DefaultList{}.GetEndpoints(defaultList)
	}

	return endptList, errs
}
