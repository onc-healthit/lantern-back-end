package endpointmanager

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
)

func Test_FHIREndpointEqual(t *testing.T) {

	// capability statement
	path := filepath.Join("../testdata", "cerner_capability_dstu2.json")
	csJSON, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
	}
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	if err != nil {
		t.Error(err)
	}

	// endpoints
	var endpoint1 = &FHIREndpoint{
		ID:                  1,
		URL:                 "example.com/FHIR/DSTU2",
		TLSVersion:          "TLS 1.1",
		MIMETypes:           []string{"application/json+fhir", "application/fhir+json"},
		HTTPResponse:        200,
		Errors:              "Example Error",
		OrganizationName:    "Example Org",
		ListSource:          "https://open.epic.com/MyApps/EndpointsJson",
		CapabilityStatement: cs}
	var endpoint2 = &FHIREndpoint{
		ID:                  1,
		URL:                 "example.com/FHIR/DSTU2",
		TLSVersion:          "TLS 1.1",
		MIMETypes:           []string{"application/json+fhir", "application/fhir+json"},
		HTTPResponse:        200,
		Errors:              "Example Error",
		OrganizationName:    "Example Org",
		ListSource:          "https://open.epic.com/MyApps/EndpointsJson",
		CapabilityStatement: cs}

	if !endpoint1.Equal(endpoint2) {
		t.Errorf("Expected endpoint1 to equal endpoint2. They are not equal.")
	}

	endpoint2.ID = 2
	if !endpoint1.Equal(endpoint2) {
		t.Errorf("Expect endpoint 1 to equal endpoint 2. ids should be ignored. %d vs %d", endpoint1.ID, endpoint2.ID)
	}
	endpoint2.ID = endpoint1.ID

	endpoint2.URL = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. URL should be different. %s vs %s", endpoint1.URL, endpoint2.URL)
	}
	endpoint2.URL = endpoint1.URL
	endpoint2.TLSVersion = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. TLSVersion should be different. %s vs %s", endpoint1.TLSVersion, endpoint2.TLSVersion)
	}
	endpoint2.TLSVersion = endpoint1.TLSVersion

	endpoint2.MIMETypes = []string{"other"}
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. MIMETypes should be different. %s vs %s", endpoint1.MIMETypes, endpoint2.MIMETypes)
	}
	endpoint2.MIMETypes = []string{"application/fhir+json", "application/json+fhir"}
	if !endpoint1.Equal(endpoint2) {
		t.Errorf("Expected endpoint1 to equal endpoint 2. MIMETypes are same but in different order. %s vs %s", endpoint1.MIMETypes, endpoint2.MIMETypes)
	}
	endpoint2.MIMETypes = endpoint1.MIMETypes

	endpoint2.HTTPResponse = 404
	if endpoint2.Equal(endpoint1) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. HTTPResponse should be different. %d vs %d", endpoint1.HTTPResponse, endpoint2.HTTPResponse)
	}
	endpoint2.HTTPResponse = endpoint1.HTTPResponse

	endpoint2.Errors = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. Errors should be different. %s vs %s", endpoint1.Errors, endpoint2.Errors)
	}
	endpoint2.Errors = endpoint1.Errors

	endpoint2.OrganizationName = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. OrganizationName should be different. %s vs %s", endpoint1.OrganizationName, endpoint2.OrganizationName)
	}
	endpoint2.OrganizationName = endpoint1.OrganizationName

	endpoint2.ListSource = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. ListSource should be different. %s vs %s", endpoint1.ListSource, endpoint2.ListSource)
	}
	endpoint2.ListSource = endpoint1.ListSource

	// @TODO Currently commented out while figuring out Capability Parsing
	// endpoint2.CapabilityStatement = nil
	// if endpoint1.Equal(endpoint2) {
	// 	t.Errorf("Did not expect endpoint1 to equal endpoint 2. CapabilityStatement should be different. %s vs %s", endpoint1.CapabilityStatement, endpoint2.CapabilityStatement)
	// }
	// endpoint2.CapabilityStatement = endpoint1.CapabilityStatement

	endpoint2 = nil
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal nil endpoint 2.")
	}
	endpoint2 = endpoint1

	endpoint1 = nil
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect nil endpoint1 to equal endpoint 2.")
	}

	endpoint2 = nil
	if !endpoint1.Equal(endpoint2) {
		t.Errorf("Nil endpoint 1 should equal nil endpoint 2.")
	}
}
