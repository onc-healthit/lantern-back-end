package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_RunValidation(t *testing.T) {
	// base test

	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, dstu2)
	th.Assert(t, err == nil, err)

	expectedFirstVal := endpointmanager.Rule{
		RuleName: endpointmanager.CapStatExistRule,
		Valid:    true,
		Expected: "true",
		Actual:   "true",
		Comment:  "The Capability Statement exists.",
	}
	expectedLastVal := endpointmanager.Rule{
		RuleName: endpointmanager.KindRule,
		Valid:    true,
		Expected: "instance",
		Comment:  "Kind value should be set to 'instance' because this is a specific system instance.",
		Actual:   "instance",
	}

	requestedFhirVersion := "None"
	defaultFhirVersion := "1.0.2"

	actualVal := validator.RunValidation(cs, 200, []string{fhir2LessJSONMIMEType}, "1.0.2", "TLS 1.2", 200, requestedFhirVersion, defaultFhirVersion)
	th.Assert(t, len(actualVal.Results) == 5, fmt.Sprintf("RunValidation should have returned 5 validation checks, instead it returned %d", len(actualVal.Results)))
	eq := reflect.DeepEqual(actualVal.Results[0], expectedFirstVal)
	th.Assert(t, eq == true, "RunValidation's first returned validation is not correct")
	eq = reflect.DeepEqual(actualVal.Results[4], expectedLastVal)
	th.Assert(t, eq == true, "RunValidation's last returned validation is not correct")

	// r4 test

	cs2, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, r4)
	th.Assert(t, err == nil, err)

	// choose two random validation values in the list to check

	expectedSixthVal := endpointmanager.Rule{
		RuleName:  endpointmanager.TLSVersion,
		Valid:     true,
		Expected:  "TLS 1.2, TLS 1.3",
		Actual:    "TLS 1.2",
		Comment:   "Systems SHALL use TLS version 1.2 or higher for all transmissions not taking place over a secure network connection.",
		Reference: "https://www.hl7.org/fhir/us/core/security.html",
		ImplGuide: "USCore 3.1",
	}
	expectedLastVal = endpointmanager.Rule{
		RuleName:  endpointmanager.SearchParamsRule,
		Valid:     true,
		Actual:    "true",
		Expected:  "true",
		Comment:   "Search parameter names must be unique in the context of a resource.",
		ImplGuide: "USCore 3.1",
		Reference: "http://hl7.org/fhir/capabilitystatement.html",
	}

	actualVal = validator2.RunValidation(cs2, 200, []string{fhir3PlusJSONMIMEType}, "4.0.1", "TLS 1.2", 200, requestedFhirVersion, defaultFhirVersion)
	th.Assert(t, len(actualVal.Results) == 17, fmt.Sprintf("RunValidation should have returned 17 validation checks, instead it returned %d", len(actualVal.Results)))
	eq = reflect.DeepEqual(actualVal.Results[5], expectedSixthVal)
	th.Assert(t, eq == true, "RunValidation's first returned validation is not correct")
	eq = reflect.DeepEqual(actualVal.Results[16], expectedLastVal)
	th.Assert(t, eq == true, "RunValidation's last returned validation is not correct")
}

func Test_CapStatExists(t *testing.T) {
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, dstu2)
	th.Assert(t, err == nil, err)

	// base test

	expectedCap := endpointmanager.Rule{
		RuleName: endpointmanager.CapStatExistRule,
		Valid:    true,
		Expected: "true",
		Actual:   "true",
		Comment:  "The Capability Statement exists.",
	}

	actualCap := validator.CapStatExists(cs)
	eq := reflect.DeepEqual(actualCap, expectedCap)
	th.Assert(t, eq == true, fmt.Sprintf("DSTU2 Capability Statement should exist, returned value is instead %+v", actualCap))

	// capability statement does not exist

	expectedCap2 := endpointmanager.Rule{
		RuleName: endpointmanager.CapStatExistRule,
		Valid:    false,
		Expected: "true",
		Actual:   "false",
		Comment:  "The Capability Statement does not exist.",
	}

	actualCap = validator.CapStatExists(nil)
	eq = reflect.DeepEqual(actualCap, expectedCap2)
	th.Assert(t, eq == true, fmt.Sprintf("Capability Statement should not exist, returned value is instead %+v", actualCap))

	// r4 test

	cs2, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, r4)
	th.Assert(t, err == nil, err)

	expectedCap.Comment = "Servers SHALL provide a Capability Statement that specifies which interactions and resources are supported."
	expectedCap.Reference = "http://hl7.org/fhir/http.html"
	expectedCap.ImplGuide = "USCore 3.1"
	actualCap = validator2.CapStatExists(cs2)
	eq = reflect.DeepEqual(actualCap, expectedCap)
	th.Assert(t, eq == true, fmt.Sprintf("R4 Capability Statement should exist, returned value is instead %+v", actualCap))
}

