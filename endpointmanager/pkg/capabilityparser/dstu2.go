package capabilityparser

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
	return publisher.(string), nil
}

func (cp dstu2CapabilityParser) GetFHIRVersion() (string, error) {
	fhirVersion := cp.capStat["fhirVersion"]
	if fhirVersion == nil {
		return "", nil
	}
	return fhirVersion.(string), nil
}
