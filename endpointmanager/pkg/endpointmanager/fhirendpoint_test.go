package endpointmanager

import (
	"testing"

	_ "github.com/lib/pq"
)

func Test_FHIREndpointEqual(t *testing.T) {
	var endpoint1 = &FHIREndpoint{
		ID:                    1,
		URL:                   "example.com/FHIR/DSTU2",
		TLSVersion:            "TLS 1.1",
		MimeType:              "application/json+fhir",
		OrganizationName:      "Example Org",
		FHIRVersion:           "DSTU2",
		AuthorizationStandard: "OAuth 2.0",
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		CapabilityStatement: &CapabilityStatement{}}
	var endpoint2 = &FHIREndpoint{
		ID:                    1,
		URL:                   "example.com/FHIR/DSTU2",
		TLSVersion:            "TLS 1.1",
		MimeType:              "application/json+fhir",
		OrganizationName:      "Example Org",
		FHIRVersion:           "DSTU2",
		AuthorizationStandard: "OAuth 2.0",
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		CapabilityStatement: &CapabilityStatement{}}

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

	endpoint2.FHIRVersion = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. FHIRVersion should be different. %s vs %s", endpoint1.FHIRVersion, endpoint2.FHIRVersion)
	}
	endpoint2.FHIRVersion = endpoint1.FHIRVersion

	endpoint2.TLSVersion = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. TLSVersion should be different. %s vs %s", endpoint1.TLSVersion, endpoint2.TLSVersion)
	}
	endpoint2.TLSVersion = endpoint1.TLSVersion

	endpoint2.MimeType = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. MimeType should be different. %s vs %s", endpoint1.MimeType, endpoint2.MimeType)
	}
	endpoint2.MimeType = endpoint1.MimeType

	endpoint2.OrganizationName = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. OrganizationName should be different. %s vs %s", endpoint1.OrganizationName, endpoint2.OrganizationName)
	}
	endpoint2.OrganizationName = endpoint1.OrganizationName

	endpoint2.AuthorizationStandard = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. AuthorizationStandard should be different. %s vs %s", endpoint1.AuthorizationStandard, endpoint2.AuthorizationStandard)
	}
	endpoint2.AuthorizationStandard = endpoint1.AuthorizationStandard

	endpoint2.Location.Address1 = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. Location should be different. %s vs %s", endpoint1.Location.Address1, endpoint2.Location.Address1)
	}
	endpoint2.Location.Address1 = endpoint1.Location.Address1

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
