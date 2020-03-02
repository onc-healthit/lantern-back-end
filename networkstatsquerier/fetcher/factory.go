package fetcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// @TODO Remove commented things
// from https://www.hl7.org/fhir/codesystem-FHIR-version.html
// looking at official and release versions only
// var dstu2 = []string{"1.0.1", "1.0.2"}
// var stu3 = []string{"3.0.0", "3.0.1"}
// var r4 = []string{"4.0.0", "4.0.1"}

// Endpoints does things @TODO
type Endpoints interface {
	// @TODO Rename this
	GetEndpointList() (string, error)
	// GetFHIRVersion() (string, error)
	// GetSoftwareName() (string, error)
	// GetSoftwareVersion() (string, error)
	// GetCopyright() (string, error)

	// Equal(CapabilityStatement) bool
	// GetJSON() ([]byte, error)
}

// NewEndpointList does things @TODO
func NewEndpointList(filePath string) (Endpoints, error) {
	var result map[string]interface{}

	jsonFile, err := os.Open(filePath)
	// If we os.Open returns an error then handle it
	if err != nil {
		return result, err
	}
	// Defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal([]byte(byteValue), &result)

	// return nil if an empty capability statement was passed in
	if result == nil {
		return nil, nil
	}

	// DSTU2, STU3, R4 all have fhirVersion in same location
	fhirVersion, ok := capStat["fhirVersion"].(string)
	if !ok {
		return nil, errors.New("unable to parse fhir version from capability/conformance statement")
	}

	if contains(dstu2, fhirVersion) {
		return newDSTU2(capStat), nil
	} else if contains(stu3, fhirVersion) {
		return newSTU3(capStat), nil
	} else if contains(r4, fhirVersion) {
		return newR4(capStat), nil
	}

	return nil, fmt.Errorf("unknown FHIR version %s", fhirVersion)
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
