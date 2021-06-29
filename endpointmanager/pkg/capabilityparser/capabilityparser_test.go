package capabilityparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/sirupsen/logrus/hooks/test"
)

// added messaging to the test CapabilityStatement since the field did not exist in any of our examples
// Using the example from FHIR's website: https://www.hl7.org/fhir/DSTU2/conformance-example.json.html
var messagingObj = []map[string]interface{}{
	{
		"endpoint": []interface{}{
			endpointObj,
		},
		"reliableCache": float64(30),
		"documentation": "ADT A08 equivalent for external system notifications",
		"event": []interface{}{
			map[string]interface{}{
				"code": map[string]interface{}{
					"system": "http://hl7.org/fhir/message-type",
					"code":   "admin-notify",
				},
				"category": "Consequence",
				"mode":     "receiver",
				"focus":    "Patient",
				"request": map[string]interface{}{
					"reference": "StructureDefinition/daf-patient",
				},
				"response": map[string]interface{}{
					"reference": "StructureDefinition/MessageHeader",
				},
				"documentation": "Notification of an update to a patient resource. changing the links is not supported",
			},
		},
	},
}

var endpointObj = map[string]interface{}{
	"protocol": map[string]interface{}{
		"system": "http://hl7.org/fhir/message-transport",
		"code":   "mllp",
	},
	"address": "mllp:10.1.1.10:9234",
}

var resourceObj = map[string]interface{}{
	"type": "AllergyIntolerance",
	"profile": map[string]interface{}{
		"reference": "StructureDefinition",
		"display":   "Definition of capabilities for the resource",
	},
	"interaction": []interface{}{
		map[string]interface{}{
			"code":          "read",
			"documentation": "",
		},
		map[string]interface{}{
			"code": "search-type",
		},
	},
	"versioning":        "no-version",
	"readHistory":       false,
	"updateCreate":      false,
	"conditionalCreate": false,
	"conditionalUpdate": false,
	"conditionalDelete": "not-supported",
	"searchParam": []interface{}{
		map[string]interface{}{
			"name": "patient",
			"type": "reference",
			"target": []interface{}{
				"Patient",
			},
		},
		map[string]interface{}{
			"name":          "date",
			"type":          "date",
			"documentation": "",
		},
	},
}

func Test_NewCapabilityStatement(t *testing.T) {
	var ok bool
	var cs CapabilityStatement
	var err error
	var path string
	var csJSON []byte
	var csInt map[string]interface{}

	// basic test dstu2

	// capability statement
	path = filepath.Join("../testdata", "epic_capability_dstu2.json")
	csJSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err = NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)
	_, ok = cs.(*dstu2CapabilityParser)
	th.Assert(t, ok, "expected to be able to convert to dstu2CapabilityParser type")
	_, ok = cs.(*stu3CapabilityParser)
	th.Assert(t, !ok, "not expected to be able to conver to stu3CapabilityParser type")

	// basic test stu3

	// capability statement
	path = filepath.Join("../testdata", "epic_capability_stu3.json")
	csJSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err = NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)
	_, ok = cs.(*stu3CapabilityParser)
	th.Assert(t, ok, "expected to be able to convert to stu3CapabilityParser type")
	_, ok = cs.(*dstu2CapabilityParser)
	th.Assert(t, !ok, "not expected to be able to conver to dstu2CapabilityParser type")

	// basic test r4
	err = json.Unmarshal(csJSON, &csInt)
	th.Assert(t, err == nil, err)
	csInt["fhirVersion"] = "4.0.1"
	csJSON, err = json.Marshal(csInt)
	th.Assert(t, err == nil, err)

	cs, err = NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)
	_, ok = cs.(*r4CapabilityParser)
	th.Assert(t, ok, "expected to be able to convert to r4CapabilityParser type")
	_, ok = cs.(*dstu2CapabilityParser)
	th.Assert(t, !ok, "not expected to be able to conver to dstu2CapabilityParser type")

	// test unknown
	err = json.Unmarshal(csJSON, &csInt)
	th.Assert(t, err == nil, err)
	csInt["fhirVersion"] = "5.3.2"
	csJSON, err = json.Marshal(csInt)
	th.Assert(t, err == nil, err)

	hook := test.NewGlobal()

	cs, err = NewCapabilityStatement(csJSON)

	th.Assert(t, len(hook.Entries) == 1, fmt.Sprintf("expected hook entries to be 1, was %d", len(hook.Entries)))
	th.Assert(t, hook.LastEntry().Message == "unknown FHIR version, 5.3.2, defaulting to DSTU2", "did not get expected log warning message for unknown FHIR version")

	th.Assert(t, err == nil, "expected no error due to unknown FHIR version defaulting to DSTU2")
	_, ok = cs.(*dstu2CapabilityParser)
	th.Assert(t, ok, "expected to be able to convert to dstu2CapabilityParser type")

	// test empty byte string
	cs, err = NewCapabilityStatement([]byte{})
	th.Assert(t, err == nil, err)
	th.Assert(t, cs == nil, "expected nil capability statement returned")

	// test null json object
	cs, err = NewCapabilityStatement([]byte("null"))
	th.Assert(t, err == nil, err)
	th.Assert(t, cs == nil, "expected nil capability statement returned")
}

