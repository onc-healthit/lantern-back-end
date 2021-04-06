package versionsoperatorparser

import (
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)


func Test_Equal(t *testing.T) {
	var vr1 VersionsResponse
	var vr2 VersionsResponse

	var equal bool

	vr1 = VersionsResponse{Response: nil}
	vr2 = VersionsResponse{Response: nil}

	// test both nil equal
	equal = vr1.Equal(vr2)
	th.Assert(t, equal, "expected equality nil to nil to be true")

	// test one nil, other not
	var resp map[string]interface{}
	resp = make(map[string]interface{})
	resp["versions"] = "[\"4.0\"]"
	resp["default"] = "4.0"
	vr2 = VersionsResponse{Response: resp}
	equal = vr1.Equal(vr2)
	th.Assert(t, !equal, "expected equality nil to not nil to be false")

	// test two same non-nil
	var resp1 map[string]interface{}
	resp1 = make(map[string]interface{})
	resp1["versions"] = "[\"4.0\"]"
	resp1["default"] = "4.0"
	vr1 = VersionsResponse{Response: resp1}
	equal = vr1.Equal(vr2)
	th.Assert(t, equal, "expected equality for same Response to be true")

	// test two different non-nil
	resp1["versions"] = "[\"1.0\"]"
	resp1["default"] = "1.0"
	vr1 = VersionsResponse{Response: resp1}
	equal = vr1.Equal(vr2)
	th.Assert(t, !equal, "expected equality for different Response to be false")
}

func Test_GetDefaultVersion(t *testing.T) {
	var vr1 VersionsResponse

	var equal bool

	var resp map[string]interface{}
	resp = make(map[string]interface{})
	resp["versions"] = "[\"1.0\"]"
	resp["default"] = "4.0"
	vr1 = VersionsResponse{Response: resp}

	// test populated versions Response
	equal = vr1.GetDefaultVersion() == "4.0"
	th.Assert(t, equal, "expected default version to be 4.0")

	vr1 = VersionsResponse{Response: nil}
	var emptySlice []string
	// test not populated VerrsionsResponse
	equal = helpers.StringArraysEqual(vr1.GetSupportedVersions(), emptySlice)
	th.Assert(t, equal, "expected non-empty versions to not be empty")

}

func Test_GetSupportedVersions(t *testing.T) {
	var vr1 VersionsResponse

	var equal bool

	var resp map[string]interface{}
	resp = make(map[string]interface{})
	resp["versions"] = []string{"4.0"}
	resp["default"] = "4.0"
	vr1 = VersionsResponse{Response: resp}

	var populatedSlice []string
	populatedSlice = append(populatedSlice, "4.0")
	// test not populated VerrsionsResponse
	equal = helpers.StringArraysEqual(vr1.GetSupportedVersions(), populatedSlice)
	th.Assert(t, equal, "expected versions Response to include version")

	vr1 = VersionsResponse{Response: nil}
	var emptySlice []string
	// test not populated VerrsionsResponse
	equal = helpers.StringArraysEqual(vr1.GetSupportedVersions(), emptySlice)
	th.Assert(t, equal, "expected empty versions to be empty")

}