func Test_MimeTypeValid(t *testing.T) {
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, dstu2)
	th.Assert(t, err == nil, err)

	// base test

	expectedVal := endpointmanager.Rule{
		RuleName: endpointmanager.GeneralMimeTypeRule,
		Valid:    true,
		Expected: fhir2LessJSONMIMEType,
		Actual:   fhir2LessJSONMIMEType,
		Comment:  "FHIR Version 1.0.2 requires the Mime Type to be application/json+fhir",
	}

	actualVal := validator.MimeTypeValid([]string{fhir2LessJSONMIMEType}, "1.0.2")
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The given mime type for version DSTU2 should be valid, is instead %+v", actualVal))

	// fhirVersion 3+ test

	expectedVal.Expected = fhir3PlusJSONMIMEType
	expectedVal.Actual = fhir3PlusJSONMIMEType
	expectedVal.Comment = "FHIR Version 3.0.0 requires the Mime Type to be " + fhir3PlusJSONMIMEType

	actualVal = validator.MimeTypeValid([]string{fhir3PlusJSONMIMEType}, "3.0.0")
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The given mime type for version STU3 should be valid, is instead %+v", actualVal))

	// no mime types

	expectedVal.Valid = false
	expectedVal.Expected = "N/A"
	expectedVal.Actual = ""
	expectedVal.Comment = "No mime type given; cannot validate mime type."

	actualVal = validator.MimeTypeValid([]string{}, "1.0.2")
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("There is no given mime type so the check should be invalid, is instead %+v", actualVal))

	// no version

	expectedVal.Actual = fhir2LessJSONMIMEType
	expectedVal.Comment = "Unknown FHIR Version; cannot validate mime type."
	actualVal = validator.MimeTypeValid([]string{fhir2LessJSONMIMEType}, "")
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("There is no given FHIR version so the check should be invalid, is instead %+v", actualVal))

	// mixmatch mime type and version

	expectedVal.Expected = fhir3PlusJSONMIMEType
	expectedVal.Comment = "FHIR Version 3.0.0 requires the Mime Type to be " + fhir3PlusJSONMIMEType
	actualVal = validator.MimeTypeValid([]string{fhir2LessJSONMIMEType}, "3.0.0")
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Mime type %s should not be valid for version 3.0.0", fhir2LessJSONMIMEType))

	// r4 test

	cs2, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Valid = true
	expectedVal.Actual = fhir3PlusJSONMIMEType
	expectedVal.Comment = "FHIR Version 4.0.1 requires the Mime Type to be " + fhir3PlusJSONMIMEType
	expectedVal.Reference = "http://hl7.org/fhir/http.html"
	expectedVal.ImplGuide = "USCore 3.1"
	actualVal = validator2.MimeTypeValid([]string{fhir3PlusJSONMIMEType}, "4.0.1")
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The given mime type for version R4 should be valid, returned value is instead %+v", actualVal))
}

func Test_HTTPResponseValid(t *testing.T) {
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, dstu2)
	th.Assert(t, err == nil, err)

	// base test

	expectedVal := endpointmanager.Rule{
		RuleName: endpointmanager.HTTPResponseRule,
		Valid:    true,
		Expected: "200",
		Actual:   "200",
		Comment:  "",
	}

	actualVal := validator.HTTPResponseValid(200)
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The http response is 200 and should be valid, is instead %+v", actualVal))

	// httpResponse is 0

	expectedVal.Valid = false
	expectedVal.Actual = "0"
	expectedVal.Comment = "The GET request failed with no returned HTTP response status code."

	actualVal = validator.HTTPResponseValid(0)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The http response is 0 and should be valid, is instead %+v", actualVal))

	// httpResponse is 404

	expectedVal.Actual = "404"
	expectedVal.Comment = "The HTTP response code was 404 instead of 200. "

	actualVal = validator.HTTPResponseValid(404)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The http response is 404 and should be validd, is instead %+v", actualVal))

	// r4 test

	cs2, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Valid = true
	expectedVal.Actual = "200"
	expectedVal.Comment = "Applications SHALL return a resource that describes the functionality of the server end-point."
	expectedVal.Reference = "http://hl7.org/fhir/http.html"
	expectedVal.ImplGuide = "USCore 3.1"
	actualVal = validator2.HTTPResponseValid(200)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The http response for R4 validator is 200 and should be valid, returned value is instead %+v", actualVal))
}

func Test_FhirVersion(t *testing.T) {
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, dstu2)
	th.Assert(t, err == nil, err)

	// base test

	expectedVal := endpointmanager.Rule{
		RuleName:  endpointmanager.FHIRVersion,
		Valid:     true,
		Expected:  "4.0.1",
		Comment:   "ONC Certification Criteria requires support of FHIR Version 4.0.1",
		Reference: "https://www.healthit.gov/cures/sites/default/files/cures/2020-03/APICertificationCriterion.pdf",
		ImplGuide: "USCore 3.1",
	}
	expectedVal.Actual = "4.0.1"

	actualVal := validator.FhirVersion("4.0.1")
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Fhir version 4.0.1 should be valid, is instead %+v", actualVal))

	// fhirVersion is not 4.0.1

	expectedVal.Actual = "1.0.2"
	expectedVal.Valid = false
	actualVal = validator.FhirVersion("1.0.2")
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Fhir version 1.0.2 should be valid, is instead %+v", actualVal))
}

