package smartparser

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_SMARTResponseEqual(t *testing.T) {

	// get SMART Response
	path := filepath.Join("../testdata", "authorization_cerner_smart_response.json")
	smartResponseJSON1, err := ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	SMARTResponse1, err := NewSMARTResp(smartResponseJSON1)
	th.Assert(t, err == nil, err)

	// test nil
	SMARTResponse2, err := NewSMARTResp(nil)
	th.Assert(t, err == nil, err)

	equal := SMARTResponse1.Equal(SMARTResponse2)
	th.Assert(t, !equal, "expected equality comparison to nil to be false")

	// test equal
	smartResponseJSON2, err := ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	SMARTResponse2, err = NewSMARTResp(smartResponseJSON2)
	th.Assert(t, err == nil, err)

	if !SMARTResponse1.Equal(SMARTResponse2) {
		t.Errorf("Expect SMARTResponse1 to equal SMARTResponse2, but they were not equal.")
	}

	// test not equal
	SMARTResponseOriginal2 := SMARTResponse2
	SMARTResponse2, err = deleteFieldFromSmartResponse(SMARTResponse2, "capabilities")
	th.Assert(t, err == nil, err)

	equal = SMARTResponse1.Equal(SMARTResponse2)
	th.Assert(t, !equal, "expected equality comparison of unequal SMART responses to be false")

	SMARTResponse2 = SMARTResponseOriginal2
	SMARTResponse2Int, _, err := getRespFormats(SMARTResponse2)
	th.Assert(t, err == nil, err)

	SMARTResponse2Int["capabilities"] = []string{
		"launch-ehr",
		"launch-standalone",
		"client-public",
		"client-confidential-symmetric",
		"sso-openid-connect",
		"context-banner",
		"context-style",
		"context-ehr-patient",
		"context-ehr-encounter",
		"permission-patient",
		"permission-user",
		"permission-v2",
		"authorize-post",
		"fakeCapability",
	}

	SMARTResponse2 = NewSMARTRespFromInterface(SMARTResponse2Int)

	equal = SMARTResponse1.Equal(SMARTResponse2)
	th.Assert(t, !equal, "expected equality comparison of unequal SMART responses to be false")

	// get different SMART Response format to ensure Equal works for any smart response format
	path = filepath.Join("../testdata", "fhir_sandbox_smart_response.json")
	smartResponseJSON1, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	SMARTResponse1, err = NewSMARTResp(smartResponseJSON1)
	th.Assert(t, err == nil, err)

	// test nil
	SMARTResponse2, err = NewSMARTResp(nil)
	th.Assert(t, err == nil, err)

	equal = SMARTResponse1.Equal(SMARTResponse2)
	th.Assert(t, !equal, "expected equality comparison to nil to be false")

	// test equal
	smartResponseJSON2, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	SMARTResponse2, err = NewSMARTResp(smartResponseJSON2)
	th.Assert(t, err == nil, err)

	if !SMARTResponse1.Equal(SMARTResponse2) {
		t.Errorf("Expect SMARTResponse1 to equal SMARTResponse2, but they were not equal.")
	}

	// test not equal
	SMARTResponse2, err = deleteFieldFromSmartResponse(SMARTResponse2, "token_endpoint_auth_signing_alg_values_supported")
	th.Assert(t, err == nil, err)

	equal = SMARTResponse1.Equal(SMARTResponse2)
	th.Assert(t, !equal, "expected equality comparison of unequal SMART responses to be false")

}