func Test_GetPublisher(t *testing.T) {
	field := "publisher"

	// basic

	expected := "Allscripts"
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetPublisher()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))

	// bad format

	cs1, err := getBadFormatCapStat(cs, field)
	th.Assert(t, err == nil, err)

	_, err = cs1.GetPublisher()
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing field

	expected = ""

	cs2, err := deleteFieldFromCapStat(cs, field)
	th.Assert(t, err == nil, err)

	actual, err = cs2.GetPublisher()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))
}

func Test_GetFHIRVersion(t *testing.T) {

	// basic

	expected := "1.0.2"
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetFHIRVersion()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))

	// can't test other forms because it messes up the constructor, which relies on this field (which is a required field...)
}

func Test_GetCopyright(t *testing.T) {
	field := "copyright"

	// basic

	expected := "Copyright 2015 Allscripts Healthcare Solutions, Inc.. All rights reserved"
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetCopyright()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))

	// bad format

	cs1, err := getBadFormatCapStat(cs, field)
	th.Assert(t, err == nil, err)

	_, err = cs1.GetCopyright()
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing field

	expected = ""

	cs2, err := deleteFieldFromCapStat(cs, field)
	th.Assert(t, err == nil, err)

	actual, err = cs2.GetCopyright()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))
}

func Test_GetSoftware(t *testing.T) {
	field := "software"
	var emptyMap map[string]interface{}

	// basic

	expected := map[string]interface{}{
		"name":        "Allscripts FHIR",
		"version":     "19.4.121.0",
		"releaseDate": "2019-11-22",
	}
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetSoftware()
	th.Assert(t, err == nil, err)
	eq := reflect.DeepEqual(actual, expected)
	th.Assert(t, eq == true, fmt.Sprintf("expected %s. received %s.", expected, actual))

	// bad format

	cs1, err := getBadFormatCapStat(cs, field)
	th.Assert(t, err == nil, err)

	_, err = cs1.GetSoftware()
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing field

	cs2, err := deleteFieldFromCapStat(cs, field)
	th.Assert(t, err == nil, err)

	actual, err = cs2.GetSoftware()
	th.Assert(t, err == nil, err)
	eq = reflect.DeepEqual(actual, emptyMap)
	th.Assert(t, eq == true, fmt.Sprintf("expected %s. received %s.", emptyMap, actual))
}

func Test_GetSoftwareName(t *testing.T) {
	field := "software"

	// basic

	expected := "Allscripts FHIR"
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetSoftwareName()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))

	// nested bad format

	cs3, err := getNestedBadFormatCapStat(cs, field, "name")
	th.Assert(t, err == nil, err)

	_, err = cs3.GetSoftwareName()
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing nested field

	expected = ""

	cs4, err := deleteNestedFieldFromCapStat(cs, field, "name")
	th.Assert(t, err == nil, err)

	actual, err = cs4.GetSoftwareName()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))

}