func Test_checkResourceList(t *testing.T) {
	// checkResourceList is private and therefore cannot be accessed, but most of it's edge cases
	// are not unique to each public function that calls it, so it'll be tested here
	// using the PatientResourceExists function

	cs, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, r4)
	th.Assert(t, err == nil, err)

	// capability statement does not exist

	expectedVal := endpointmanager.Rule{
		RuleName:  endpointmanager.PatResourceExists,
		Valid:     false,
		Actual:    "false",
		Expected:  "true",
		Comment:   "The Capability Statement does not exist; cannot check resource profiles. The US Core Server SHALL support the US Core Patient resource profile.",
		ImplGuide: "USCore 3.1",
		Reference: "https://www.hl7.org/fhir/us/core/CapabilityStatement-us-core-server.html",
	}
	actualVal := validator.PatientResourceExists(nil)
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, "PatientResourceExists check should be invalid because capability statement does not exist.")

	// type is not formatted properly

	cs2, err := nLevelNestedValueChange(cs, []string{"rest", "resource", "type"}, []int{0, 0}, 2, badFormat, "")
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Comment = "The Resource Profiles are not properly formatted. The US Core Server SHALL support the US Core Patient resource profile."
	actualVal = validator2.PatientResourceExists(cs2)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("PatientResourceExists check should be invalid because type is malformed, is instead %+v", actualVal))

	// type does not exist

	cs3, err := nLevelNestedValueChange(cs, []string{"rest", "resource", "type"}, []int{0, 0}, 2, deleteField, "")
	th.Assert(t, err == nil, err)

	validator3, err := getValidator(cs3, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Comment = "The Resource Profiles are not properly formatted. The US Core Server SHALL support the US Core Patient resource profile."
	actualVal = validator3.PatientResourceExists(cs3)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("PatientResourceExists check should be invalid because type does not exist, is instead %+v", actualVal))

	// resourceList does not exist

	cs4, err := nLevelNestedValueChange(cs, []string{"rest", "resource"}, []int{0}, 1, deleteField, "")
	th.Assert(t, err == nil, err)

	validator4, err := getValidator(cs4, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Comment = "The Resource Profiles do not exist. The US Core Server SHALL support the US Core Patient resource profile."
	actualVal = validator4.PatientResourceExists(cs4)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("PatientResourceExists check should be invalid because resources do not exist, is instead %+v", actualVal))

	// restList does not exist

	cs5, err := deleteFieldFromCapStat(cs, "rest")
	th.Assert(t, err == nil, err)

	validator5, err := getValidator(cs5, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Comment = "Rest field does not exist. The US Core Server SHALL support the US Core Patient resource profile."
	actualVal = validator5.PatientResourceExists(cs5)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("PatientResourceExists check should be invalid because the rest field does not exist, is instead %+v", actualVal))
}

func Test_PatientResourceExists(t *testing.T) {
	cs, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, r4)
	th.Assert(t, err == nil, err)

	// base test

	expectedVal := endpointmanager.Rule{
		RuleName:  endpointmanager.PatResourceExists,
		Valid:     true,
		Actual:    "true",
		Expected:  "true",
		Comment:   "The US Core Server SHALL support the US Core Patient resource profile.",
		ImplGuide: "USCore 3.1",
		Reference: "https://www.hl7.org/fhir/us/core/CapabilityStatement-us-core-server.html",
	}
	actualVal := validator.PatientResourceExists(cs)
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The Patient Resource exists and validation should be valid, is instead %+v", actualVal))

	// if there is no Patient resource

	cs2, err := nLevelNestedValueChange(cs, []string{"rest", "resource", "type"}, []int{0, 0}, 2, updateString, "unknown")
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Valid = false
	expectedVal.Actual = "false"
	actualVal = validator2.PatientResourceExists(cs2)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The Patient Resource does not exist and validation should be invalid, is instead %+v", actualVal))
}
func Test_OtherResourceExists(t *testing.T) {
	cs, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, r4)
	th.Assert(t, err == nil, err)

	// base test

	expectedVal := endpointmanager.Rule{
		RuleName:  endpointmanager.OtherResourceExists,
		Valid:     true,
		Actual:    "true",
		Expected:  "true",
		Comment:   "The US Core Server SHALL support at least one additional resource profile (besides Patient) from the list of US Core Profiles.",
		ImplGuide: "USCore 3.1",
		Reference: "https://www.hl7.org/fhir/us/core/CapabilityStatement-us-core-server.html",
	}
	actualVal := validator.OtherResourceExists(cs)
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Another resource exists and the check should be valid, is instead %+v", actualVal))

	// if there is no Patient resource

	cs2, err := nLevelNestedValueChange(cs, []string{"rest", "resource", "type"}, []int{0, 1}, 2, updateString, "unknown")
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Valid = false
	expectedVal.Actual = "false"
	actualVal = validator2.OtherResourceExists(cs2)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Another resource does not exist and the check should be invalid, is instead %+v", actualVal))
}

