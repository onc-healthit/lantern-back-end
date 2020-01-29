package capabilityparser

import "errors"

type dstu2CapabilityParser struct {
	capStat map[string]interface{}
}

func newDSTU2(capStat map[string]interface{}) CapabilityStatement {
	return &dstu2CapabilityParser{
		capStat: capStat,
	}
}

func (cp dstu2CapabilityParser) GetPublisher() (string, error) {
	publisher := cp.capStat["publisher"]
	if publisher == nil {
		return "", nil
	}
	publisherStr, ok := publisher.(string)
	if !ok {
		return "", errors.New("unable to cast DSTU2 capability statement publisher value to a string")
	}
	return publisherStr, nil
}

func (cp dstu2CapabilityParser) GetFHIRVersion() (string, error) {
	fhirVersion := cp.capStat["fhirVersion"]
	if fhirVersion == nil {
		return "", nil
	}
	fhirVersionStr, ok := fhirVersion.(string)
	if !ok {
		return "", errors.New("unable to cast DSTU2 capability statement fhirVersion value to a string")
	}
	return fhirVersionStr, nil
}

func (cp dstu2CapabilityParser) GetSoftwareName() (string, error) {
	software := cp.capStat["software"]
	if software == nil {
		return "", nil
	}
	softwareMap, ok := software.(map[string]interface{})
	if !ok {
		return "", errors.New("unable to cast DSTU2 capability statement software value to a map[string]interface{}")
	}
	name := softwareMap["name"]
	if name == nil {
		return "", nil
	}
	nameStr, ok := name.(string)
	if !ok {
		return "", errors.New("unable to cast DSTU2 capability statement software.name value to a string")
	}
	return nameStr, nil
}

func (cp dstu2CapabilityParser) GetSoftwareVersion() (string, error) {
	software := cp.capStat["software"]
	if software == nil {
		return "", nil
	}
	softwareMap, ok := software.(map[string]interface{})
	if !ok {
		return "", errors.New("unable to cast DSTU2 capability statement software value to a map[string]interface{}")
	}
	version := softwareMap["version"]
	if version == nil {
		return "", nil
	}
	versionStr, ok := version.(string)
	if !ok {
		return "", errors.New("unable to cast DSTU2 capability statement software.version value to a string")
	}
	return versionStr, nil
}
