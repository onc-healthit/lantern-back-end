package endpointmanager

import (
	// "fmt"
	"testing"

	// "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	_ "github.com/lib/pq"
	// th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_FHIREndpoinNormalizeEndpointURL(t *testing.T) {
	if NormalizeEndpointURL("foobar.com") != "https://foobar.com/metadata" {
		t.Errorf("Expected foobar.com to be normalized to https://foobar.com/metadata")
	}
	if NormalizeEndpointURL("http://foobar.com") != "http://foobar.com/metadata" {
		t.Errorf("Expected http://foobar.com to be normalized to http://foobar.com/metadata")
	}
	if NormalizeEndpointURL("https://foobar.com") != "https://foobar.com/metadata" {
		t.Errorf("Expected https://foobar.com to be normalized to https://foobar.com/metadata")
	}
	if NormalizeEndpointURL("foobar.com/metadata") != "https://foobar.com/metadata" {
		t.Errorf("Expected foobar.com/metadata to be normalized to https://foobar.com/metadata")
	}
	if NormalizeEndpointURL("http://foobar.com/metadata") != "http://foobar.com/metadata" {
		t.Errorf("Expected http://foobar.com/metadata to be normalized to http://foobar.com/metadata")
	}
	if NormalizeEndpointURL("https://foobar.com/metadata") != "https://foobar.com/metadata" {
		t.Errorf("Expected https://foobar.com/metadata to be normalized to https://foobar.com/metadata")
	}
	if NormalizeEndpointURL("http://foobar.com/metadata/") != "http://foobar.com/metadata/" {
		t.Errorf("Expected http://foobar.com/metadata/ to be normalized to http://foobar.com/metadata/")
	}
	if NormalizeEndpointURL("https://foobar.com/metadata/") != "https://foobar.com/metadata/" {
		t.Errorf("Expected https://foobar.com/metadata/ to be normalized to https://foobar.com/metadata/")
	}
	if NormalizeEndpointURL("foobar.com/metadata/") != "https://foobar.com/metadata/" {
		t.Errorf("Expected foobar.com/metadata/ to be normalized to https://foobar.com/metadata/")
	}
}
func Test_FHIREndpoinNormalizeWellKnownURL(t *testing.T) {
	if NormalizeWellKnownURL("foobar.com") != "https://foobar.com/.well-known/smart-configuration" {
		t.Errorf("Expected foobar.com to be normalized to https://foobar.com/.well-known/smart-configuration")
	}
	if NormalizeWellKnownURL("http://foobar.com") != "http://foobar.com/.well-known/smart-configuration" {
		t.Errorf("Expected http://foobar.com to be normalized to http://foobar.com/.well-known/smart-configuration")
	}
	if NormalizeWellKnownURL("https://foobar.com") != "https://foobar.com/.well-known/smart-configuration" {
		t.Errorf("Expected https://foobar.com to be normalized to https://foobar.com/.well-known/smart-configuration")
	}
	if NormalizeWellKnownURL("foobar.com/.well-known/smart-configuration") != "https://foobar.com/.well-known/smart-configuration" {
		t.Errorf("Expected foobar.com/metadata to be normalized to https://foobar.com/.well-known/smart-configuration")
	}
	if NormalizeWellKnownURL("http://foobar.com/.well-known/smart-configuration") != "http://foobar.com/.well-known/smart-configuration" {
		t.Errorf("Expected http://foobar.com/.well-known/smart-configuration to be normalized to http://foobar.com/.well-known/smart-configuration")
	}
	if NormalizeWellKnownURL("https://foobar.com/.well-known/smart-configuration") != "https://foobar.com/.well-known/smart-configuration" {
		t.Errorf("Expected https://foobar.com/.well-known/smart-configuration to be normalized to https://foobar.com/.well-known/smart-configuration")
	}
	if NormalizeWellKnownURL("http://foobar.com/.well-known/smart-configuration/") != "http://foobar.com/.well-known/smart-configuration/" {
		t.Errorf("Expected http://foobar.com/.well-known/smart-configuration/ to be normalized to http://foobar.com/.well-known/smart-configuration/")
	}
	if NormalizeWellKnownURL("https://foobar.com/.well-known/smart-configuration/") != "https://foobar.com/.well-known/smart-configuration/" {
		t.Errorf("Expected https://foobar.com/.well-known/smart-configuration/ to be normalized to https://foobar.com/.well-known/smart-configuration/")
	}
	if NormalizeWellKnownURL("foobar.com/.well-known/smart-configuration/") != "https://foobar.com/.well-known/smart-configuration/" {
		t.Errorf("Expected foobar.com/.well-known/smart-configuration/ to be normalized to https://foobar.com/.well-known/smart-configuration/")
	}
}
func Test_FHIREndpoinNormalizeURL(t *testing.T) {
	if NormalizeURL("foobar.com") != "https://foobar.com" {
		t.Errorf("Expected foobar.com to be normalized to https://foobar.com")
	}
	if NormalizeURL("http://foobar.com") != "http://foobar.com" {
		t.Errorf("Expected http://foobar.com to be normalized to http://foobar.com")
	}
	if NormalizeURL("https://foobar.com") != "https://foobar.com" {
		t.Errorf("Expected https://foobar.com to be normalized to https://foobar.com")
	}
	if NormalizeURL("foobar.com/") != "https://foobar.com/" {
		t.Errorf("Expected foobar.com/ to be normalized to https://foobar.com/")
	}
	if NormalizeURL("http://foobar.com/") != "http://foobar.com/" {
		t.Errorf("Expected http://foobar.com/ to be normalized to http://foobar.com/")
	}
	if NormalizeURL("https://foobar.com/") != "https://foobar.com/" {
		t.Errorf("Expected https://foobar.com/ to be normalized to https://foobar.com/")
	}
}

