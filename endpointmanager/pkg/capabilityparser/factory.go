package capabilityparser

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
)

type CapabilityStatement interface {
	GetPublisher() (string, error)
	GetFHIRVersion() (string, error)
	GetSoftwareName() (string, error)
	GetSoftwareVersion() (string, error)
}

func NewCapabilityStatement(capJSON []byte) (CapabilityStatement, error) {
	var err error
	var capStat map[string]interface{}

	err = json.Unmarshal(capJSON, &capStat)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling JSON capability statement")
	}

	// DSTU2, STU3, R4 all have fhirVersion in same location
	fhirVersion := capStat["fhirVersion"].(string)

	// DSTU2 always 1.0.2
	// STU3 can be 3.x.x
	// R4 can be 4.x.x
	stu3Regex := regexp.MustCompile(`^3\.[0-9]+\.[0-9]+$`)
	r4Regex := regexp.MustCompile(`^4\.[0-9]+\.[0-9]+$`)

	if fhirVersion == "1.0.2" {
		return newDSTU2(capStat), nil
	} else if stu3Regex.MatchString(fhirVersion) {
		return newSTU3(capStat), nil
	} else if r4Regex.MatchString(fhirVersion) {
		return newR4(capStat), nil
	}

	return nil, fmt.Errorf("unknown FHIR version %s", fhirVersion)
}
