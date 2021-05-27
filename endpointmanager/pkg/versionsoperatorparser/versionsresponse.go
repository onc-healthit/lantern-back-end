package versionsoperatorparser

import (
	"bytes"
	"encoding/json"
)

// VersionsResponse is a wrapper struct for a response to the $versions FHIR operation
// this implementation assumes that we are requesting and receiving the application/json response
// future support for the FHIR representation of this will require making the VersionsResponse an
// interface following the same pattern used for the CapabilityStatement
type VersionsResponse struct {
	Response map[string]interface{}
}

// Equal checks if the versions response is equal to the given versionsresponse statement.
func (vr1 *VersionsResponse) Equal(vr2 VersionsResponse) bool {
	j1, err := vr1.GetJSON()
	if err != nil {
		return false
	}
	j2, err := vr2.GetJSON()
	if err != nil {
		return false
	}
	if !bytes.Equal(j1, j2) {
		return false
	}

	return true
}

// GetJSON returns the JSON representation of the versions response
func (vr *VersionsResponse) GetJSON() ([]byte, error) {
	return json.Marshal(vr)
}

// GetDefaultVersion gets the default FHIR version out of the versions response
func (vr *VersionsResponse) GetDefaultVersion() string {
	if vr.Response == nil {
		return ""
	}
	return vr.Response["default"].(string)
}

// GetSupportedVersions gets the supported FHIR versions out of the versions response
func (vr *VersionsResponse) GetSupportedVersions() []string {
	if vr.Response == nil || vr.Response["versions"] == nil {
		var empty []string
		return empty
	}
	var versions []string
	for _, ver := range vr.Response["versions"].([]interface{}) {
		versions = append(versions, ver.(string))
	}
	return versions
}
