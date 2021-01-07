package capabilityhandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
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
	"availability":      1.0,
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
		Field:     "rest.mode",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "rest.resource.type",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "rest.resource.interaction.code",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "rest.resource.versioning",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "rest.resource.conditionalRead",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "rest.resource.conditionalDelete",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "rest.resource.referencePolicy",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "rest.resource.searchParam.type",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "rest.interaction.code",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "messaging.supportedMessage.mode",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "document.mode",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "conformance-supported-system",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "conformance-search-parameter-combination",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "DSTU2-oauth-uris",
		Exists:    false,
		Extension: true,
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
		Field:     "conformance-expectation",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "conformance-prohibited",
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

var testFhirEndpointMetadata = endpointmanager.FHIREndpointMetadata{
	URL:               "http://example.com/DTSU2/",
	HTTPResponse:      200,
	Errors:            "",
	SMARTHTTPResponse: 0,
	ResponseTime:      0.1234,
	Availability:      1.0,
}

var testFhirEndpointInfo = endpointmanager.FHIREndpointInfo{
	URL:                "http://example.com/DTSU2/",
	MIMETypes:          []string{"application/json+fhir"},
	TLSVersion:         "TLS 1.2",
	SMARTResponse:      nil,
	Validation:         testValidationObj,
	IncludedFields:     testIncludedFields,
	SupportedResources: testSupportedResources,
	Metadata:           testFhirEndpointMetadata,
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
	// formatMessage does not check for availability field in JSON because availability is written by a trigger
	endpt.Metadata.Availability = 1.0
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

	// test incorrect smart http response
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
	th.Assert(t, includedFields[8].Exists == false, "Expected contact in includedFields to be false, was true")
	th.Assert(t, includedFields[18].Exists == false, "Expected software.name in includedFields to be false, was true")
	th.Assert(t, includedFields[25].Exists == true, "Expected format in includedFields to be true, was false")
	th.Assert(t, includedFields[32].Exists == true, "Expected rest.mode in includedFields to be true, was false")
	th.Assert(t, includedFields[37].Exists == false, "Expected rest.resource.conditionalDelete in includedFields to be false, was true")
	th.Assert(t, includedFields[39].Exists == true, "Expected rest.resource.searchParam.type in includedFields to be true, was false")
	th.Assert(t, includedFields[55].Exists == false, "Expected conformance expectation extension in includedFields to be false, was true")
	th.Assert(t, includedFields[50].Exists == true, "Expected oauth-uris extension in includedFields to be true, was false")
	th.Assert(t, includedFields[46].Exists == false, "Expected capabilities extension in includedFields to be false, was true")

	setupCapabilityStatement(t, filepath.Join("../../testdata", "wellstar_capability_tester.json"))
	capInt = testQueueMsg["capabilityStatement"].(map[string]interface{})
	includedFields = RunIncludedFieldsAndExtensionsChecks(capInt)

	th.Assert(t, includedFields[0].Exists == true, "Expected url in includedFields to be true, was false")
	th.Assert(t, includedFields[2].Exists == false, "Expected name in includedFields to be false, was true")
	th.Assert(t, includedFields[8].Exists == true, "Expected contact in includedFields to be true, was false")
	th.Assert(t, includedFields[18].Exists == true, "Expected software.name in includedFields to be true, was false")
	th.Assert(t, includedFields[19].Exists == true, "Expected software.version in includedFields to be true, was false")
	th.Assert(t, includedFields[25].Exists == true, "Expected format in includedFields to be true, was false")
	th.Assert(t, includedFields[34].Exists == true, "Expected rest.resource.interaction.code in includedFields to be true, was false")
	th.Assert(t, includedFields[37].Exists == true, "Expected rest.resource.conditionalDelete in includedFields to be true, was false")
	th.Assert(t, includedFields[42].Exists == false, "Expected document.mode in includedFields to be false, was true")
	th.Assert(t, includedFields[57].Exists == false, "Expected capabilitystatement expectation extension in includedFields to be false, was true")
	th.Assert(t, includedFields[46].Exists == false, "Expected capabilities extension in includedFields to be false, was true")
	th.Assert(t, includedFields[50].Exists == true, "Expected oauth-uris extension in includedFields to be true, was false")
	th.Assert(t, includedFields[52].Exists == false, "Expected resource-approvalDate extension in includedFields to be false, was true")

	//Testing for R4 Capability Statement extensions where all extensions present
	setupCapabilityStatement(t, filepath.Join("../../testdata", "test_r4_capability_statement_extensions.json"))
	capInt = testQueueMsg["capabilityStatement"].(map[string]interface{})
	includedFields = RunIncludedFieldsAndExtensionsChecks(capInt)

	th.Assert(t, includedFields[46].Exists == true, "Expected capabilities extension in includedFields to be true, was false")
	th.Assert(t, includedFields[46].Field == "capabilities", fmt.Sprintf("Expected field to be capabilities, was %s", includedFields[46].Field))
	th.Assert(t, includedFields[47].Exists == true, "Expected capabilitystatement-search-parameter-combination extension in includedFields to be true, was false")
	th.Assert(t, includedFields[47].Field == "capabilitystatement-search-parameter-combination", fmt.Sprintf("Expected field to be capabilitystatement-search-parameter-combination, was %s", includedFields[47].Field))
	th.Assert(t, includedFields[48].Exists == true, "Expected capabilitystatement-supported-system extension in includedFields to be true, was false")
	th.Assert(t, includedFields[48].Field == "capabilitystatement-supported-system", fmt.Sprintf("Expected field to be capabilitystatement-supported-system, was %s", includedFields[48].Field))
	th.Assert(t, includedFields[49].Exists == true, "Expected capabilitystatement-websocket extension in includedFields to be true, was false")
	th.Assert(t, includedFields[49].Field == "capabilitystatement-websocket", fmt.Sprintf("Expected field to be capabilitystatement-websocket extension, was %s", includedFields[49].Field))
	th.Assert(t, includedFields[50].Exists == true, "Expected oauth-uris extension in includedFields to be true, was false")
	th.Assert(t, includedFields[50].Field == "oauth-uris", fmt.Sprintf("Expected field to be oauth-uris, was %s", includedFields[50].Field))
	th.Assert(t, includedFields[51].Exists == true, "Expected replaces extension in includedFields to be true, was false")
	th.Assert(t, includedFields[51].Field == "replaces", fmt.Sprintf("Expected field to be replaces, was %s", includedFields[51].Field))
	th.Assert(t, includedFields[52].Exists == true, "Expected resource-approvalDate extension in includedFields to be true, was false")
	th.Assert(t, includedFields[52].Field == "resource-approvalDate", fmt.Sprintf("Expected field to be resource-approvalDate, was %s", includedFields[52].Field))
	th.Assert(t, includedFields[53].Exists == true, "Expected resource-effectivePeriod extension in includedFields to be true, was false")
	th.Assert(t, includedFields[53].Field == "resource-effectivePeriod", fmt.Sprintf("Expected field to be resource-effectivePeriod, was %s", includedFields[53].Field))
	th.Assert(t, includedFields[54].Exists == true, "Expected resource-lastReviewDate extension in includedFields to be true, was false")
	th.Assert(t, includedFields[54].Field == "resource-lastReviewDate", fmt.Sprintf("Expected field to be resource-lastReviewDate, was %s", includedFields[54].Field))
	th.Assert(t, includedFields[57].Exists == true, "Expected capabilitystatement-expectation extension in includedFields to be true, was false")
	th.Assert(t, includedFields[57].Field == "capabilitystatement-expectation", fmt.Sprintf("Expected field to be capabilitystatement-expectation, was %s", includedFields[57].Field))
	th.Assert(t, includedFields[58].Exists == true, "Expected capabilitystatement-prohibited extension in includedFields to be true, was false")
	th.Assert(t, includedFields[58].Field == "capabilitystatement-prohibited", fmt.Sprintf("Expected field to be capabilitystatement-prohibited, was %s", includedFields[58].Field))

	// Test for additional nested included fields that appear in r4_capability_statement
	th.Assert(t, includedFields[35].Exists == true, "Expected rest.resource.versioning in includedFields to be true, was false")
	th.Assert(t, includedFields[36].Exists == true, "Expected rest.resource.conditionalRead in includedFields to be true, was false")
	th.Assert(t, includedFields[40].Exists == true, "Expected rest.interaction.code in includedFields to be true, was false")
	th.Assert(t, includedFields[41].Exists == true, "Expected messaging.supportedMessage.mode in includedFields to be true, was false")
	th.Assert(t, includedFields[42].Exists == true, "Expected document.mode in includedFields to be true, was false")

	//Testing for DSTU2 Capability Statement extensions where all extensions present
	setupCapabilityStatement(t, filepath.Join("../../testdata", "test_cerner_capability_dstu2_extensions.json"))
	capInt = testQueueMsg["capabilityStatement"].(map[string]interface{})
	includedFields = RunIncludedFieldsAndExtensionsChecks(capInt)

	th.Assert(t, includedFields[43].Exists == true, "Expected conformance-supported-system extension in includedFields to be true, was false")
	th.Assert(t, includedFields[43].Field == "conformance-supported-system", fmt.Sprintf("Expected field to be conformance-supported-system, was %s", includedFields[43].Field))
	th.Assert(t, includedFields[44].Exists == true, "Expected conformance-search-parameter-combination extension in includedFields to be true, was false")
	th.Assert(t, includedFields[44].Field == "conformance-search-parameter-combination", fmt.Sprintf("Expected field to be conformance-search-parameter-combination, was %s", includedFields[44].Field))
	th.Assert(t, includedFields[45].Exists == true, "Expected DSTU2-oauth-uris extension in includedFields to be true, was false")
	th.Assert(t, includedFields[45].Field == "DSTU2-oauth-uris", fmt.Sprintf("Expected field to be DSTU2-oauth-uris, was %s", includedFields[45].Field))
	th.Assert(t, includedFields[55].Exists == true, "Expected conformance-expectation extension in includedFields to be true, was false")
	th.Assert(t, includedFields[55].Field == "conformance-expectation", fmt.Sprintf("Expected field to be conformance-expectation, was %s", includedFields[55].Field))
	th.Assert(t, includedFields[56].Exists == true, "Expected conformance-prohibited extension in includedFields to be true, was false")
	th.Assert(t, includedFields[56].Field == "conformance-prohibited", fmt.Sprintf("Expected field to be conformance-prohibited extension, was %s", includedFields[56].Field))
}

func Test_RunSupportedResourcesChecks(t *testing.T) {
	setupCapabilityStatement(t, filepath.Join("../../testdata", "cerner_capability_dstu2.json"))
	capInt := testQueueMsg["capabilityStatement"].(map[string]interface{})
	supportedResources := RunSupportedResourcesChecks(capInt)
	th.Assert(t, len(supportedResources) == 27, fmt.Sprintf("Expected there to be 27 supported resources in supportedResources array, were %v", len(supportedResources)))
	th.Assert(t, helpers.StringArrayContains(supportedResources, "ProcedureRequest"), "Expected supportedResources to contain ProcedureRequest resource type")
	th.Assert(t, helpers.StringArrayContains(supportedResources, "MedicationStatement"), "Expected supportedResources to contain MedicationStatement resource type")
	th.Assert(t, !helpers.StringArrayContains(supportedResources, "other"), "Did not expect supportedResources to contain other resource type")
}
