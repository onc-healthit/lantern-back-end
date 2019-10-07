package fhir

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

func ParseConformanceStatement(resp *http.Response) (CapabilityStatement){
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// TODO: Use a logging solution instead of println
		println("Capability Statement Response Body Reading Error: ", err.Error())
	}
	var capabilityStatement CapabilityStatement
	err = xml.Unmarshal(bodyBytes, &capabilityStatement)
	if err != nil {
		// TODO: Use a logging solution instead of println
		println("Capability Statement Parsing Error: ", err.Error())
	}
	return capabilityStatement
}