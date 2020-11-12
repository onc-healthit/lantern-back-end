package jsonexport

import (
	"fmt"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_getFHIRVersion(t *testing.T) {
	// Base case

	testCapStat := []byte(`
	{
		"fhirVersion": "4.0.1",
		"kind": "instance"
	}`)
	fhirVersion := getFHIRVersion(testCapStat)
	th.Assert(t, fhirVersion == "4.0.1", fmt.Sprintf("FHIR Version in capability statement should be 4.0.1. It is instead %s", fhirVersion))

	// If capability statement is nonsense, fhirVersion should be an empty string

	testCapStat = []byte("this is not JSON")
	fhirVersion = getFHIRVersion(testCapStat)
	th.Assert(t, fhirVersion == "", fmt.Sprintf("FHIR Version should be an empty string. It is instead %s", fhirVersion))

	// If capability statement is null, fhirVersion should be an empty string
	testCapStat = []byte("null")
	fhirVersion = getFHIRVersion(testCapStat)
	th.Assert(t, fhirVersion == "", fmt.Sprintf("FHIR Version should be an empty string. It is instead %s", fhirVersion))

}

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