func Test_SmartHTTPResponseValid(t *testing.T) {
	cs, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, r4)
	th.Assert(t, err == nil, err)

	// base test

	baseComment := "FHIR endpoints requiring authorization SHALL serve a JSON document at the location formed by appending /.well-known/smart-configuration to their base URL."
	expectedVal := endpointmanager.Rule{
		RuleName:  endpointmanager.SmartHTTPRespRule,
		Valid:     true,
		Expected:  "200",
		Actual:    "200",
		Comment:   baseComment,
		Reference: "http://www.hl7.org/fhir/smart-app-launch/conformance/index.html",
		ImplGuide: "USCore 3.1",
	}

	actualVal := validator.SmartHTTPResponseValid(200)
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("SMART-on-FHIR response is 200 so it should be valid, is instead %+v", actualVal))

	// httpResponse is 0

	expectedVal.Valid = false
	expectedVal.Actual = "0"

	actualVal = validator.SmartHTTPResponseValid(0)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("SMART-on-FHIR response is 0 so it should be invalid, is instead %+v", actualVal))

	// httpResponse is 404

	expectedVal.Actual = "404"
	expectedVal.Comment = "The HTTP response code was 404 instead of 200. " + baseComment

	actualVal = validator.SmartHTTPResponseValid(404)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("SMART-on-FHIR response is 404 so it should be invalid, is instead %+v", actualVal))
}

func Test_KindValid(t *testing.T) {
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, dstu2)
	th.Assert(t, err == nil, err)

	// base test

	baseComment := "Kind value should be set to 'instance' because this is a specific system instance."
	expectedVal := endpointmanager.Rule{
		RuleName: endpointmanager.KindRule,
		Valid:    true,
		Expected: "instance",
		Comment:  baseComment,
		Actual:   "instance",
	}
	expectedArray := []endpointmanager.Rule{
		expectedVal,
	}

	actualVal := validator.KindValid(cs)
	eq := reflect.DeepEqual(actualVal, expectedArray)
	th.Assert(t, eq == true, fmt.Sprintf("Kind value should equal instance, is instead %+v", actualVal))

	// cap stat is nil

	expectedVal.Valid = false
	expectedVal.Actual = ""
	expectedVal.Comment = "Capability Statement does not exist; cannot check kind value. " + baseComment
	expectedArray = []endpointmanager.Rule{
		expectedVal,
	}

	actualVal = validator.KindValid(nil)
	eq = reflect.DeepEqual(actualVal, expectedArray)
	th.Assert(t, eq == true, fmt.Sprintf("Can't check kind when capability statement does not exist, is instead %+v", actualVal))

	// returns invalid if kind does not exist

	cs2, err := deleteFieldFromCapStat(cs, "kind")
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, dstu2)
	th.Assert(t, err == nil, err)

	expectedVal.Comment = baseComment
	expectedArray = []endpointmanager.Rule{
		expectedVal,
	}

	actualVal = validator2.KindValid(cs2)
	eq = reflect.DeepEqual(actualVal, expectedArray)
	th.Assert(t, eq == true, fmt.Sprintf("Malformed kind value should return an invalid check, is instead %+v", actualVal))

	// r4 base test

	cs3, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator3, err := getValidator(cs3, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Valid = true
	expectedVal.Actual = "instance"
	expectedVal.Reference = "http://hl7.org/fhir/capabilitystatement.html"
	expectedVal.ImplGuide = "USCore 3.1"

	expectedInstanceVal := endpointmanager.Rule{
		RuleName:  endpointmanager.InstanceRule,
		Valid:     true,
		Expected:  "true",
		Actual:    "true",
		Comment:   "If kind = instance, implementation must be present. This endpoint must be an instance.",
		Reference: "http://hl7.org/fhir/capabilitystatement.html",
		ImplGuide: "USCore 3.1",
	}

	actualVal = validator3.KindValid(cs3)
	eq = reflect.DeepEqual(actualVal[0], expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("R4 KindValid first check should be valid, is instead %+v", actualVal[0]))
	eq = reflect.DeepEqual(actualVal[1], expectedInstanceVal)
	th.Assert(t, eq == true, fmt.Sprintf("R4 KindValid second check should be valid, is instead %+v", actualVal[1]))

	// if implementation doesn't exist, then the second check is invalid

	cs4, err := deleteFieldFromCapStat(cs3, "implementation")
	th.Assert(t, err == nil, err)

	validator4, err := getValidator(cs4, r4)
	th.Assert(t, err == nil, err)

	expectedInstanceVal.Valid = false
	expectedInstanceVal.Actual = "false"
	actualVal = validator4.KindValid(cs4)
	eq = reflect.DeepEqual(actualVal[1], expectedInstanceVal)
	th.Assert(t, eq == true, fmt.Sprintf("Implementation does not exist so KindValid check should be invalid, is instead %+v", actualVal[1]))
}

