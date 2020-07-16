package capabilityhandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

var testQueueMsg = map[string]interface{}{
	"url":               "http://example.com/DTSU2/",
	"err":               "",
	"mimeTypes":         []string{"application/json+fhir"},
	"httpResponse":      200,
	"tlsVersion":        "TLS 1.2",
	"smarthttpResponse": 0,
	"smartResp":         nil,
	"responseTime":      0.1234,
}

var testValidationObj = endpointmanager.Validation{
	Results: []endpointmanager.Rule{
		{
			RuleName: endpointmanager.CapStatExistRule,
			Valid:    true,
			Expected: "true",
			Actual:   "true",
			Comment:  "The Capability Statement exists.",
		},
	},
	Warnings: []endpointmanager.Rule{},
}

var testIncludedFields = []endpointmanager.IncludedField{
	{
		Field:  "url",
		Exists: true,
	},
	{
		Field:  "version",
		Exists: false,
	},
	{
		Field:  "name",
		Exists: true,
	},
	{
		Field:  "title",
		Exists: false,
	},
	{
		Field:  "status",
		Exists: true,
	},
	{
		Field:  "experimental",
		Exists: false,
	},
	{
		Field:  "date",
		Exists: true,
	},
	{
		Field:  "publisher",
		Exists: true,
	},
	{
		Field:  "contact",
		Exists: false,
	},
	{
		Field:  "description",
		Exists: true,
	},
	{
		Field:  "requirements",
		Exists: false,
	},
	{
		Field:  "useContext",
		Exists: false,
	},
	{
		Field:  "jurisdiction",
		Exists: false,
	},
	{
		Field:  "purpose",
		Exists: false,
	},
	{
		Field:  "copyright",
		Exists: false,
	},
	{
		Field:  "kind",
		Exists: true,
	},
	{
		Field:  "instantiates",
		Exists: false,
	},
	{
		Field:  "imports",
		Exists: false,
	},
	{
		Field:  "software.name",
		Exists: false,
	},
	{
		Field:  "software.version",
		Exists: false,
	},
	{
		Field:  "software.releaseDate",
		Exists: false,
	},
	{
		Field:  "implementation.description",
		Exists: false,
	},
	{
		Field:  "implementation.url",
		Exists: false,
	},
	{
		Field:  "implementation.custodian",
		Exists: false,
	},
	{
		Field:  "fhirVersion",
		Exists: true,
	},
	{
		Field:  "format",
		Exists: true,
	},
	{
		Field:  "patchFormat",
		Exists: false,
	},
	{
		Field:  "acceptUnknown",
		Exists: true,
	},
	{
		Field:  "implementationGuide",
		Exists: false,
	},
	{
		Field:  "profile",
		Exists: false,
	},
	{
		Field:  "messaging",
		Exists: false,
	},
	{
		Field:  "document",
		Exists: false,
	},
}

var testSupportedResources = []string{
	"Conformance",
	"AllergyIntolerance",
	"Appointment",
	"Binary",
	"CarePlan",
	"Condition",
	"Contract",
	"Device",
	"DiagnosticReport",
	"DocumentReference",
	"Encounter",
	"Goal",
	"Immunization",
	"MedicationAdministration",
	"MedicationOrder",
	"MedicationStatement",
	"Observation",
	"OperationDefinition",
	"Patient",
	"Person",
	"Practitioner",
	"Procedure",
	"ProcedureRequest",
	"RelatedPerson",
	"Schedule",
	"Slot",
	"StructureDefinition"}

var testFhirEndpointInfo = endpointmanager.FHIREndpointInfo{
	URL:                "http://example.com/DTSU2/",
	MIMETypes:          []string{"application/json+fhir"},
	TLSVersion:         "TLS 1.2",
	HTTPResponse:       200,
	Errors:             "",
	SMARTHTTPResponse:  0,
	SMARTResponse:      nil,
	Validation:         testValidationObj,
	IncludedFields:     testIncludedFields,
	SupportedResources: testSupportedResources,
	ResponseTime:       0.1234,
}

// Convert the test Queue Message into []byte format for testing purposes
func convertInterfaceToBytes(message map[string]interface{}) ([]byte, error) {
	returnMsg, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return returnMsg, nil
}

func setupCapabilityStatement(t *testing.T, path string) {
	// capability statement
	csJSON, err := ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)
	testFhirEndpointInfo.CapabilityStatement = cs
	var capStat map[string]interface{}
	err = json.Unmarshal(csJSON, &capStat)
	th.Assert(t, err == nil, err)
	testQueueMsg["capabilityStatement"] = capStat
}