func Test_NormalizeVersionsURLL(t *testing.T) {
	if NormalizeVersionsURL("foobar.com") != "https://foobar.com/$versions" {
		t.Errorf("Expected foobar.com to be normalized to https://foobar.com/$versions")
	}
	if NormalizeVersionsURL("http://foobar.com") != "http://foobar.com/$versions" {
		t.Errorf("Expected http://foobar.com to be normalized to http://foobar.com/$versions")
	}
	if NormalizeVersionsURL("https://foobar.com") != "https://foobar.com/$versions" {
		t.Errorf("Expected https://foobar.com to be normalized to https://foobar.com/$versions")
	}
	if NormalizeVersionsURL("foobar.com/") != "https://foobar.com/$versions" {
		t.Errorf("Expected foobar.com/ to be normalized to https://foobar.com/$versions")
	}
	if NormalizeVersionsURL("http://foobar.com/") != "http://foobar.com/$versions" {
		t.Errorf("Expected http://foobar.com/ to be normalized to http://foobar.com/$versions")
	}
	if NormalizeVersionsURL("https://foobar.com/") != "https://foobar.com/$versions" {
		t.Errorf("Expected https://foobar.com/ to be normalized to https://foobar.com/$versions")
	}
}

func Test_FHIREndpointEqual(t *testing.T) {
	// endpoints

	var epOrg1 = &FHIREndpointOrganization{
		OrganizationName:  "Example Org 1",
		OrganizationNPIID: "1"}

	var epOrg2 = &FHIREndpointOrganization{
		OrganizationName:  "Example Org 2",
		OrganizationNPIID: "2"}

	var epOrg3 = &FHIREndpointOrganization{
		OrganizationName:  "Example Org 1",
		OrganizationNPIID: "1"}

	var epOrg4 = &FHIREndpointOrganization{
		OrganizationName:  "Example Org 2",
		OrganizationNPIID: "2"}

	var endpoint1 = &FHIREndpoint{
		ID:               1,
		URL:              "example.com/FHIR/DSTU2",
		OrganizationList: []*FHIREndpointOrganization{epOrg1, epOrg2},
		ListSource:       "https://open.epic.com/Endpoints/DSTU2"}
	var endpoint2 = &FHIREndpoint{
		ID:               1,
		URL:              "example.com/FHIR/DSTU2",
		OrganizationList: []*FHIREndpointOrganization{epOrg3, epOrg4},
		ListSource:       "https://open.epic.com/Endpoints/DSTU2"}

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

	endpoint2.OrganizationList[0].OrganizationName = "Other 1"
	if endpoint1.Equal(endpoint2) {
		t.Error("Did not expect endpoint1 to equal endpoint 2. OrganizationNames should be different.")
	}
	endpoint2.OrganizationList[0].OrganizationName = endpoint1.OrganizationList[0].OrganizationName

	endpoint2.OrganizationList[0].OrganizationNPIID = "Other 1"
	if endpoint1.Equal(endpoint2) {
		t.Error("Did not expect endpoint1 to equal endpoint 2. NPIIDs should be different.")
	}
	endpoint2.OrganizationList[0].OrganizationNPIID = endpoint1.OrganizationList[0].OrganizationNPIID

	endpoint2.ListSource = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. ListSource should be different. %s vs %s", endpoint1.ListSource, endpoint2.ListSource)
	}
	endpoint2.ListSource = endpoint1.ListSource

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