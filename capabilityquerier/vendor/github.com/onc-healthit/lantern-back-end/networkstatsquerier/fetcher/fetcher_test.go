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

var testDefault = []byte(`{"Entries":[
	{
		"OrganizationName":"Test Default",
		"FHIRPatientFacingURI":"https://example.com"
	}]}`)

func Test_GetEndpointsFromFilepath(t *testing.T) {

	// test default list

	var expectedEndpoints = 397
	var endpoints, _ = GetEndpointsFromFilepath("../resources/EndpointSources.json", "CareEvolution")
	var endpointsCount = len(endpoints.Entries)
	th.Assert(t, endpointsCount == expectedEndpoints, fmt.Sprintf("Number of endpoints read from resource file incorrect, got: %d, want: %d.", endpointsCount, expectedEndpoints))

	// test epic list

	expectedEndpoints = 364
	endpoints, _ = GetEndpointsFromFilepath("../resources/EpicEndpointSources.json", "Epic")
	endpointsCount = len(endpoints.Entries)
	th.Assert(t, endpointsCount == expectedEndpoints, fmt.Sprintf("Number of endpoints read from epic file incorrect, got: %d, want: %d.", endpointsCount, expectedEndpoints))

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

	// test empty list

	_, err = GetListOfEndpointsKnownSource([]byte(""), Epic)
	th.Assert(t, err == nil, fmt.Sprintf("An empty list should have returned nil, it instead returned %s", err))

	// test invalid source

	_, err = GetListOfEndpointsKnownSource(testEpic, "string")
	th.Assert(t, err != nil, "An invalid source should have thrown an error")
}

func Test_GetListOfEndpoints(t *testing.T) {

	// test default list

	defaultResult, err := GetListOfEndpoints(testDefault, "Test")
	th.Assert(t, err == nil, err)
	th.Assert(t, defaultResult.Entries[0].ListSource == "Test", fmt.Sprintf("The list source should have been CareEvolution, it instead returned %s", defaultResult.Entries[0].ListSource))

	// test empty list

	_, err = GetListOfEndpoints([]byte(""), "Test")
	th.Assert(t, err == nil, fmt.Sprintf("An empty list should have returned nil, it instead returned %s", err))

	// test invalid formatting

	_, err = GetListOfEndpoints(testCerner, "Test")
	th.Assert(t, err != nil, "An invalid source format should have thrown an error")
}
