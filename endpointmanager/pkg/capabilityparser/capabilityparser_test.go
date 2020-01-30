package capabilityparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

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

	_, err = NewCapabilityStatement(csJSON)
	th.Assert(t, err != nil, "expected error due to unknown FHIR version")

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

func Test_GetSoftwareName(t *testing.T) {
	field := "software"
	//field2 := "name"

	// basic

	expected := "Allscripts FHIR"
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetSoftwareName()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))

	// bad format

	cs1, err := getBadFormatCapStat(cs, field)
	th.Assert(t, err == nil, err)

	_, err = cs1.GetSoftwareName()
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing field

	expected = ""

	cs2, err := deleteFieldFromCapStat(cs, field)
	th.Assert(t, err == nil, err)

	actual, err = cs2.GetSoftwareName()
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
	//field2 := "name"

	// basic

	expected := "19.4.121.0"
	cs, err := getDSTU2CapStat()
	th.Assert(t, err == nil, err)

	actual, err := cs.GetSoftwareVersion()
	th.Assert(t, err == nil, err)
	th.Assert(t, actual == expected, fmt.Sprintf("expected %s. received %s.", expected, actual))

	// bad format

	cs1, err := getBadFormatCapStat(cs, field)
	th.Assert(t, err == nil, err)

	_, err = cs1.GetSoftwareVersion()
	th.Assert(t, err != nil, "expected error due to bad format")

	// missing field

	expected = ""

	cs2, err := deleteFieldFromCapStat(cs, field)
	th.Assert(t, err == nil, err)

	actual, err = cs2.GetSoftwareVersion()
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
	var csInt map[string]interface{}

	csJSON, err := cs.GetJSON()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(csJSON, &csInt)
	if err != nil {
		return nil, err
	}

	csInt[field] = []int{1, 2, 3} // bad format for publisher
	csJSON, err = json.Marshal(csInt)
	if err != nil {
		return nil, err
	}

	return NewCapabilityStatement(csJSON)
}

func getNestedBadFormatCapStat(cs CapabilityStatement, field1 string, field2 string) (CapabilityStatement, error) {
	var csInt map[string]interface{}

	csJSON, err := cs.GetJSON()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(csJSON, &csInt)
	if err != nil {
		return nil, err
	}

	innerField := csInt[field1]
	innerFieldMap, ok := innerField.(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to cast to a map[string]interface{}")
	}

	innerFieldMap[field2] = []int{1, 2, 3} // bad format for publisher
	csJSON, err = json.Marshal(csInt)
	if err != nil {
		return nil, err
	}

	return NewCapabilityStatement(csJSON)
}

func deleteFieldFromCapStat(cs CapabilityStatement, field string) (CapabilityStatement, error) {
	var csInt map[string]interface{}

	csJSON, err := cs.GetJSON()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(csJSON, &csInt)
	if err != nil {
		return nil, err
	}

	delete(csInt, field)

	csJSON, err = json.Marshal(csInt)
	if err != nil {
		return nil, err
	}

	return NewCapabilityStatement(csJSON)
}

func deleteNestedFieldFromCapStat(cs CapabilityStatement, field1 string, field2 string) (CapabilityStatement, error) {
	var csInt map[string]interface{}

	csJSON, err := cs.GetJSON()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(csJSON, &csInt)
	if err != nil {
		return nil, err
	}

	innerField := csInt[field1]
	innerFieldMap, ok := innerField.(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to cast to a map[string]interface{}")
	}

	delete(innerFieldMap, field2)

	csJSON, err = json.Marshal(csInt)
	if err != nil {
		return nil, err
	}

	return NewCapabilityStatement(csJSON)
}