func Test_GetSoftwareVersion(t *testing.T) {
	field := "software"

	// basic

	expected := "19.4.121.0"
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetSoftwareVersion()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))

	// nested bad format

	cs3, err := getNestedBadFormatCapStat(cs, field, "version")
	th.Assert(t, err == nil, err)

	_, err = cs3.GetSoftwareVersion()
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing nested field

	expected = ""

	cs4, err := deleteNestedFieldFromCapStat(cs, field, "version")
	th.Assert(t, err == nil, err)

	actual, err = cs4.GetSoftwareVersion()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))
}

func Test_GetRest(t *testing.T) {
	field := "rest"

	// basic
	// Unnecessary to check every field so just checking two
	expectedMode := "server"
	expectedDocumentation := "Information about the system's restful capabilities that apply across all applications, such as security"
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetRest()
	th.Assert(t, err == nil, err)
	th.Assert(t, len(actual) == 1, fmt.Sprintf("length of rest array should be 1. instead it is %d", len(actual)))
	th.Assert(t, expectedMode == actual[0]["mode"], fmt.Sprintf("expected mode %s. received mode %s.", expectedMode, actual[0]["mode"]))
	th.Assert(t, expectedDocumentation == actual[0]["documentation"], fmt.Sprintf("expected mode %s. received mode %s.", expectedMode, actual[0]["documentation"]))

	// bad format

	cs1, err := getBadFormatCapStat(cs, field)
	th.Assert(t, err == nil, err)

	_, err = cs1.GetRest()
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing field

	cs2, err := deleteFieldFromCapStat(cs, field)
	th.Assert(t, err == nil, err)

	actual, err = cs2.GetRest()
	th.Assert(t, err == nil, err)
	th.Assert(t, len(actual) == 0, fmt.Sprintf("length of rest array should be 1. instead it is %d", len(actual)))
}

func Test_GetResourceList(t *testing.T) {
	field := "rest"
	var emptyMap []map[string]interface{}

	// basic

	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	// Get rest object to use for the test
	expectedMode := "server"
	expectedDocumentation := "Information about the system's restful capabilities that apply across all applications, such as security"
	rest, err := cs.GetRest()
	th.Assert(t, err == nil, err)
	th.Assert(t, len(rest) == 1, fmt.Sprintf("length of rest array should be 1. instead it is %d", len(rest)))
	th.Assert(t, expectedMode == rest[0]["mode"], fmt.Sprintf("expected mode %s. received mode %s.", expectedMode, rest[0]["mode"]))
	th.Assert(t, expectedDocumentation == rest[0]["documentation"], fmt.Sprintf("expected mode %s. received mode %s.", expectedMode, rest[0]["documentation"]))

	expectedResource := resourceObj
	actualRecs, err := cs.GetResourceList(rest[0])
	th.Assert(t, err == nil, err)
	check := false
	for _, resource := range actualRecs {
		eq := reflect.DeepEqual(resource, expectedResource)
		if eq {
			check = true
		}
	}
	th.Assert(t, check == true, "expected resource was not in given resource list")

	// bad format

	cs1, err := getArrayNestedBadFormatCapStat(cs, field, "resource", 0)
	th.Assert(t, err == nil, err)

	// Have to get rest field again to get the updated value
	rest, err = cs1.GetRest()
	th.Assert(t, err == nil, err)
	th.Assert(t, len(rest) == 1, fmt.Sprintf("length of rest array should be 1. instead it is %d", len(rest)))
	th.Assert(t, expectedMode == rest[0]["mode"], fmt.Sprintf("expected mode %s. received mode %s.", expectedMode, rest[0]["mode"]))
	th.Assert(t, expectedDocumentation == rest[0]["documentation"], fmt.Sprintf("expected mode %s. received mode %s.", expectedMode, rest[0]["documentation"]))

	_, err = cs1.GetResourceList(rest[0])
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing field

	cs2, err := deleteArrayNestedFieldCapStat(cs, field, "resource", 0)
	th.Assert(t, err == nil, err)

	// Have to get rest field again to get the updated value
	rest, err = cs2.GetRest()
	th.Assert(t, err == nil, err)
	th.Assert(t, len(rest) == 1, fmt.Sprintf("length of rest array should be 1. instead it is %d", len(rest)))
	th.Assert(t, expectedMode == rest[0]["mode"], fmt.Sprintf("expected mode %s. received mode %s.", expectedMode, rest[0]["mode"]))
	th.Assert(t, expectedDocumentation == rest[0]["documentation"], fmt.Sprintf("expected mode %s. received mode %s.", expectedMode, rest[0]["documentation"]))

	actualRecs, err = cs2.GetResourceList(rest[0])
	th.Assert(t, err == nil, err)
	eq := reflect.DeepEqual(actualRecs, emptyMap)
	th.Assert(t, eq == true, fmt.Sprintf("expected %s. received %s.", emptyMap, actualRecs))
}

