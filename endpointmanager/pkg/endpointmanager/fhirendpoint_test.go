package endpointmanager

import (
	"testing"

	_ "github.com/lib/pq"
)

func Test_FHIREndpoinNormalizeEndpointURL(t *testing.T) {
	if NormalizeURL("foobar.com") != "https://foobar.com/metadata" {
		t.Errorf("Expected foobar.com to be normalized to https://foobar.com/metadata")
	}
	if NormalizeURL("http://foobar.com") != "http://foobar.com/metadata" {
		t.Errorf("Expected http://foobar.com to be normalized to http://foobar.com/metadata")
	}
	if NormalizeURL("https://foobar.com") != "https://foobar.com/metadata" {
		t.Errorf("Expected https://foobar.com to be normalized to https://foobar.com/metadata")
	}
	if NormalizeURL("foobar.com/metadata") != "https://foobar.com/metadata" {
		t.Errorf("Expected foobar.com/metadata to be normalized to https://foobar.com/metadata")
	}
	if NormalizeURL("http://foobar.com/metadata") != "http://foobar.com/metadata" {
		t.Errorf("Expected http://foobar.com/metadata to be normalized to http://foobar.com/metadata")
	}
	if NormalizeURL("https://foobar.com/metadata") != "https://foobar.com/metadata" {
		t.Errorf("Expected https://foobar.com/metadata to be normalized to https://foobar.com/metadata")
	}
	if NormalizeURL("http://foobar.com/metadata/") != "http://foobar.com/metadata/" {
		t.Errorf("Expected http://foobar.com/metadata/ to be normalized to http://foobar.com/metadata/")
	}
	if NormalizeURL("https://foobar.com/metadata/") != "https://foobar.com/metadata/" {
		t.Errorf("Expected https://foobar.com/metadata/ to be normalized to https://foobar.com/metadata/")
	}
	if NormalizeURL("foobar.com/metadata/") != "https://foobar.com/metadata/" {
		t.Errorf("Expected foobar.com/metadata/ to be normalized to https://foobar.com/metadata/")
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
func Test_FHIREndpointEqual(t *testing.T) {
	// endpoints
	var endpoint1 = &FHIREndpoint{
		ID:               1,
		URL:              "example.com/FHIR/DSTU2",
		OrganizationName: "Example Org",
		ListSource:       "https://open.epic.com/MyApps/EndpointsJson"}
	var endpoint2 = &FHIREndpoint{
		ID:               1,
		URL:              "example.com/FHIR/DSTU2",
		OrganizationName: "Example Org",
		ListSource:       "https://open.epic.com/MyApps/EndpointsJson"}

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