func Test_formatMessage(t *testing.T) {
	setupCapabilityStatement(t, filepath.Join("../../testdata", "cerner_capability_dstu2.json"))
	expectedEndpt := testFhirEndpointInfo
	tmpMessage := testQueueMsg

	message, err := convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)

	// basic test
	endpt, returnErr := formatMessage(message)
	th.Assert(t, returnErr == nil, returnErr)
	// Just check that the first validation field is valid
	endpt.Validation.Results = []endpointmanager.Rule{endpt.Validation.Results[0]}
	th.Assert(t, expectedEndpt.Equal(endpt), "An error was thrown because the endpoints are not equal")

	// should not throw error if metadata is not in the URL
	tmpMessage["url"] = "http://example.com/DTSU2/"
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr == nil, "An error was thrown because metadata was not included in the url")

	// test incorrect error message
	tmpMessage["err"] = nil
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to an incorrect error message")
	tmpMessage["err"] = ""

	// test incorrect URL
	tmpMessage["url"] = nil
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to an incorrect URL")

	tmpMessage["url"] = "http://example.com/DTSU2/"
	// test incorrect TLS Version
	tmpMessage["tlsVersion"] = 1
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to an incorrect TLS Version")
	tmpMessage["tlsVersion"] = "TLS 1.2"

	// test incorrect MIME Type
	tmpMessage["mimeTypes"] = 1
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to incorrect MIME Types")
	tmpMessage["mimeTypes"] = []string{"application/json+fhir"}

	// test incorrect http response
	tmpMessage["httpResponse"] = "200"
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to an incorrect HTTP response")
	tmpMessage["httpResponse"] = 200

	// test incorrect http response
	tmpMessage["smarthttpResponse"] = "200"
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to an incorrect smart HTTP response")
	tmpMessage["smarthttpResponse"] = 200

	// test incorrect response time
	tmpMessage["responseTime"] = "0.1234"
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to an incorrect responseTime")
	tmpMessage["responseTime"] = 0.1234
}

func Test_RunIncludedFieldsChecks(t *testing.T) {
	setupCapabilityStatement(t, filepath.Join("../../testdata", "cerner_capability_dstu2.json"))
	capInt := testQueueMsg["capabilityStatement"].(map[string]interface{})
	includedFields := RunIncludedFieldsChecks(capInt)
	th.Assert(t, includedFields[0].Exists == true, "Expected url in includedFields to be true, was false")
	th.Assert(t, includedFields[2].Exists == true, "Expected name in includedFields to be true, was false")
	th.Assert(t, includedFields[18].Exists == false, "Expected software.name in includedFields to be false, was true")
	th.Assert(t, includedFields[25].Exists == true, "Expected format in includedFields to be true, was false")
	th.Assert(t, includedFields[8].Exists == false, "Expected contact in includedFields to be false, was true")

	setupCapabilityStatement(t, filepath.Join("../../testdata", "wellstar_capability_tester.json"))
	capInt = testQueueMsg["capabilityStatement"].(map[string]interface{})
	includedFields = RunIncludedFieldsChecks(capInt)

	th.Assert(t, includedFields[0].Exists == true, "Expected url in includedFields to be true, was false")
	th.Assert(t, includedFields[2].Exists == false, "Expected name in includedFields to be false, was true")
	th.Assert(t, includedFields[18].Exists == true, "Expected software.name in includedFields to be true, was false")
	th.Assert(t, includedFields[19].Exists == true, "Expected software.version in includedFields to be true, was false")
	th.Assert(t, includedFields[25].Exists == true, "Expected format in includedFields to be true, was false")
	th.Assert(t, includedFields[8].Exists == true, "Expected contact in includedFields to be true, was false")
}

func Test_RunSupportedResourcesChecks(t *testing.T) {
	setupCapabilityStatement(t, filepath.Join("../../testdata", "cerner_capability_dstu2.json"))
	capInt := testQueueMsg["capabilityStatement"].(map[string]interface{})
	supportedResources := RunSupportedResourcesChecks(capInt)
	th.Assert(t, len(supportedResources) == 27, fmt.Sprintf("Expected there to be 27 supported resources in supportedResources array, were %v", len(supportedResources)))
	th.Assert(t, contains(supportedResources, "ProcedureRequest"), "Expected supportedResources to contain ProcedureRequest resource type")
	th.Assert(t, contains(supportedResources, "MedicationStatement"), "Expected supportedResources to contain MedicationStatement resource type")
	th.Assert(t, !contains(supportedResources, "other"), "Did not expect supportedResources to contain other resource type")
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
