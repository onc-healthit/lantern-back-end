package main

// CapabilityStatement represents a FHIR capability statement (or conformance statement if FHIR DSTU2 and below)
type CapabilityStatement struct {
}

// Equal checks if the CapabilityStatement is equal to the given CapabilityStatement
func (cs *CapabilityStatement) Equal(cs2 *CapabilityStatement) bool {
	if cs == nil && cs2 == nil {
		return true
	} else if cs == nil {
		return false
	} else if cs2 == nil {
		return false
	}

	return true
}