func Test_MessagingEndpointValid(t *testing.T) {
	cs, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, r4)
	th.Assert(t, err == nil, err)

	// base test

	baseComment := "Messaging end-point is required (and is only permitted) when a statement is for an implementation. This endpoint must be an implementation."
	expectedVal := endpointmanager.Rule{
		RuleName:  endpointmanager.MessagingEndptRule,
		Valid:     true,
		Expected:  "true",
		Actual:    "true",
		Comment:   baseComment,
		Reference: "http://hl7.org/fhir/capabilitystatement.html",
		ImplGuide: "USCore 3.1",
	}
	actualVal := validator.MessagingEndpointValid(cs)
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Messaging endpoint should exist, is instead %+v", actualVal))

	// Remove messaging endpoint

	cs2, err := nLevelNestedValueChange(cs, []string{"messaging", "endpoint"}, []int{0}, 1, deleteField, "")
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Valid = false
	expectedVal.Actual = "false"
	expectedVal.Comment = "Endpoint field in Messaging does not exist. " + baseComment
	actualVal = validator2.MessagingEndpointValid(cs2)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Removing the messaging endpoint should make check invalid, is instead %+v", actualVal))

	// Remove messaging

	cs3, err := deleteFieldFromCapStat(cs, "messaging")
	th.Assert(t, err == nil, err)

	validator3, err := getValidator(cs3, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Comment = "Messaging does not exist. " + baseComment
	actualVal = validator3.MessagingEndpointValid(cs3)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Removing the messaging field should make check invalid, is instead %+v", actualVal))

	// Remove kind

	cs4, err := deleteFieldFromCapStat(cs, "kind")
	th.Assert(t, err == nil, err)

	validator4, err := getValidator(cs4, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Comment = "Kind value should be set to 'instance' because this is a specific system instance. " + baseComment
	actualVal = validator4.MessagingEndpointValid(cs4)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Removing the kind field should make check invalid, is instead %+v", actualVal))
}

func Test_EndpointFunctionValid(t *testing.T) {
	cs, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, r4)
	th.Assert(t, err == nil, err)

	// base test

	expectedVal := endpointmanager.Rule{
		RuleName:  endpointmanager.EndptFunctionRule,
		Valid:     true,
		Actual:    "rest,messaging,document",
		Expected:  "rest,messaging,document",
		Comment:   "A Capability Statement SHALL have at least one of REST, messaging or document element.",
		Reference: "http://hl7.org/fhir/capabilitystatement.html",
		ImplGuide: "USCore 3.1",
	}

	actualVal := validator.EndpointFunctionValid(cs)
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Rest, messaging, and document should exist, is instead %+v", actualVal))

	// removing one of the fields will still be valid

	cs2, err := deleteFieldFromCapStat(cs, "messaging")
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Actual = "rest,document"
	actualVal = validator2.EndpointFunctionValid(cs2)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Rest and document should exist, is instead %+v", actualVal))

	// removing all fields will be invalid

	cs3, err := deleteFieldFromCapStat(cs2, "rest")
	th.Assert(t, err == nil, err)
	cs4, err := deleteFieldFromCapStat(cs3, "document")
	th.Assert(t, err == nil, err)

	validator4, err := getValidator(cs4, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Actual = ""
	expectedVal.Valid = false
	actualVal = validator4.EndpointFunctionValid(cs4)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Rest, messaging, and document should not exist, is instead %+v", actualVal))
}

func Test_DescribeEndpointValid(t *testing.T) {
	cs, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, r4)
	th.Assert(t, err == nil, err)

	// base test

	expectedVal := endpointmanager.Rule{
		RuleName:  endpointmanager.DescribeEndptRule,
		Valid:     true,
		Actual:    "description,software,implementation",
		Expected:  "description,software,implementation",
		Comment:   "A Capability Statement SHALL have at least one of description, software, or implementation element.",
		Reference: "http://hl7.org/fhir/capabilitystatement.html",
		ImplGuide: "USCore 3.1",
	}

	actualVal := validator.DescribeEndpointValid(cs)
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Description, software, and implementation should exist, is instead %+v", actualVal))

	// removing one of the fields will still be valid

	cs2, err := deleteFieldFromCapStat(cs, "software")
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Actual = "description,implementation"
	actualVal = validator2.DescribeEndpointValid(cs2)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Description and implementation should exist, is instead %+v", actualVal))

	// removing all fields will be invalid

	cs3, err := deleteFieldFromCapStat(cs2, "description")
	th.Assert(t, err == nil, err)
	cs4, err := deleteFieldFromCapStat(cs3, "implementation")
	th.Assert(t, err == nil, err)

	validator4, err := getValidator(cs4, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Actual = ""
	expectedVal.Valid = false
	actualVal = validator4.DescribeEndpointValid(cs4)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("Description, software, and implementation should not exist, is instead %+v", actualVal))
}

