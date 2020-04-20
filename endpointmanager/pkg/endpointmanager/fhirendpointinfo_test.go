package endpointmanager

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
)

func Test_FHIREndpointInfoEqual(t *testing.T) {

	// capability statement
	path := filepath.Join("../testdata", "cerner_capability_dstu2.json")
	csJSON, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
	}
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	if err != nil {
		t.Error(err)
	}

	// endpointInfos
	var endpointInfo1 = &FHIREndpointInfo{
		ID:                  1,
		FHIREndpointID:      2,
		HealthITProductID:   3,
		TLSVersion:          "TLS 1.1",
		MIMETypes:           []string{"application/json+fhir", "application/fhir+json"},
		HTTPResponse:        200,
		Errors:              "Example Error",
		Vendor:              "Cerner Corporation",
		CapabilityStatement: cs}
	var endpointInfo2 = &FHIREndpointInfo{
		ID:                  1,
		FHIREndpointID:      2,
		HealthITProductID:   3,
		TLSVersion:          "TLS 1.1",
		MIMETypes:           []string{"application/json+fhir", "application/fhir+json"},
		HTTPResponse:        200,
		Errors:              "Example Error",
		Vendor:              "Cerner Corporation",
		CapabilityStatement: cs}

	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expected endpointInfo1 to equal endpointInfo2. They are not equal.")
	}

	endpointInfo2.ID = 2
	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expect endpointInfo 1 to equal endpointInfo 2. ids should be ignored. %d vs %d", endpointInfo1.ID, endpointInfo2.ID)
	}
	endpointInfo2.ID = endpointInfo1.ID

	endpointInfo2.FHIREndpointID = 3
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expect endpointInfo 1 to not equal endpointInfo 2. FHIREndpointID should be different. %d vs %d", endpointInfo1.FHIREndpointID, endpointInfo2.FHIREndpointID)
	}
	endpointInfo2.FHIREndpointID = endpointInfo1.FHIREndpointID

	endpointInfo2.HealthITProductID = 4
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expect endpointInfo 1 to not equal endpointInfo 2. HealthITProductID should be different. %d vs %d", endpointInfo1.HealthITProductID, endpointInfo2.HealthITProductID)
	}
	endpointInfo2.HealthITProductID = endpointInfo1.HealthITProductID

	endpointInfo2.Vendor = "other"
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expect endpointInfo 1 to not equal endpointInfo 2. Vendor should be different. %s vs %s", endpointInfo1.Vendor, endpointInfo2.Vendor)
	}
	endpointInfo2.Vendor = endpointInfo1.Vendor

	endpointInfo2.TLSVersion = "other"
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. TLSVersion should be different. %s vs %s", endpointInfo1.TLSVersion, endpointInfo2.TLSVersion)
	}
	endpointInfo2.TLSVersion = endpointInfo1.TLSVersion

	endpointInfo2.MIMETypes = []string{"other"}
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. MIMETypes should be different. %s vs %s", endpointInfo1.MIMETypes, endpointInfo2.MIMETypes)
	}
	endpointInfo2.MIMETypes = []string{"application/fhir+json", "other"}
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. MIMETypes should be different. %s vs %s", endpointInfo1.MIMETypes, endpointInfo2.MIMETypes)
	}
	endpointInfo2.MIMETypes = []string{"application/fhir+json", "application/json+fhir"}
	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expected endpointInfo1 to equal endpointInfo 2. MIMETypes are same but in different order. %s vs %s", endpointInfo1.MIMETypes, endpointInfo2.MIMETypes)
	}
	endpointInfo2.MIMETypes = endpointInfo1.MIMETypes

	endpointInfo2.HTTPResponse = 404
	if endpointInfo2.Equal(endpointInfo1) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. HTTPResponse should be different. %d vs %d", endpointInfo1.HTTPResponse, endpointInfo2.HTTPResponse)
	}
	endpointInfo2.HTTPResponse = endpointInfo1.HTTPResponse

	endpointInfo2.Errors = "other"
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. Errors should be different. %s vs %s", endpointInfo1.Errors, endpointInfo2.Errors)
	}
	endpointInfo2.Errors = endpointInfo1.Errors

	endpointInfo2.CapabilityStatement = nil
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. CapabilityStatement should be different. %s vs %s", endpointInfo1.CapabilityStatement, endpointInfo2.CapabilityStatement)
	}
	endpointInfo2.CapabilityStatement = endpointInfo1.CapabilityStatement

	endpointInfo1.CapabilityStatement = nil
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. CapabilityStatement should be different. %s vs %s", endpointInfo1.CapabilityStatement, endpointInfo2.CapabilityStatement)
	}
	endpointInfo1.CapabilityStatement = endpointInfo2.CapabilityStatement

	endpointInfo2 = nil
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal nil endpointInfo 2.")
	}
	endpointInfo2 = endpointInfo1

	endpointInfo1 = nil
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect nil endpointInfo1 to equal endpointInfo 2.")
	}

	endpointInfo2 = nil
	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Nil endpointInfo 1 should equal nil endpointInfo 2.")
	}
}