// The EqualIgnore function for now should get the same results as the Equal function
func Test_SMARTResponseEqualIgnore(t *testing.T) {

	ignoredFields := []string{}

	// get SMART Response
	path := filepath.Join("../testdata", "authorization_cerner_smart_response.json")
	smartResponseJSON1, err := ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	SMARTResponse1, err := NewSMARTResp(smartResponseJSON1)
	th.Assert(t, err == nil, err)

	// test nil
	SMARTResponse2, err := NewSMARTResp(nil)
	th.Assert(t, err == nil, err)

	equal := SMARTResponse1.EqualIgnore(SMARTResponse2, ignoredFields)
	th.Assert(t, !equal, "expected equality comparison to nil to be false")

	// test equal
	smartResponseJSON2, err := ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	SMARTResponse2, err = NewSMARTResp(smartResponseJSON2)
	th.Assert(t, err == nil, err)

	equal = SMARTResponse1.EqualIgnore(SMARTResponse2, ignoredFields)
	th.Assert(t, equal, "Expect SMARTResponse1 to equal SMARTResponse2, but they were not equal.")

	// test still equal when field added to ignored field list
	ignoredFields = append(ignoredFields, "capabilities")

	equal = SMARTResponse1.EqualIgnore(SMARTResponse2, ignoredFields)
	th.Assert(t, equal, "Expect SMARTResponse1 to equal SMARTResponse2, but they were not equal.")

	ignoredFields = []string{}

	// test not equal when deleting capabilities field
	SMARTResponseOriginal2 := SMARTResponse2
	SMARTResponse2, err = deleteFieldFromSmartResponse(SMARTResponse2, "capabilities")
	th.Assert(t, err == nil, err)

	equal = SMARTResponse1.EqualIgnore(SMARTResponse2, ignoredFields)
	th.Assert(t, !equal, "expected equality comparison of unequal SMART responses to be false")

	// test equal when deleted capabilities field added to ignored field list
	ignoredFields = append(ignoredFields, "capabilities")
	equal = SMARTResponse1.EqualIgnore(SMARTResponse2, ignoredFields)
	th.Assert(t, equal, "expected equality comparison of SMART responses to be true since they only differ by ignored field")

	SMARTResponse2 = SMARTResponseOriginal2
	ignoredFields = []string{}

	// test not equal when deleting authorization_endpoint field
	SMARTResponse2, err = deleteFieldFromSmartResponse(SMARTResponse2, "authorization_endpoint")
	th.Assert(t, err == nil, err)

	equal = SMARTResponse1.EqualIgnore(SMARTResponse2, ignoredFields)
	th.Assert(t, !equal, "expected equality comparison of unequal SMART responses to be false")

	// test equal when deleted authorization_endpoint field added to ignored field list
	ignoredFields = append(ignoredFields, "authorization_endpoint")
	equal = SMARTResponse1.EqualIgnore(SMARTResponse2, ignoredFields)
	th.Assert(t, equal, "expected equality comparison of SMART responses to be true since they only differ by ignored field")

	SMARTResponse2 = SMARTResponseOriginal2
	ignoredFields = []string{}

	// test altering fields
	SMARTResponse2Int, _, err := getRespFormats(SMARTResponse2)
	th.Assert(t, err == nil, err)

	SMARTResponse2Int["capabilities"] = []string{
		"launch-ehr",
		"launch-standalone",
		"client-public",
		"client-confidential-symmetric",
		"sso-openid-connect",
		"context-banner",
		"context-style",
		"context-ehr-patient",
		"context-ehr-encounter",
		"permission-patient",
		"permission-user",
		"permission-v2",
		"authorize-post",
		"fakeCapability",
	}

	SMARTResponse2 = NewSMARTRespFromInterface(SMARTResponse2Int)

	// test not equal when capabilities field not in ignoredFields list
	equal = SMARTResponse1.EqualIgnore(SMARTResponse2, ignoredFields)
	th.Assert(t, !equal, "expected equality comparison of unequal SMART responses to be false")

	// test equal when capabilities field is added to ignoredFields list
	ignoredFields = append(ignoredFields, "capabilities")
	equal = SMARTResponse1.EqualIgnore(SMARTResponse2, ignoredFields)
	th.Assert(t, equal, "expected equality comparison of SMART responses to be true since they only differ by ignored field")

	// test more than one ignored field
	SMARTResponse2Int["authorization_endpoint"] = "fake_authorization_endpoint"
	SMARTResponse2 = NewSMARTRespFromInterface(SMARTResponse2Int)

	equal = SMARTResponse1.EqualIgnore(SMARTResponse2, ignoredFields)
	th.Assert(t, !equal, "expected equality comparison of unequal SMART responses to be false")

	ignoredFields = append(ignoredFields, "authorization_endpoint")
	equal = SMARTResponse1.EqualIgnore(SMARTResponse2, ignoredFields)
	th.Assert(t, equal, "expected equality comparison of SMART responses to be true since they only differ by ignored fields")

	SMARTResponse2 = SMARTResponseOriginal2
	ignoredFields = []string{}
}
