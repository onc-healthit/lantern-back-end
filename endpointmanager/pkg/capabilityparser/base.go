package capabilityparser

import (
	"fmt"
)

// base struct to handle any methods that don't change between the versions of of FHIR
// capability statements
type baseParser struct {
	capStat map[string]interface{}
	version string
}

func newBase(capStat map[string]interface{}, version string) *baseParser {
	return &baseParser{
		capStat: capStat,
		version: version,
	}
}

func (cp baseParser) GetPublisher() (string, error) {
	publisher := cp.capStat["publisher"]
	if publisher == nil {
		return "", nil
	}
	publisherStr, ok := publisher.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement publisher value to a string", cp.version)
	}
	return publisherStr, nil
}

func (cp baseParser) GetFHIRVersion() (string, error) {
	fhirVersion := cp.capStat["fhirVersion"]
	if fhirVersion == nil {
		return "", nil
	}
	fhirVersionStr, ok := fhirVersion.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement fhirVersion value to a string", cp.version)
	}
	return fhirVersionStr, nil
}

func (cp baseParser) GetSoftwareName() (string, error) {
	software := cp.capStat["software"]
	if software == nil {
		return "", nil
	}
	softwareMap, ok := software.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement software value to a map[string]interface{}", cp.version)
	}
	name := softwareMap["name"]
	if name == nil {
		return "", nil
	}
	nameStr, ok := name.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement software.name value to a string", cp.version)
	}
	return nameStr, nil
}

func (cp baseParser) GetSoftwareVersion() (string, error) {
	software := cp.capStat["software"]
	if software == nil {
		return "", nil
	}
	softwareMap, ok := software.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement software value to a map[string]interface{}", cp.version)
	}
	version := softwareMap["version"]
	if version == nil {
		return "", nil
	}
	versionStr, ok := version.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement software.version value to a string", cp.version)
	}
	return versionStr, nil
}

func (cp baseParser) GetCopyright() (string, error) {
	copyright := cp.capStat["copyright"]
	if copyright == nil {
		return "", nil
	}
	copyrightStr, ok := copyright.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement copyright value to a string", cp.version)
	}
	return copyrightStr, nil
}
