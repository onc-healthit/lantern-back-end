package versionsoperatorparser

import (
	"encoding/json"
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
	resp := "{\"versions\": [\"4.0\",\"1.0\"],\"default\": \"4.0\"}"
	var jsonResponse interface{}
	json.Unmarshal([]byte(resp), &(jsonResponse))
	vr2 = VersionsResponse{Response: jsonResponse.(map[string]interface{})}
	equal = vr1.Equal(vr2)
	th.Assert(t, !equal, "expected equality nil to not nil to be false")

	// test two same non-nil
	resp = "{\"versions\": [\"4.0\",\"1.0\"],\"default\": \"4.0\"}"
	json.Unmarshal([]byte(resp), &(jsonResponse))
	vr1 = VersionsResponse{Response: jsonResponse.(map[string]interface{})}
	equal = vr1.Equal(vr2)
	th.Assert(t, equal, "expected equality for same Response to be true")

	// test two different non-nil
	resp = "{\"versions\": [\"4.0\"],\"default\": \"4.0\"}"
	json.Unmarshal([]byte(resp), &(jsonResponse))
	vr1 = VersionsResponse{Response: jsonResponse.(map[string]interface{})}
	equal = vr1.Equal(vr2)
	th.Assert(t, !equal, "expected equality for different Response to be false")
}

func Test_GetDefaultVersion(t *testing.T) {
	var vr1 VersionsResponse

	var equal bool

	resp := "{\"versions\": [\"4.0\",\"1.0\"],\"default\": \"4.0\"}"
	var jsonResponse interface{}
	json.Unmarshal([]byte(resp), &(jsonResponse))
	vr1 = VersionsResponse{Response: jsonResponse.(map[string]interface{})}

	// test populated versions Response
	equal = vr1.GetDefaultVersion() == "4.0"
	th.Assert(t, equal, "expected default version to be 4.0")

	vr1 = VersionsResponse{Response: nil}

	equal = vr1.GetDefaultVersion() == ""
	th.Assert(t, equal, "expected default version to be empty string")

}

func Test_GetSupportedVersions(t *testing.T) {
	var vr1 VersionsResponse

	var equal bool

	resp := "{\"versions\": [\"4.0\"],\"default\": \"4.0\"}"
	var jsonResponse interface{}
	json.Unmarshal([]byte(resp), &(jsonResponse))
	vr1 = VersionsResponse{Response: jsonResponse.(map[string]interface{})}

	var populatedSlice []string
	populatedSlice = append(populatedSlice, "4.0")
	// test not populated VersionsResponse
	equal = helpers.StringArraysEqual(vr1.GetSupportedVersions(), populatedSlice)
	th.Assert(t, equal, "expected versions Response to include version")

	vr1 = VersionsResponse{Response: nil}
	var emptySlice []string
	// test not populated VersionsResponse
	equal = helpers.StringArraysEqual(vr1.GetSupportedVersions(), emptySlice)
	th.Assert(t, equal, "expected empty versions to be empty")

}