package fetcher

import (
	"encoding/json"
	"testing"
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

func Test_GetListOfEndpoints(t *testing.T) {
	var expectedEndpoints = 397
	var endpoints, _ = GetListOfEndpoints("../resources/EndpointSources.json")
	var endpointsCount = len(endpoints.Entries)
	if endpointsCount != expectedEndpoints {
		t.Errorf("Number of endpoints read from resource file incorrect, got: %d, want: %d.", endpointsCount, expectedEndpoints)
	}
}

func Test_formatList(t *testing.T) {

	// test cerner list

	var cernerList map[string]interface{}
	err := json.Unmarshal(testCerner, &cernerList)

	if cernerList == nil {
		t.Errorf("The cerner list was not unmarshalled properly")
	}
	if err != nil {
		t.Errorf("%s", err)
	}

	cernerResult, err := formatList(cernerList)
	if err != nil {
		t.Errorf("%s", err)
	}
	if cernerResult.Entries[0].ListSource != "Cerner" {
		t.Errorf("The list source should have been cerner, it instead returned %s", cernerResult.Entries[0].ListSource)
	}

	// test epic list

	var epicList map[string]interface{}
	err = json.Unmarshal(testEpic, &epicList)

	if epicList == nil {
		t.Errorf("The epic list was not unmarshalled properly")
	}
	if err != nil {
		t.Errorf("%s", err)
	}

	epicResult, err := formatList(epicList)
	if err != nil {
		t.Errorf("%s", err)
	}
	if epicResult.Entries[0].ListSource != "Epic" {
		t.Errorf("The endpoint list source should have been epic, it instead returned %s", epicResult.Entries[0].ListSource)
	}

	// test unknown format list
	var emptyList map[string]interface{}
	emptyResult, _ := formatList(emptyList)
	if len(emptyResult.Entries) > 0 {
		t.Errorf("The endpoint list should have been empty, it instead is of length %d", len(epicResult.Entries))
	}
}
