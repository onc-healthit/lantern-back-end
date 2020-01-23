package capabilityparser

type r4CapabilityParser struct {
	capStat map[string]interface{}
}

func newR4(capStat map[string]interface{}) CapabilityStatement {
	return &dstu2CapabilityParser{
		capStat: capStat,
	}
}

func (cp r4CapabilityParser) GetPublisher() (string, error) {
	publisher := cp.capStat["publisher"]
	if publisher == nil {
		return "", nil
	}
	return publisher.(string), nil
}

func (cp r4CapabilityParser) GetFHIRVersion() (string, error) {
	fhirVersion := cp.capStat["fhirVersion"]
	if fhirVersion == nil {
		return "", nil
	}
	return fhirVersion.(string), nil
}