func Test_GetKind(t *testing.T) {
	field := "kind"

	// basic

	expected := "instance"
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetKind()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))

	// bad format

	cs1, err := getBadFormatCapStat(cs, field)
	th.Assert(t, err == nil, err)

	_, err = cs1.GetKind()
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing field

	expected = ""

	cs2, err := deleteFieldFromCapStat(cs, field)
	th.Assert(t, err == nil, err)

	actual, err = cs2.GetKind()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))
}

func Test_GetImplementation(t *testing.T) {
	field := "implementation"
	var emptyMap map[string]interface{}

	// basic

	expected := map[string]interface{}{
		"description": "Local Client Implementation",
		"url":         "https://fhir.fhirpoint.open.allscripts.com/fhirroute/fhir/10028551",
	}
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetImplementation()
	th.Assert(t, err == nil, err)
	eq := reflect.DeepEqual(actual, expected)
	th.Assert(t, eq == true, fmt.Sprintf("expected %s. received %s.", expected, actual))

	// bad format

	cs1, err := getBadFormatCapStat(cs, field)
	th.Assert(t, err == nil, err)

	_, err = cs1.GetImplementation()
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing field

	cs2, err := deleteFieldFromCapStat(cs, field)
	th.Assert(t, err == nil, err)

	actual, err = cs2.GetImplementation()
	th.Assert(t, err == nil, err)
	eq = reflect.DeepEqual(actual, emptyMap)
	th.Assert(t, eq == true, fmt.Sprintf("expected %s. received %s.", emptyMap, actual))
}

func Test_GetMessaging(t *testing.T) {
	field := "messaging"
	var emptyMap []map[string]interface{}

	// basic

	expected := messagingObj
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetMessaging()
	th.Assert(t, err == nil, err)
	eq := reflect.DeepEqual(actual, expected)
	th.Assert(t, eq == true, fmt.Sprintf("expected %s. received %s.", expected, actual))

	// bad format

	cs1, err := getBadFormatCapStat(cs, field)
	th.Assert(t, err == nil, err)

	_, err = cs1.GetMessaging()
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing field

	cs2, err := deleteFieldFromCapStat(cs, field)
	th.Assert(t, err == nil, err)

	actual, err = cs2.GetMessaging()
	th.Assert(t, err == nil, err)
	eq = reflect.DeepEqual(actual, emptyMap)
	th.Assert(t, eq == true, fmt.Sprintf("expected %s. received %s.", emptyMap, actual))
}

func Test_GetMessagingEndpoint(t *testing.T) {
	field := "messaging"
	var emptyMap []map[string]interface{}

	// basic

	// Get messaging object to use for the test
	expected := messagingObj
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	messaging, err := cs.GetMessaging()
	th.Assert(t, err == nil, err)
	eq := reflect.DeepEqual(messaging, expected)
	th.Assert(t, eq == true, fmt.Sprintf("expected %s. received %s.", expected, messaging))

	expectedEndpts := []map[string]interface{}{
		endpointObj,
	}
	actualEndpts, err := cs.GetMessagingEndpoint(messaging[0])
	th.Assert(t, err == nil, err)
	eq = reflect.DeepEqual(actualEndpts, expectedEndpts)
	th.Assert(t, eq == true, fmt.Sprintf("expected %s. received %s.", expectedEndpts, actualEndpts))

	// bad format

	cs1, err := getArrayNestedBadFormatCapStat(cs, field, "endpoint", 0)
	th.Assert(t, err == nil, err)

	// Have to get messaging again to get the updated value
	messaging, err = cs1.GetMessaging()
	th.Assert(t, err == nil, err)

	_, err = cs1.GetMessagingEndpoint(messaging[0])
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing field

	cs2, err := deleteArrayNestedFieldCapStat(cs, field, "endpoint", 0)
	th.Assert(t, err == nil, err)

	// Have to get messaging again to get the updated value
	messaging, err = cs2.GetMessaging()
	th.Assert(t, err == nil, err)

	actualEndpts, err = cs2.GetMessagingEndpoint(messaging[0])
	th.Assert(t, err == nil, err)
	eq = reflect.DeepEqual(actualEndpts, emptyMap)
	th.Assert(t, eq == true, fmt.Sprintf("expected %s. received %s.", emptyMap, actualEndpts))
}

