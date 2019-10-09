package fhir

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

// ParseCapabilityStatement parses the Capability Statement in the body of the provided http response into a CapabilityStatement struct
// TODO: Make this function return appropriate version (DSTU2, DSTU3...)
func ParseCapabilityStatement(resp *http.Response) (DSTU2CapabilityStatement, error) {
	var capabilityStatement DSTU2CapabilityStatement

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return capabilityStatement, err
	}
	// TODO: Add Capability Statement JSON parser
	err = xml.Unmarshal(bodyBytes, &capabilityStatement)

	return capabilityStatement, err
}
