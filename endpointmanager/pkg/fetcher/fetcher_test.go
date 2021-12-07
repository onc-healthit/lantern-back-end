package fetcher

import (
	"encoding/json"
	"fmt"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	logtest "github.com/sirupsen/logrus/hooks/test"
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

var testCareEvolution = []byte(`{"Entries":[
	{
		"OrganizationName":"Holy Cross in Florida - Trinity Health",
		"FHIRPatientFacingURI":"https://hcfl.patient.trinity-health.org/api/fhir"
	}]}`)

var test1Up = []byte(`{"Entries":[
		{
			"OrganizationName":"Spectrum Health",
			"FHIRPatientFacingURI":"https://epicarr02.spectrumhealth.org/EpicFHIR/api/FHIR/DSTU2"
		}]}`)

var testFHIR = []byte(`{"resourceType": "Bundle",
		"entry": [
			{
				"fullUrl": "http://hl7.org/fhir/Endpoint/71",
				"resource": {
					"name": "CarePlan repository",
					"managingOrganization": {
						"reference": "Telstra Health"
					},
					"address": "http://example2.com/DTSU2"
	}}]}`)

var testDefault = []byte(`{"Entries":[
	{
		"OrganizationName":"Test Default",
		"FHIRPatientFacingURI":"https://example.com"
	}]}`)

func Test_GetEndpointsFromFilepath(t *testing.T) {

	// test default list

	var expectedEndpoints = 1194
	var endpoints, _ = GetEndpointsFromFilepath("../../resources/CernerEndpointSources.json", "Cerner", "")
	var endpointsCount = len(endpoints.Entries)
	th.Assert(t, endpointsCount == expectedEndpoints, fmt.Sprintf("Number of endpoints read from resource file incorrect, got: %d, want: %d.", endpointsCount, expectedEndpoints))

	// test epic list

	expectedEndpoints = 364
	endpoints, _ = GetEndpointsFromFilepath("../../resources/EpicEndpointSources.json", "Epic", "")
	endpointsCount = len(endpoints.Entries)
	th.Assert(t, endpointsCount == expectedEndpoints, fmt.Sprintf("Number of endpoints read from epic file incorrect, got: %d, want: %d.", endpointsCount, expectedEndpoints))

	// test lantern list

	expectedEndpoints = 4
	endpoints, _ = GetEndpointsFromFilepath("../../resources/LanternEndpointSources.json", "Lantern", "")
	endpointsCount = len(endpoints.Entries)
	th.Assert(t, endpointsCount == expectedEndpoints, fmt.Sprintf("Number of endpoints read from lantern file incorrect, got: %d, want: %d.", endpointsCount, expectedEndpoints))

	// test CareEvolution list

	expectedEndpoints = 10
	endpoints, _ = GetEndpointsFromFilepath("../../resources/CareEvolutionEndpointSources.json", "CareEvolution", "")
	endpointsCount = len(endpoints.Entries)
	th.Assert(t, endpointsCount == expectedEndpoints, fmt.Sprintf("Number of endpoints read from CareEvolution file incorrect, got: %d, want: %d.", endpointsCount, expectedEndpoints))

	// test 1Up list

	expectedEndpoints = 472
	endpoints, _ = GetEndpointsFromFilepath("../../resources/1UpEndpointSources.json", "1Up", "")
	endpointsCount = len(endpoints.Entries)
	th.Assert(t, endpointsCount == expectedEndpoints, fmt.Sprintf("Number of endpoints read from 1Up file incorrect, got: %d, want: %d.", endpointsCount, expectedEndpoints))

	// test fhir list

	expectedEndpoints = 14
	endpoints, _ = GetEndpointsFromFilepath("../../resources/FHIREndpointSources.json", "FHIR", "")
	endpointsCount = len(endpoints.Entries)
	th.Assert(t, endpointsCount == expectedEndpoints, fmt.Sprintf("Number of endpoints read from FHIR file incorrect, got: %d, want: %d.", endpointsCount, expectedEndpoints))
}

func Test_GetListOfEndpointsKnownSource(t *testing.T) {

	// test cerner list

	cernerListSource := "cerner.com/fhir-endpoints"
	cernerResult, err := GetListOfEndpointsKnownSource(testCerner, "Cerner", cernerListSource)
	th.Assert(t, err == nil, err)
	th.Assert(t, cernerResult.Entries[0].ListSource == cernerListSource, fmt.Sprintf("The list source should have been %s, it instead returned %s", cernerListSource, cernerResult.Entries[0].ListSource))

	// test epic list

	epicResult, err := GetListOfEndpointsKnownSource(testEpic, "Epic", "")
	th.Assert(t, err == nil, err)
	th.Assert(t, epicResult.Entries[0].ListSource == "Epic", fmt.Sprintf("The list source should have been Epic, it instead returned %s", epicResult.Entries[0].ListSource))

	// test lantern list

	lanternResult, err := GetListOfEndpointsKnownSource(testLantern, "Lantern", "")
	th.Assert(t, err == nil, err)
	th.Assert(t, lanternResult.Entries[0].ListSource == "Lantern", fmt.Sprintf("The list source should have been Lantern, it instead returned %s", lanternResult.Entries[0].ListSource))

	// test CareEvolution list
	careEvolutionResult, err := GetListOfEndpointsKnownSource(testCareEvolution, "CareEvolution", "")
	th.Assert(t, err == nil, err)
	th.Assert(t, careEvolutionResult.Entries[0].ListSource == "CareEvolution", fmt.Sprintf("The list source should have been CareEvolution, it instead returned %s", careEvolutionResult.Entries[0].ListSource))

	// test 1Up list
	oneUpResult, err := GetListOfEndpointsKnownSource(test1Up, "1Up", "")
	th.Assert(t, err == nil, err)
	th.Assert(t, oneUpResult.Entries[0].ListSource == "1Up", fmt.Sprintf("The list source should have been 1Up, it instead returned %s", oneUpResult.Entries[0].ListSource))

	// test FHIR list

	fhirListSource := "www.thisisafhirlist.com"
	fhirResult, err := GetListOfEndpointsKnownSource(testFHIR, "FHIR", fhirListSource)
	th.Assert(t, err == nil, err)
	th.Assert(t, fhirResult.Entries[0].ListSource == fhirListSource, fmt.Sprintf("The list source should have been %s, it instead returned %s", fhirListSource, fhirResult.Entries[0].ListSource))

	// test empty values

	_, err = GetListOfEndpointsKnownSource([]byte("null"), "Epic", "")
	th.Assert(t, err == nil, fmt.Sprintf("A null value should have returned nil, it instead returned %s", err))

	_, err = GetListOfEndpointsKnownSource([]byte("{}"), "Epic", "")
	th.Assert(t, err == nil, fmt.Sprintf("An empty map {} should have returned nil, it instead returned %s", err))

	_, err = GetListOfEndpointsKnownSource([]byte("[]"), "Epic", "")
	th.Assert(t, err != nil, "An empty list [] should have returned an error, it instead returned nil")

	// test improperly formatted list

	_, err = GetListOfEndpointsKnownSource([]byte(`{ "endpoints": "string" }`), "Epic", "")
	th.Assert(t, err != nil, "An improperly formatted list should have returned an error, it instead returned nil")

	// test improperly formatted fhir list
	hook := logtest.NewGlobal()
	expectedErr := "No resource field in FHIR list. Returning an empty list of entries."
	_, _ = GetListOfEndpointsKnownSource([]byte(`{ "entry": [{ "notresource": {}}] }`), "FHIR", "")
	// expect presence of a log message
	found := false
	for i := range hook.Entries {
		if hook.Entries[i].Message == expectedErr {
			found = true
			break
		}
	}
	th.Assert(t, found, "expected a resource field missing message to be logged")

	// test fhir list entry with no address
	expectedErr = "No address field in the resource. Ignoring resource."
	_, _ = GetListOfEndpointsKnownSource([]byte(`{ "entry": [{ "resource": { "notAddress" : "" }}] }`), "FHIR", "")
	// expect presence of a log message
	found = false
	for i := range hook.Entries {
		if hook.Entries[i].Message == expectedErr {
			found = true
			break
		}
	}
	th.Assert(t, found, "expected an address field missing message to be logged")

	// test invalid source

	_, err = GetListOfEndpointsKnownSource(testEpic, "string", "")
	th.Assert(t, err != nil, "An invalid source should have thrown an error")
}

func Test_GetListOfEndpoints(t *testing.T) {

	// test default list

	defaultResult, err := GetListOfEndpoints(testDefault, "Test", "")
	th.Assert(t, err == nil, err)
	th.Assert(t, defaultResult.Entries[0].ListSource == "Test", fmt.Sprintf("The list source should have been 'Test', it instead returned %s", defaultResult.Entries[0].ListSource))

	// test default with given list source name

	testListSource := "test.com/fhir-lists"
	defaultResult2, err := GetListOfEndpoints(testDefault, "Test", testListSource)
	th.Assert(t, err == nil, err)
	th.Assert(t, defaultResult2.Entries[0].ListSource == testListSource, fmt.Sprintf("The list source should have been %s, it instead returned %s", testListSource, defaultResult2.Entries[0].ListSource))

	// test empty list

	_, err = GetListOfEndpoints([]byte("null"), "Test", "")
	th.Assert(t, err == nil, fmt.Sprintf("A null value should have returned nil, it instead returned %s", err))

	_, err = GetListOfEndpoints([]byte("{}"), "Test", "")
	th.Assert(t, err == nil, fmt.Sprintf("An empty map {} should have returned nil, it instead returned %s", err))

	_, err = GetListOfEndpoints([]byte("[]"), "Test", "")
	th.Assert(t, err != nil, "An empty list [] should have returned an error, it instead returned nil")

	// test improperly formatted list

	_, err = GetListOfEndpoints([]byte(`{ "endpoints": "string" }`), "Test", "")
	th.Assert(t, err != nil, "An improperly formatted list should have returned an error, it instead returned nil")

	// test invalid formatting

	_, err = GetListOfEndpoints(testCerner, "Test", "")
	th.Assert(t, err != nil, "An invalid source format should have thrown an error")
}

func Test_convertInterfaceToList(t *testing.T) {
	// base test

	var initialList map[string]interface{}
	err := json.Unmarshal(testFHIR, &initialList)
	th.Assert(t, err == nil, "The given JSON should have been valid")
	resultList, err := convertInterfaceToList(initialList, "entry")
	th.Assert(t, err == nil, "The given list should have been converted to a []map[string]interface{}")
	th.Assert(t, len(resultList) == 1, fmt.Sprintf("The result should have a length of 1, instead has %d", len(resultList)))

	// incorrect reference value

	_, err = convertInterfaceToList(initialList, "Entries")
	th.Assert(t, err != nil, fmt.Sprintf("Should have thrown an incorrect reference value error, instead threw %s", err))

	// the referenced value is not an array

	var initialList2 map[string]interface{}
	testNoArray := []byte(`{"resourceType": "Bundle",
		"entry": "broken JSON" }`)
	err = json.Unmarshal(testNoArray, &initialList2)
	th.Assert(t, err == nil, "The given JSON should have been valid")
	_, err = convertInterfaceToList(initialList2, "entry")
	th.Assert(t, err != nil, fmt.Sprintf("Should have thrown endpoint list is not an array error, instead threw %s", err))

	// the referenced array is not made of map[string]interface{}

	var initialList3 map[string]interface{}
	testNoMap := []byte(`{"resourceType": "Bundle",
		"entry": [1, 2, 3] }`)
	err = json.Unmarshal(testNoMap, &initialList3)
	th.Assert(t, err == nil, "The given JSON should have been valid")
	_, err = convertInterfaceToList(initialList3, "entry")
	th.Assert(t, err != nil, fmt.Sprintf("Should have thrown endpoint list is not map[string]interface{} error, instead threw %s", err))
}