func Test_GetDocument(t *testing.T) {
	field := "document"
	var emptyMap []map[string]interface{}

	// basic

	// added document to the test CapabilityStatement since the field did not exist in any of our examples
	// Using the example from FHIR's website: https://www.hl7.org/fhir/DSTU2/conformance-example.json.html
	expected := []map[string]interface{}{
		{
			"mode":          "consumer",
			"documentation": "Basic rules for all documents in the EHR system",
			"profile": map[string]interface{}{
				"reference": "http://fhir.hl7.org/base/Profilebc054d23-75e1-4dc6-aca5-838b6b1ac81d/_history/b5fdd9fc-b021-4ea1-911a-721a60663796",
			},
		},
	}
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetDocument()
	th.Assert(t, err == nil, err)
	eq := reflect.DeepEqual(actual, expected)
	th.Assert(t, eq == true, fmt.Sprintf("expected %s. received %s.", expected, actual))

	// bad format

	cs1, err := getBadFormatCapStat(cs, field)
	th.Assert(t, err == nil, err)

	_, err = cs1.GetDocument()
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing field

	cs2, err := deleteFieldFromCapStat(cs, field)
	th.Assert(t, err == nil, err)

	actual, err = cs2.GetDocument()
	th.Assert(t, err == nil, err)
	eq = reflect.DeepEqual(actual, emptyMap)
	th.Assert(t, eq == true, fmt.Sprintf("expected %s. received %s.", emptyMap, actual))
}

func Test_GetDescription(t *testing.T) {
	field := "description"

	// basic

	expected := "Conformance statement for Allscripts FHIR service."
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetDescription()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))

	// bad format

	cs1, err := getBadFormatCapStat(cs, field)
	th.Assert(t, err == nil, err)

	_, err = cs1.GetDescription()
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing field

	expected = ""

	cs2, err := deleteFieldFromCapStat(cs, field)
	th.Assert(t, err == nil, err)

	actual, err = cs2.GetDescription()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))
}

func Test_Equal(t *testing.T) {
	var cs1 CapabilityStatement
	var cs2 CapabilityStatement
	var equal bool
	var err error

	cs1, err = getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	// test nil
	cs2, err = NewCapabilityStatement(nil)
	th.Assert(t, err == nil, err)

	equal = cs1.Equal(cs2)
	th.Assert(t, !equal, "expected equality comparison to nil to be false")

	// test equal
	cs2, err = getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	equal = cs1.Equal(cs2)
	th.Assert(t, equal, "expected equality comparison of equal capability statement to be true")

	// test not equal
	cs2, err = deleteFieldFromCapStat(cs2, "publisher")
	th.Assert(t, err == nil, err)

	equal = cs1.Equal(cs2)
	th.Assert(t, !equal, "expected equality comparison of unequal capability statement to be false")

}

