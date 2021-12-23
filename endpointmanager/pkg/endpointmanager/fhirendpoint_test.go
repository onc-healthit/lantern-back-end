package endpointmanager

import (
	"fmt"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	_ "github.com/lib/pq"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
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
	var endpoint1 = &FHIREndpoint{
		ID:                1,
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{"Example Org 1", "Example Org 2"},
		NPIIDs:            []string{"1", "2", "3"},
		ListSource:        "https://open.epic.com/Endpoints/DSTU2"}
	var endpoint2 = &FHIREndpoint{
		ID:                1,
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{"Example Org 1", "Example Org 2"},
		NPIIDs:            []string{"1", "2", "3"},
		ListSource:        "https://open.epic.com/Endpoints/DSTU2"}

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

	endpoint2.OrganizationNames = []string{"Other 1"}
	if endpoint1.Equal(endpoint2) {
		t.Error("Did not expect endpoint1 to equal endpoint 2. OrganizationNames should be different.")
	}
	endpoint2.OrganizationNames = endpoint1.OrganizationNames

	endpoint2.NPIIDs = []string{"Other 1"}
	if endpoint1.Equal(endpoint2) {
		t.Error("Did not expect endpoint1 to equal endpoint 2. NPIIDs should be different.")
	}
	endpoint2.NPIIDs = endpoint1.NPIIDs

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

func Test_AddOrganizationName(t *testing.T) {
	var endpoint = &FHIREndpoint{
		ID:                1,
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{},
		NPIIDs:            []string{"1", "2", "3"},
		ListSource:        "https://open.epic.com/Endpoints/DSTU2"}

	var orgName string
	var expected []string

	// test with empty org names list
	orgName = "Example Org Name 1"
	expected = []string{"Example Org Name 1"}
	endpoint.AddOrganizationName(orgName)
	th.Assert(t, helpers.StringArraysEqual(endpoint.OrganizationNames, expected), fmt.Sprintf("expected %v to equal %v", endpoint.OrganizationNames, expected))

	// test with non-empty org names list
	orgName = "Example Org Name 2"
	expected = []string{"Example Org Name 2", "Example Org Name 1"}
	endpoint.AddOrganizationName(orgName)
	th.Assert(t, helpers.StringArraysEqual(endpoint.OrganizationNames, expected), fmt.Sprintf("expected %v to equal %v", endpoint.OrganizationNames, expected))

	// test with org name that's already in list
	orgName = "Example Org Name 2"
	expected = []string{"Example Org Name 2", "Example Org Name 1"}
	endpoint.AddOrganizationName(orgName)
	th.Assert(t, helpers.StringArraysEqual(endpoint.OrganizationNames, expected), fmt.Sprintf("expected %v to equal %v", endpoint.OrganizationNames, expected))
}

func Test_AddNPIID(t *testing.T) {
	var endpoint = &FHIREndpoint{
		ID:                1,
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{"Example Org Name 1", "Example Org Name 2"},
		NPIIDs:            []string{},
		ListSource:        "https://open.epic.com/Endpoints/DSTU2"}

	var npiID string
	var expected []string

	// test with empty npi ids list
	npiID = "1"
	expected = []string{"1"}
	endpoint.AddNPIID(npiID)
	th.Assert(t, helpers.StringArraysEqual(endpoint.NPIIDs, expected), fmt.Sprintf("expected %v to equal %v", endpoint.NPIIDs, expected))

	// test with non-empty npi ids list
	npiID = "2"
	expected = []string{"2", "1"}
	endpoint.AddNPIID(npiID)
	th.Assert(t, helpers.StringArraysEqual(endpoint.NPIIDs, expected), fmt.Sprintf("expected %v to equal %v", endpoint.NPIIDs, expected))

	// test with npi id that's already in list
	npiID = "2"
	expected = []string{"2", "1"}
	endpoint.AddNPIID(npiID)
	th.Assert(t, helpers.StringArraysEqual(endpoint.NPIIDs, expected), fmt.Sprintf("expected %v to equal %v", endpoint.NPIIDs, expected))
}