func Test_DocumentSetValid(t *testing.T) {
	cs, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, r4)
	th.Assert(t, err == nil, err)

	// base test

	baseComment := "The set of documents must be unique by the combination of profile and mode."
	expectedVal := endpointmanager.Rule{
		RuleName:  endpointmanager.DocumentValidRule,
		Valid:     true,
		Expected:  "true",
		Actual:    "true",
		Comment:   baseComment,
		Reference: "http://hl7.org/fhir/capabilitystatement.html",
		ImplGuide: "USCore 3.1",
	}

	actualVal := validator.DocumentSetValid(cs)
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The set of documents should be unique, is instead %+v", actualVal))

	// make mode invalid

	cs2, err := nLevelNestedValueChange(cs, []string{"document", "mode"}, []int{0}, 1, badFormat, "")
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Valid = false
	expectedVal.Actual = "false"
	expectedVal.Comment = "Document field is not formatted correctly. Cannot check if the set of documents are unique. " + baseComment

	actualVal = validator2.DocumentSetValid(cs2)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("An invalid mode should make the check invalid, is instead %+v", actualVal))

	// make profile invalid

	cs3, err := nLevelNestedValueChange(cs, []string{"document", "profile"}, []int{0}, 1, badFormat, "")
	th.Assert(t, err == nil, err)

	validator3, err := getValidator(cs3, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Comment = "Document field is not formatted correctly. Cannot check if the set of documents are unique. " + baseComment
	actualVal = validator3.DocumentSetValid(cs3)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("An invalid profile should make the check invalid, is instead %+v", actualVal))

	// make both modes the same

	cs4, err := nLevelNestedValueChange(cs, []string{"document", "mode"}, []int{0}, 1, updateString, "producer")
	th.Assert(t, err == nil, err)

	validator4, err := getValidator(cs4, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Comment = "The set of documents are not unique. " + baseComment

	actualVal = validator4.DocumentSetValid(cs4)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The documents not being unique should make the check invalid, is instead %+v", actualVal))

	// bad format document field

	cs5, err := nLevelNestedValueChange(cs, []string{"document"}, []int{}, 0, badFormat, "")
	th.Assert(t, err == nil, err)

	validator5, err := getValidator(cs5, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Comment = "Document field is not formatted correctly. Cannot check if the set of documents are unique. " + baseComment
	actualVal = validator5.DocumentSetValid(cs5)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("An improperly formatted document field should make the check invalid, is instead %+v", actualVal))

	// remove document field

	cs6, err := deleteFieldFromCapStat(cs, "document")
	th.Assert(t, err == nil, err)

	validator6, err := getValidator(cs6, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Valid = true
	expectedVal.Actual = "true"
	expectedVal.Comment = "Document field does not exist, but is not required. " + baseComment
	actualVal = validator6.DocumentSetValid(cs6)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The check should be valid if the document field does not exist, is instead %+v", actualVal))
}

func Test_TLSVersion(t *testing.T) {
	cs, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, r4)
	th.Assert(t, err == nil, err)

	// base test

	expectedVal := endpointmanager.Rule{
		RuleName:  endpointmanager.TLSVersion,
		Valid:     true,
		Expected:  "TLS 1.2, TLS 1.3",
		Actual:    "TLS 1.2",
		Comment:   "Systems SHALL use TLS version 1.2 or higher for all transmissions not taking place over a secure network connection.",
		Reference: "https://www.hl7.org/fhir/us/core/security.html",
		ImplGuide: "USCore 3.1",
	}
	actualVal := validator.TLSVersion("TLS 1.2")
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("TLSVersion check should be valid, returned value is instead %+v", actualVal))

	// tls version is not valid

	expectedVal.Valid = false
	expectedVal.Actual = "TLS 1.1"
	actualVal = validator.TLSVersion("TLS 1.1")
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("TLSVersion check should be invalid, returned value is instead %+v", actualVal))
}

func Test_UniqueResources(t *testing.T) {
	cs, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, r4)
	th.Assert(t, err == nil, err)

	// base test

	baseComment := "A given resource can only be described once per RESTful mode."
	expectedVal := endpointmanager.Rule{
		RuleName:  endpointmanager.UniqueResourcesRule,
		Valid:     true,
		Actual:    "true",
		Expected:  "true",
		Comment:   baseComment,
		ImplGuide: "USCore 3.1",
		Reference: "http://hl7.org/fhir/capabilitystatement.html",
	}
	actualVal := validator.UniqueResources(cs)
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The given resources should be unique, is instead %+v", actualVal))

	// If there are two patient resources

	cs2, err := nLevelNestedValueChange(cs, []string{"rest", "resource", "type"}, []int{0, 1}, 2, updateString, "Patient")
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Valid = false
	expectedVal.Actual = "false"
	expectedVal.Comment = "The resource type Patient is not unique. " + baseComment
	actualVal = validator2.UniqueResources(cs2)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The given resources should not be unique, is instead %+v", actualVal))
}