func Test_Equal_Ignore(t *testing.T) {
	var cs1 CapabilityStatement
	var cs2 CapabilityStatement
	var equal bool
	var err error

	cs1, err = getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	// test nil
	cs2, err = NewCapabilityStatement(nil)
	th.Assert(t, err == nil, err)

	equal = cs1.EqualIgnore(cs2)
	th.Assert(t, !equal, "expected equality comparison to nil to be false")

	// test equal
	cs2, err = getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	equal = cs1.EqualIgnore(cs2)
	th.Assert(t, equal, "expected equality comparison of equal capability statement to be true")

	// test equal when cs2 has different date
	cs2, err = getDSTU2CapStat()
	th.Assert(t, err == nil, err)
	cs2, err = getBadFormatCapStat(cs2, "date")
	th.Assert(t, err == nil, err)

	equal = cs1.EqualIgnore(cs2)
	th.Assert(t, equal, "expected equality comparison of equal capability statement to be true")

	// test equal when cs1 has different date
	cs2, err = getDSTU2CapStat()
	th.Assert(t, err == nil, err)
	cs1, err = getBadFormatCapStat(cs1, "date")
	th.Assert(t, err == nil, err)

	equal = cs1.EqualIgnore(cs2)
	th.Assert(t, equal, "expected equality comparison of equal capability statement to be true")

	// test not equal
	cs2, err = deleteFieldFromCapStat(cs2, "publisher")
	th.Assert(t, err == nil, err)

	equal = cs1.Equal(cs2)
	th.Assert(t, !equal, "expected equality comparison of unequal capability statement to be false")

}

func getDSTU2CapStat() (CapabilityStatement, error) {
	path := filepath.Join("../testdata", "allscripts_capability_dstu2.json")
	csJSON, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cs, err := NewCapabilityStatement(csJSON)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

func getBadFormatCapStat(cs CapabilityStatement, field string) (CapabilityStatement, error) {
	csInt, _, err := getCapFormats(cs)
	if err != nil {
		return nil, err
	}

	csInt[field] = []int{1, 2, 3} // bad format for given field
	csJSON, err := json.Marshal(csInt)
	if err != nil {
		return nil, err
	}

	return NewCapabilityStatement(csJSON)
}

func getNestedBadFormatCapStat(cs CapabilityStatement, field1 string, field2 string) (CapabilityStatement, error) {
	csInt, _, err := getCapFormats(cs)
	if err != nil {
		return nil, err
	}

	innerField := csInt[field1]
	innerFieldMap, ok := innerField.(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to cast to a map[string]interface{}")
	}

	innerFieldMap[field2] = []int{1, 2, 3} // bad format for given field
	csJSON, err := json.Marshal(csInt)
	if err != nil {
		return nil, err
	}

	return NewCapabilityStatement(csJSON)
}

func getArrayNestedBadFormatCapStat(cs CapabilityStatement, field1 string, field2 string, index int) (CapabilityStatement, error) {
	csInt, _, err := getCapFormats(cs)
	if err != nil {
		return nil, err
	}

	field1Val := csInt[field1]
	fieldArray, ok := field1Val.([]interface{})
	if !ok {
		return nil, errors.New("unable to cast to an []interface{}")
	}

	innerFieldMap, ok := fieldArray[index].(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to cast to a map[string]interface{}")
	}

	innerFieldMap[field2] = []int{1, 2, 3} // bad format for given field
	csJSON, err := json.Marshal(csInt)
	if err != nil {
		return nil, err
	}

	return NewCapabilityStatement(csJSON)
}

func deleteNestedFieldFromCapStat(cs CapabilityStatement, field1 string, field2 string) (CapabilityStatement, error) {
	csInt, _, err := getCapFormats(cs)
	if err != nil {
		return nil, err
	}

	innerField := csInt[field1]
	innerFieldMap, ok := innerField.(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to cast to a map[string]interface{}")
	}

	delete(innerFieldMap, field2)

	csJSON, err := json.Marshal(csInt)
	if err != nil {
		return nil, err
	}

	return NewCapabilityStatement(csJSON)
}

func deleteArrayNestedFieldCapStat(cs CapabilityStatement, field1 string, field2 string, index int) (CapabilityStatement, error) {
	csInt, _, err := getCapFormats(cs)
	if err != nil {
		return nil, err
	}

	field1Val := csInt[field1]
	fieldArray, ok := field1Val.([]interface{})
	if !ok {
		return nil, errors.New("unable to cast to an []interface{}")
	}

	innerFieldMap, ok := fieldArray[index].(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to cast to a map[string]interface{}")
	}

	delete(innerFieldMap, field2)

	csJSON, err := json.Marshal(csInt)
	if err != nil {
		return nil, err
	}

	return NewCapabilityStatement(csJSON)
}
