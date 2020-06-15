package capabilityparser

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// base struct to handle any methods that don't change between the versions of FHIR
// capability statements
type baseParser struct {
	capStat map[string]interface{}
	version string
}

// GetPublisher returns the publisher field from the conformance/capability statement.
func (cp *baseParser) GetPublisher() (string, error) {
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

// GetFHIRVersion returns the FHIR version specifiedin the conformance/capability statement.
func (cp *baseParser) GetFHIRVersion() (string, error) {
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

// GetSoftwareName returns the software name specified in the conformance/capability statement.
func (cp *baseParser) GetSoftwareName() (string, error) {
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

// GetSoftwareVersion returns the software version specified in the conformance/capability statement.
func (cp *baseParser) GetSoftwareVersion() (string, error) {
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

// GetCopyright returns the copyright specified in the capability/conformance statement.
func (cp *baseParser) GetCopyright() (string, error) {
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

// Equal checks if the conformance/capability statement is equal to the given conformance/capability statement.
func (cp *baseParser) Equal(cs2 CapabilityStatement) bool {
	if cs2 == nil {
		return false
	}

	j1, err := cp.GetJSON()
	if err != nil {
		return false
	}
	j2, err := cs2.GetJSON()
	if err != nil {
		return false
	}
	if !bytes.Equal(j1, j2) {
		return false
	}

	return true
}

// GetJSON returns the JSON representation of the capability statement.
func (cp *baseParser) GetJSON() ([]byte, error) {
	return json.Marshal(cp.capStat)
}
