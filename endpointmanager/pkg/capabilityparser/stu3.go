package capabilityparser

type stu3CapabilityParser struct {
	capStat map[string]interface{}
}

func newSTU3(capStat map[string]interface{}) CapabilityStatement {
	return &dstu2CapabilityParser{
		capStat: capStat,
	}
}

func (cp stu3CapabilityParser) GetPublisher() (string, error) {
	publisher := cp.capStat["publisher"]
	if publisher == nil {
		return "", nil
	}
	return publisher.(string), nil
}

func (cp stu3CapabilityParser) GetFHIRVersion() (string, error) {
	fhirVersion := cp.capStat["fhirVersion"]
	if fhirVersion == nil {
		return "", nil
	}
	return fhirVersion.(string), nil
}
