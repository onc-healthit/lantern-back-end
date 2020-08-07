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
		Field:     "url",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "version",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "name",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "title",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "status",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "experimental",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "date",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "publisher",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "contact",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "description",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "requirements",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "useContext",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "jurisdiction",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "purpose",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "copyright",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "kind",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "instantiates",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "imports",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "software.name",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "software.version",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "software.releaseDate",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "implementation.description",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "implementation.url",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "implementation.custodian",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "fhirVersion",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "format",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "patchFormat",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "acceptUnknown",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "implementationGuide",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "profile",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "messaging",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "document",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "capabilities",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "capabilitystatement-search-parameter-combination",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "capabilitystatement-supported-system",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "capabilitystatement-websocket",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "oauth-uris",
		Exists:    true,
		Extension: true,
	},
	{
		Field:     "replaces",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "resource-approvalDate",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "resource-effectivePeriod",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "resource-lastReviewDate",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "capabilitystatement-expectation",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "capabilitystatement-prohibited",
		Exists:    false,
		Extension: true,
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

func Test_RunIncludedFieldsAndExtensionsChecks(t *testing.T) {
	setupCapabilityStatement(t, filepath.Join("../../testdata", "cerner_capability_dstu2.json"))
	capInt := testQueueMsg["capabilityStatement"].(map[string]interface{})
	includedFields := RunIncludedFieldsAndExtensionsChecks(capInt)
	th.Assert(t, includedFields[0].Exists == true, "Expected url in includedFields to be true, was false")
	th.Assert(t, includedFields[2].Exists == true, "Expected name in includedFields to be true, was false")
	th.Assert(t, includedFields[18].Exists == false, "Expected software.name in includedFields to be false, was true")
	th.Assert(t, includedFields[25].Exists == true, "Expected format in includedFields to be true, was false")
	th.Assert(t, includedFields[8].Exists == false, "Expected contact in includedFields to be false, was true")
	th.Assert(t, includedFields[41].Exists == false, "Expected expectation extension in includedFields to be false, was true")
	th.Assert(t, includedFields[36].Exists == true, "Expected oauth-uris extension in includedFields to be true, was false")

	setupCapabilityStatement(t, filepath.Join("../../testdata", "wellstar_capability_tester.json"))
	capInt = testQueueMsg["capabilityStatement"].(map[string]interface{})
	includedFields = RunIncludedFieldsAndExtensionsChecks(capInt)

	th.Assert(t, includedFields[0].Exists == true, "Expected url in includedFields to be true, was false")
	th.Assert(t, includedFields[2].Exists == false, "Expected name in includedFields to be false, was true")
	th.Assert(t, includedFields[18].Exists == true, "Expected software.name in includedFields to be true, was false")
	th.Assert(t, includedFields[19].Exists == true, "Expected software.version in includedFields to be true, was false")
	th.Assert(t, includedFields[25].Exists == true, "Expected format in includedFields to be true, was false")
	th.Assert(t, includedFields[8].Exists == true, "Expected contact in includedFields to be true, was false")
	th.Assert(t, includedFields[32].Exists == false, "Expected capabilities extension in includedFields to be false, was true")
	th.Assert(t, includedFields[36].Exists == true, "Expected oauth-uris extension in includedFields to be true, was false")

	//Testing for extensions where all extensions present
	setupCapabilityStatement(t, filepath.Join("../../testdata", "test_cerner_capability_dstu2_extensionsAdded.json"))
	capInt = testQueueMsg["capabilityStatement"].(map[string]interface{})
	includedFields = RunIncludedFieldsAndExtensionsChecks(capInt)

	th.Assert(t, includedFields[32].Exists == true, "Expected capabilities extension in includedFields to be true, was false")
	th.Assert(t, includedFields[32].Field == "capabilities", fmt.Sprintf("Expected field to be capabilities, was %s", includedFields[32].Field))
	th.Assert(t, includedFields[33].Exists == true, "Expected capabilitystatement-search-parameter-combination extension in includedFields to be true, was false")
	th.Assert(t, includedFields[33].Field == "capabilitystatement-search-parameter-combination", fmt.Sprintf("Expected field to be capabilitystatement-search-parameter-combination, was %s", includedFields[33].Field))
	th.Assert(t, includedFields[34].Exists == true, "Expected capabilitystatement-supported-system extension in includedFields to be true, was false")
	th.Assert(t, includedFields[34].Field == "capabilitystatement-supported-system", fmt.Sprintf("Expected field to be capabilitystatement-supported-system, was %s", includedFields[34].Field))
	th.Assert(t, includedFields[35].Exists == true, "Expected capabilitystatement-websocket extension in includedFields to be true, was false")
	th.Assert(t, includedFields[35].Field == "capabilitystatement-websocket", fmt.Sprintf("Expected field to be capabilitystatement-websocket extension, was %s", includedFields[35].Field))
	th.Assert(t, includedFields[36].Exists == true, "Expected oauth-uris extension in includedFields to be true, was false")
	th.Assert(t, includedFields[36].Field == "oauth-uris", fmt.Sprintf("Expected field to be oauth-uris, was %s", includedFields[36].Field))
	th.Assert(t, includedFields[37].Exists == true, "Expected replaces extension in includedFields to be true, was false")
	th.Assert(t, includedFields[37].Field == "replaces", fmt.Sprintf("Expected field to be replaces, was %s", includedFields[37].Field))
	th.Assert(t, includedFields[38].Exists == true, "Expected resource-approvalDate extension in includedFields to be true, was false")
	th.Assert(t, includedFields[38].Field == "resource-approvalDate", fmt.Sprintf("Expected field to be resource-approvalDate, was %s", includedFields[38].Field))
	th.Assert(t, includedFields[39].Exists == true, "Expected resource-effectivePeriod extension in includedFields to be true, was false")
	th.Assert(t, includedFields[39].Field == "resource-effectivePeriod", fmt.Sprintf("Expected field to be resource-effectivePeriod, was %s", includedFields[39].Field))
	th.Assert(t, includedFields[40].Exists == true, "Expected resource-lastReviewDate extension in includedFields to be true, was false")
	th.Assert(t, includedFields[40].Field == "resource-lastReviewDate", fmt.Sprintf("Expected field to be resource-lastReviewDate, was %s", includedFields[40].Field))
	th.Assert(t, includedFields[41].Exists == true, "Expected capabilitystatement-expectation extension in includedFields to be true, was false")
	th.Assert(t, includedFields[41].Field == "capabilitystatement-expectation", fmt.Sprintf("Expected field to be capabilitystatement-expectation, was %s", includedFields[41].Field))
	th.Assert(t, includedFields[42].Exists == true, "Expected capabilitystatement-prohibited extension in includedFields to be true, was false")
	th.Assert(t, includedFields[42].Field == "capabilitystatement-prohibited", fmt.Sprintf("Expected field to be capabilitystatement-prohibited, was %s", includedFields[42].Field))
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
