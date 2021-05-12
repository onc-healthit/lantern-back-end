package jsonexport

import (
	"fmt"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_getSMARTResponse(t *testing.T) {
	// Base case

	testSmartResp := []byte(`
	{
		"authorization_endpoint": "https://ehr.example.com/auth/authorize",
		"token_endpoint": "https://ehr.example.com/auth/token"
	}`)
	smartRsp := getSMARTResponse(testSmartResp)
	th.Assert(t, smartRsp["authorization_endpoint"] == "https://ehr.example.com/auth/authorize", fmt.Sprintf("SMART Response should have correct authorization endpoint URL. It is instead %s", smartRsp["authorization_endpoint"]))

	// If smart response is nonsense, should be an empty map[string]interface{}

	testSmartResp = []byte("this is not JSON")
	smartRsp = getSMARTResponse(testSmartResp)
	th.Assert(t, len(smartRsp) == 0, fmt.Sprintf("SMART Response should be empty. It is instead %d", len(smartRsp)))
}

func Test_getSupportedResources(t *testing.T) {
	// Base case : all resources are different
	testSupportedResources := []byte(`{
		"read": ["Device", "Encounter"],
		"search-type": ["DiagnosticReport", "DocumentReference"]
	}`)
	supRes := getSupportedResources(testSupportedResources)
	th.Assert(t, len(supRes) == 4, fmt.Sprintf("There should be 4 supported resources, is instead %d", len(supRes)))
	th.Assert(t, helpers.StringArrayContains(supRes, "Device"), "The supported resources should include the 'Device' resource")
	th.Assert(t, helpers.StringArrayContains(supRes, "DocumentReference"), "The supported resources should include the 'DocumentReference' resource")

	// Base case : there are repeated resources
	testSupportedResources = []byte(`{
		"read": ["Device", "DocumentReference"],
		"search-type": ["Device", "DocumentReference"]
	}`)
	supRes = getSupportedResources(testSupportedResources)
	th.Assert(t, len(supRes) == 2, fmt.Sprintf("There should be 2 supported resources, is instead %d", len(supRes)))
	th.Assert(t, helpers.StringArrayContains(supRes, "Device"), "The supported resources should include the 'Device' resource")
	th.Assert(t, helpers.StringArrayContains(supRes, "DocumentReference"), "The supported resources should include the 'DocumentReference' resource")

	// If the value is nonsense, return an empty array
	testSupportedResources = []byte(`null`)
	supRes = getSupportedResources(testSupportedResources)
	th.Assert(t, len(supRes) == 0, fmt.Sprintf("There should be 0 supported resources, is instead %d", len(supRes)))
}