func Test_SearchParamsUnique(t *testing.T) {
	cs, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, r4)
	th.Assert(t, err == nil, err)

	// base test

	baseComment := "Search parameter names must be unique in the context of a resource."
	expectedVal := endpointmanager.Rule{
		RuleName:  endpointmanager.SearchParamsRule,
		Valid:     true,
		Actual:    "true",
		Expected:  "true",
		Comment:   baseComment,
		ImplGuide: "USCore 3.1",
		Reference: "http://hl7.org/fhir/capabilitystatement.html",
	}
	actualVal := validator.SearchParamsUnique(cs)
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The search parameters in each resource should be unique, is instead %+v", actualVal))

	// If there are two of the same search parameter

	cs2, err := nLevelNestedValueChange(cs, []string{"rest", "resource", "searchParam", "name"}, []int{0, 0, 0}, 3, updateString, "general-practitioner")
	th.Assert(t, err == nil, err)

	validator2, err := getValidator(cs2, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Valid = false
	expectedVal.Actual = "false"
	expectedVal.Comment = "The resource type Patient does not have unique searchParams. " + baseComment
	actualVal = validator2.SearchParamsUnique(cs2)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("The search parameters in each resource should not be unique, is instead %+v", actualVal))

	// malformed searchParam name value

	cs3, err := nLevelNestedValueChange(cs, []string{"rest", "resource", "searchParam", "name"}, []int{0, 0, 0}, 3, badFormat, "")
	th.Assert(t, err == nil, err)

	validator3, err := getValidator(cs3, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Comment = "The resource type Patient is not formatted properly. " + baseComment
	actualVal = validator3.SearchParamsUnique(cs3)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("SearchParamsUnique check should be invalid because a searchParam name field is malformed, is instead %+v", actualVal))

	// name does not exist, should return same values as previous test

	cs4, err := nLevelNestedValueChange(cs, []string{"rest", "resource", "searchParam", "name"}, []int{0, 0, 0}, 3, deleteField, "")
	th.Assert(t, err == nil, err)

	validator4, err := getValidator(cs4, r4)
	th.Assert(t, err == nil, err)

	actualVal = validator4.SearchParamsUnique(cs4)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("SearchParamsUnique check should be invalid because a searchParam name field does not exist, is instead %+v", actualVal))

	// searchParam is malformed, should return same values as previous test

	cs5, err := nLevelNestedValueChange(cs, []string{"rest", "resource", "searchParam"}, []int{0, 0}, 2, badFormat, "")
	th.Assert(t, err == nil, err)

	validator5, err := getValidator(cs5, r4)
	th.Assert(t, err == nil, err)

	actualVal = validator5.SearchParamsUnique(cs5)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("SearchParamsUnique check should be invalid because a searchParam field is malformed, is instead %+v", actualVal))

	// searchParam does not exist, which does not throw an error because they aren't required

	cs6, err := nLevelNestedValueChange(cs, []string{"rest", "resource", "searchParam"}, []int{0, 0}, 2, deleteField, "")
	th.Assert(t, err == nil, err)

	validator6, err := getValidator(cs6, r4)
	th.Assert(t, err == nil, err)

	expectedVal.Valid = true
	expectedVal.Actual = "true"
	expectedVal.Comment = baseComment
	actualVal = validator6.SearchParamsUnique(cs6)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("SearchParamsUnique check should be valid even though searchParams do not exist, is instead %+v", actualVal))
}

func Test_VersionResponseValid(t *testing.T) {
	cs, err := getR4CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, r4)
	th.Assert(t, err == nil, err)

	fhirVersion := "4.0.1"
	defaultFhirVersion := "4.0.1"

	// base test

	expectedVal := endpointmanager.Rule{
		RuleName: endpointmanager.VersionsResponseRule,
		Valid:    true,
		Expected: "4.0.1",
		Actual:   "4.0.1",
		Comment:  "Expected $versions operation to be supported, and expected default fhir version to be returned from server when no version specified.",
	}

	actualVal := validator.VersionResponseValid(fhirVersion, defaultFhirVersion)
	eq := reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("$version operation and default version should be valid, is instead %+v", actualVal))

	// defaultVersion is not 4.0.1

	defaultFhirVersion = "1.0.2"
	expectedVal.Expected = "1.0.2"
	expectedVal.Valid = false
	actualVal = validator.VersionResponseValid(fhirVersion, defaultFhirVersion)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("$version operation should be valid, but default version should not match fhir version, is instead %+v", actualVal))

	// defaultVersion is 4.0

	defaultFhirVersion = "4.0"
	expectedVal.Expected = "4.0"
	expectedVal.Valid = true
	actualVal = validator.VersionResponseValid(fhirVersion, defaultFhirVersion)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("$version operation should be valid, and default version's publication and major components should match fhir version, is instead %+v", actualVal))

	// $version operation no response

	defaultFhirVersion = "None"
	expectedVal.Actual = "4.0.1"
	expectedVal.Expected = ""
	expectedVal.Valid = false
	expectedVal.Comment = "Expected $versions operation to be supported, but no response was received"
	actualVal = validator.VersionResponseValid(fhirVersion, defaultFhirVersion)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("$version operation should not be valid, is instead %+v", actualVal))
}

