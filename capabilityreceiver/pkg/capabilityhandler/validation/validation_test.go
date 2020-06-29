package validation

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_RunValidation(t *testing.T) {
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	validator, err := getValidator(cs, dstu2)
	th.Assert(t, err == nil, err)

	// base test
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

	actualVal := validator.RunValidation(cs, 200, []string{fhir2LessJSONMIMEType}, "1.0.2", "TLS 1.2", 200)
	th.Assert(t, len(actualVal.Results) == 5, fmt.Sprintf("RunValidation should have returned 5 validation checks, instead it returned %d", len(actualVal.Results)))
	eq := reflect.DeepEqual(actualVal.Results[0], expectedFirstVal)
	th.Assert(t, eq == true, "RunValidation's first returned validation is not correct")
	eq = reflect.DeepEqual(actualVal.Results[4], expectedLastVal)
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
	th.Assert(t, eq == true, fmt.Sprintf("CapStatExists check should be true, returned value is instead %+v", actualCap))

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
	th.Assert(t, eq == true, fmt.Sprintf("CapStatExists check should be false, returned value is instead %+v", actualCap))

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
	th.Assert(t, eq == true, fmt.Sprintf("CapStatExists check should be true, returned value is instead %+v", actualCap))
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
	th.Assert(t, eq == true, fmt.Sprintf("MimeTypeValid check should be valid, is instead %+v", actualVal))

	// fhirVersion 3+ test

	expectedVal.Expected = fhir3PlusJSONMIMEType
	expectedVal.Actual = fhir3PlusJSONMIMEType
	expectedVal.Comment = "FHIR Version 3.0.0 requires the Mime Type to be " + fhir3PlusJSONMIMEType

	actualVal = validator.MimeTypeValid([]string{fhir3PlusJSONMIMEType}, "3.0.0")
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("MimeTypeValid check should be valid, is instead %+v", actualVal))

	// no mime types

	expectedVal.Valid = false
	expectedVal.Expected = "N/A"
	expectedVal.Actual = ""
	expectedVal.Comment = "No mime type given; cannot validate mime type."

	actualVal = validator.MimeTypeValid([]string{}, "1.0.2")
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("MimeTypeValid check should be invalid, is instead %+v", actualVal))

	// no version

	expectedVal.Actual = fhir2LessJSONMIMEType
	expectedVal.Comment = "Unknown FHIR Version; cannot validate mime type."
	actualVal = validator.MimeTypeValid([]string{fhir2LessJSONMIMEType}, "")
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("MimeTypeValid check should be invalid, is instead %+v", actualVal))

	// mixmatch mime type and version

	expectedVal.Expected = fhir3PlusJSONMIMEType
	expectedVal.Comment = "FHIR Version 3.0.0 requires the Mime Type to be " + fhir3PlusJSONMIMEType
	actualVal = validator.MimeTypeValid([]string{fhir2LessJSONMIMEType}, "3.0.0")
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("MimeTypeValid check should be invalid, is instead %+v", actualVal))

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
	th.Assert(t, eq == true, fmt.Sprintf("MimeTypeValid check should be valid, returned value is instead %+v", actualVal))
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
	th.Assert(t, eq == true, fmt.Sprintf("HTTPResponseValid check should be valid, is instead %+v", actualVal))

	// httpResponse is 0

	expectedVal.Valid = false
	expectedVal.Actual = "0"
	expectedVal.Comment = "The GET request failed with no returned HTTP response status code."

	actualVal = validator.HTTPResponseValid(0)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("HTTPResponseValid check should be invalid, is instead %+v", actualVal))

	// httpResponse is 404

	expectedVal.Actual = "404"
	expectedVal.Comment = "The HTTP response code was 404 instead of 200. "

	actualVal = validator.HTTPResponseValid(404)
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("HTTPResponseValid check should be invalid, is instead %+v", actualVal))

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
	th.Assert(t, eq == true, fmt.Sprintf("MimeTypeValid check should be valid, returned value is instead %+v", actualVal))
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
	th.Assert(t, eq == true, fmt.Sprintf("FhirVersion check should be valid, is instead %+v", actualVal))

	// fhirVersion is not 4.0.1

	expectedVal.Actual = "1.0.2"
	expectedVal.Valid = false
	actualVal = validator.FhirVersion("1.0.2")
	eq = reflect.DeepEqual(actualVal, expectedVal)
	th.Assert(t, eq == true, fmt.Sprintf("FhirVersion check should be invalid, is instead %+v", actualVal))
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
	th.Assert(t, eq == true, fmt.Sprintf("KindValid check should be valid, is instead %+v", actualVal))

	// cap stat is nil

	expectedVal.Valid = false
	expectedVal.Actual = ""
	expectedVal.Comment = "Capability Statement does not exist; cannot check kind value. " + baseComment
	expectedArray = []endpointmanager.Rule{
		expectedVal,
	}

	actualVal = validator.KindValid(nil)
	eq = reflect.DeepEqual(actualVal, expectedArray)
	th.Assert(t, eq == true, fmt.Sprintf("KindValid check should be invalid, is instead %+v", actualVal))

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
	th.Assert(t, eq == true, fmt.Sprintf("KindValid check should be invalid, is instead %+v", actualVal))

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
	th.Assert(t, eq == true, fmt.Sprintf("KindValid first check should be valid, is instead %+v", actualVal[0]))
	eq = reflect.DeepEqual(actualVal[1], expectedInstanceVal)
	th.Assert(t, eq == true, fmt.Sprintf("KindValid second check should be valid, is instead %+v", actualVal[1]))

	// if implementation doesn't exist, then the second check is invalid

	cs4, err := deleteFieldFromCapStat(cs3, "implementation")
	th.Assert(t, err == nil, err)

	validator4, err := getValidator(cs4, r4)
	th.Assert(t, err == nil, err)

	expectedInstanceVal.Valid = false
	expectedInstanceVal.Actual = "false"
	actualVal = validator4.KindValid(cs4)
	eq = reflect.DeepEqual(actualVal[1], expectedInstanceVal)
	th.Assert(t, eq == true, fmt.Sprintf("KindValid second check should be valid, is instead %+v", actualVal[1]))
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

func getValidator(capStat capabilityparser.CapabilityStatement, checkVersions []string) (Validator, error) {
	fhirVersion, err := capStat.GetFHIRVersion()
	if err != nil {
		return nil, err
	}
	if !contains(checkVersions, fhirVersion) {
		return nil, fmt.Errorf("capstat's returned version %s is not one of the expected versions %+v", fhirVersion, checkVersions)
	}
	validator := GetValidationForVersion(fhirVersion)
	return validator, nil
}

func deleteFieldFromCapStat(cs capabilityparser.CapabilityStatement, field string) (capabilityparser.CapabilityStatement, error) {
	csInt, csJSON, err := getCapFormats(cs)
	if err != nil {
		return nil, err
	}

	delete(csInt, field)

	csJSON, err = json.Marshal(csInt)
	if err != nil {
		return nil, err
	}

	return capabilityparser.NewCapabilityStatement(csJSON)
}

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
