package endpointmanager

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_SMARTResponseEqual(t *testing.T) {

	// get two differing formats for SMART Response
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

}