// getDSTU2CapStat gets a DSTU2 Capability Statement
func getDSTU2CapStat() (capabilityparser.CapabilityStatement, error) {
	path := filepath.Join("../../../testdata", "test_dstu2_capability_statement.json")
	csJSON, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

// getDSTU2CapStat gets a R4 Capability Statement
func getR4CapStat() (capabilityparser.CapabilityStatement, error) {
	path := filepath.Join("../../../testdata", "test_r4_capability_statement.json")
	csJSON, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

// getValidator gets the validator for the correct version
func getValidator(capStat capabilityparser.CapabilityStatement, checkVersions []string) (Validator, error) {
	fhirVersion, err := capStat.GetFHIRVersion()
	if err != nil {
		return nil, err
	}
	if !helpers.StringArrayContains(checkVersions, fhirVersion) {
		return nil, fmt.Errorf("capstat's returned version %s is not one of the expected versions %+v", fhirVersion, checkVersions)
	}
	validator := ValidatorForFHIRVersion(fhirVersion)
	return validator, nil
}

// deleteFieldFromCapStat delets the given field from the capability statement
func deleteFieldFromCapStat(cs capabilityparser.CapabilityStatement, field string) (capabilityparser.CapabilityStatement, error) {
	csInt, _, err := getCapFormats(cs)
	if err != nil {
		return nil, err
	}

	delete(csInt, field)

	csJSON, err := json.Marshal(csInt)
	if err != nil {
		return nil, err
	}

	return capabilityparser.NewCapabilityStatement(csJSON)
}

// nLevelNestedValueChange runs the given function on a field specified by the fields, indicies, and level
// e.g. If you want to change a name value in  { rest: [ { resource: [ name: "hello" ] } ] }, then
// fields would be ["rest", "resource", "name"]
// indices would be [0, 0] since you're getting the first element in both the rest & resource arrays
// level would be 2 since you're trying to access an array that's nested twice
func nLevelNestedValueChange(cs capabilityparser.CapabilityStatement,
	fields []string,
	indices []int,
	level int,
	functionToRun func(map[string]interface{}, string, string),
	functionVar string) (capabilityparser.CapabilityStatement, error) {

	csInt, _, err := getCapFormats(cs)
	if err != nil {
		return nil, err
	}
	if len(fields) > level+1 || len(indices) > level {
		return nil, errors.New("an unexpected number of parameters")
	}

	innerFieldMap := csInt
	loop := 0
	for loop < level {
		innerFieldMap, err = getInnerFieldMap(innerFieldMap, fields[loop], indices[loop])
		if err != nil {
			return nil, err
		}
		loop++
	}

	functionToRun(innerFieldMap, fields[loop], functionVar)

	csJSON, err := json.Marshal(csInt)
	if err != nil {
		return nil, err
	}

	return capabilityparser.NewCapabilityStatement(csJSON)
}

// badFormat sets the given field to an improperly formatted value
func badFormat(innerField map[string]interface{}, field string, optional string) {
	innerField[field] = []int{1, 2, 3}
}

// deleteField deletes a field from the given map
func deleteField(innerField map[string]interface{}, field string, optional string) {
	delete(innerField, field)
}

// updateString updates the given field to the given value
func updateString(innerField map[string]interface{}, field string, optional string) {
	innerField[field] = optional
}

// getInnerFieldMap gets the map inside of an array of maps at the given index
func getInnerFieldMap(csInt map[string]interface{}, field string, index int) (map[string]interface{}, error) {
	fieldVal := csInt[field]
	fieldArray, ok := fieldVal.([]interface{})
	if !ok {
		return nil, errors.New("unable to cast to an []interface{}")
	}

	innerFieldMap, ok := fieldArray[index].(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to cast to a map[string]interface{}")
	}
	return innerFieldMap, nil
}

// getCapFormats gets the JSON & the interface version of a capability statement
func getCapFormats(cs capabilityparser.CapabilityStatement) (map[string]interface{}, []byte, error) {
	var csInt map[string]interface{}

	csJSON, err := cs.GetJSON()
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal(csJSON, &csInt)
	if err != nil {
		return nil, nil, err
	}

	return csInt, csJSON, nil
}
