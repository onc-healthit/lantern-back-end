package fhir

import (
	"testing"
	"io/ioutil"
	"net/http"
	"time"
	"bytes"
)

var netClient = &http.Client{
	Timeout: time.Second * 35,
}

func Test_ParseConformanceStatement(t *testing.T) {
	var EXPECTED_FHIR_VERSION = "1.0.1"
	contents, err := ioutil.ReadFile("testdata/DSTU2capabilityStatement.xml")
	resp := http.Response{
        Body: ioutil.NopCloser(bytes.NewBufferString(string(contents))),
    }
	if err != nil {
		t.Errorf("Error in sending mock request in test %s", err.Error())
	}
	var capabilityStatement = ParseConformanceStatement(&resp)
	var FHIRVersion = capabilityStatement.FhirVersion.Value
	if FHIRVersion != EXPECTED_FHIR_VERSION {
		t.Errorf("Number of endpoints read from resource file incorrect, got: %s, want: %s.", FHIRVersion, EXPECTED_FHIR_VERSION)
	}
}
