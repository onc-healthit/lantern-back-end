package fhir

import (
	"testing"
	"io/ioutil"
	"net/http"
	"bytes"
)

func Test_ParseConformanceStatement(t *testing.T) {
	var EXPECTED_FHIR_VERSION = "3.0.1"
	contents, err := ioutil.ReadFile("testdata/DSTU3CapabilityStatement.xml")
	resp := http.Response{
        Body: ioutil.NopCloser(bytes.NewBufferString(string(contents))),
    }
	if err != nil {
		t.Errorf("Error in sending mock request in test %s", err.Error())
	}
	var capabilityStatement = ParseCapabilityStatement(&resp)
	var FHIRVersion = capabilityStatement.FhirVersion.Value
	if FHIRVersion != EXPECTED_FHIR_VERSION {
		t.Errorf("Parsed incorrect FHIR version from capability statement got: %s, want: %s.", FHIRVersion, EXPECTED_FHIR_VERSION)
	}
}
