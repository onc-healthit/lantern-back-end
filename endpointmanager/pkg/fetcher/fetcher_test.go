package fetcher

import (
	"fmt"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

var testCerner = []byte(`{"endpoints": [
    {
      "name": "A Woman's Place, LLC",
      "baseUrl": "https://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2/",
      "type": "prod"
	}]}`)

var testEpic = []byte(`{"Entries":[
	{
		"OrganizationName":"Access Community Health Network",
		"FHIRPatientFacingURI":"https://eprescribing.accesscommunityhealth.net/FHIR/api/FHIR/DSTU2/"
	}]}`)

var testLantern = []byte(`{"Endpoints": [
    {
		"URL": "http://example.com/DTSU2/",
        "OrganizationName": "fakeOrganization",
        "NPIID": "1"
	}]}`)

var testDefault = []byte(`{"Entries":[
	{
		"OrganizationName":"Test Default",
		"FHIRPatientFacingURI":"https://example.com"
	}]}`)

func Test_GetEndpointsFromFilepath(t *testing.T) {

	// test default list

	var expectedEndpoints = 1194
	var endpoints, _ = GetEndpointsFromFilepath("../../resources/CernerEndpointSources.json", "Cerner")
	var endpointsCount = len(endpoints.Entries)
	th.Assert(t, endpointsCount == expectedEndpoints, fmt.Sprintf("Number of endpoints read from resource file incorrect, got: %d, want: %d.", endpointsCount, expectedEndpoints))

	// test epic list

	expectedEndpoints = 364
	endpoints, _ = GetEndpointsFromFilepath("../../resources/EpicEndpointSources.json", "Epic")
	endpointsCount = len(endpoints.Entries)
	th.Assert(t, endpointsCount == expectedEndpoints, fmt.Sprintf("Number of endpoints read from epic file incorrect, got: %d, want: %d.", endpointsCount, expectedEndpoints))

	// test lantern list

	expectedEndpoints = 4
	endpoints, _ = GetEndpointsFromFilepath("../../resources/LanternEndpointSources.json", "Lantern")
	endpointsCount = len(endpoints.Entries)
	th.Assert(t, endpointsCount == expectedEndpoints, fmt.Sprintf("Number of endpoints read from lantern file incorrect, got: %d, want: %d.", endpointsCount, expectedEndpoints))
}

func Test_GetListOfEndpointsKnownSource(t *testing.T) {

	// test cerner list

	cernerResult, err := GetListOfEndpointsKnownSource(testCerner, Cerner)
	th.Assert(t, err == nil, err)
	th.Assert(t, cernerResult.Entries[0].ListSource == string(Cerner), fmt.Sprintf("The list source should have been %s, it instead returned %s", Cerner, cernerResult.Entries[0].ListSource))

	// test epic list

	epicResult, err := GetListOfEndpointsKnownSource(testEpic, Epic)
	th.Assert(t, err == nil, err)
	th.Assert(t, epicResult.Entries[0].ListSource == string(Epic), fmt.Sprintf("The list source should have been %s, it instead returned %s", Epic, epicResult.Entries[0].ListSource))

	// test lantern list

	lanternResult, err := GetListOfEndpointsKnownSource(testLantern, Lantern)
	th.Assert(t, err == nil, err)
	th.Assert(t, lanternResult.Entries[0].ListSource == string(Lantern), fmt.Sprintf("The list source should have been %s, it instead returned %s", Lantern, lanternResult.Entries[0].ListSource))

	// test empty values

	_, err = GetListOfEndpointsKnownSource([]byte("null"), Epic)
	th.Assert(t, err == nil, fmt.Sprintf("A null value should have returned nil, it instead returned %s", err))

	_, err = GetListOfEndpointsKnownSource([]byte("{}"), Epic)
	th.Assert(t, err == nil, fmt.Sprintf("An empty map {} should have returned nil, it instead returned %s", err))

	_, err = GetListOfEndpointsKnownSource([]byte("[]"), Epic)
	th.Assert(t, err != nil, "An empty list [] should have returned an error, it instead returned nil")

	// test improperly formatted list

	_, err = GetListOfEndpointsKnownSource([]byte(`{ "endpoints": "string" }`), Epic)
	th.Assert(t, err != nil, "An improperly formatted list should have returned an error, it instead returned nil")

	// test invalid source

	_, err = GetListOfEndpointsKnownSource(testEpic, "string")
	th.Assert(t, err != nil, "An invalid source should have thrown an error")
}

func Test_GetListOfEndpoints(t *testing.T) {

	// test default list

	defaultResult, err := GetListOfEndpoints(testDefault, "Test")
	th.Assert(t, err == nil, err)
	th.Assert(t, defaultResult.Entries[0].ListSource == "Test", fmt.Sprintf("The list source should have been 'Test', it instead returned %s", defaultResult.Entries[0].ListSource))

	// test empty list

	_, err = GetListOfEndpoints([]byte("null"), "Test")
	th.Assert(t, err == nil, fmt.Sprintf("A null value should have returned nil, it instead returned %s", err))

	_, err = GetListOfEndpoints([]byte("{}"), "Test")
	th.Assert(t, err == nil, fmt.Sprintf("An empty map {} should have returned nil, it instead returned %s", err))

	_, err = GetListOfEndpoints([]byte("[]"), "Test")
	th.Assert(t, err != nil, "An empty list [] should have returned an error, it instead returned nil")

	// test improperly formatted list

	_, err = GetListOfEndpoints([]byte(`{ "endpoints": "string" }`), "Test")
	th.Assert(t, err != nil, "An improperly formatted list should have returned an error, it instead returned nil")

	// test invalid formatting

	_, err = GetListOfEndpoints(testCerner, "Test")
	th.Assert(t, err != nil, "An invalid source format should have thrown an error")
}
