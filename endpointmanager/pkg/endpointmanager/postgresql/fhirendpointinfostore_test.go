// +build integration

package postgresql

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_PersistFHIREndpointInfo(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

	// capability statement
	path := filepath.Join("../../testdata", "cerner_capability_dstu2.json")
	csJSON, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
	}
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	if err != nil {
		t.Error(err)
	}

	// add an endpoint that can later be referenced
	var endpoint1 = &endpointmanager.FHIREndpoint{
		URL:              "example.com/FHIR/DSTU2/",
		OrganizationName: "Example Inc.",
		ListSource:       "https://github.com/cerner/ignite-endpoints"}
	var endpoint2 = &endpointmanager.FHIREndpoint{
		URL:              "other.example.com/FHIR/DSTU2/",
		OrganizationName: "Other Example Inc."}
	store.AddFHIREndpoint(ctx, endpoint1)
	store.AddFHIREndpoint(ctx, endpoint2)

	// endpointInfos
	var endpointInfo1 = &endpointmanager.FHIREndpointInfo{
		URL:                 endpoint1.URL,
		TLSVersion:          "TLS 1.1",
		MIMETypes:           []string{"application/json+fhir"},
		HTTPResponse:        200,
		Errors:              "Example Error",
		Vendor:              "Cerner",
		CapabilityStatement: cs}
	var endpointInfo2 = &endpointmanager.FHIREndpointInfo{
		URL:          endpoint2.URL,
		TLSVersion:   "TLS 1.2",
		MIMETypes:    []string{"application/fhir+json"},
		HTTPResponse: 404,
		Errors:       "Example Error 2"}

	// add endpointInfos

	err = store.AddFHIREndpointInfo(ctx, endpointInfo1)
	if err != nil {
		t.Errorf("Error adding fhir endpointInfo: %s", err.Error())
	}

	err = store.AddFHIREndpointInfo(ctx, endpointInfo2)
	if err != nil {
		t.Errorf("Error adding fhir endpointInfo: %+v", err)
	}

	// retrieve endpointInfos

	e1, err := store.GetFHIREndpointInfoUsingURL(ctx, endpoint1.URL)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if !e1.Equal(endpointInfo1) {
		t.Errorf("retrieved endpointInfo is not equal to saved endpointInfo.")
	}

	e2, err := store.GetFHIREndpointInfoUsingURL(ctx, endpoint2.URL)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if !e2.Equal(endpointInfo2) {
		t.Errorf("retrieved endpointInfo is not equal to saved endpointInfo.")
	}

	eID1, err := store.GetFHIREndpointInfo(ctx, e1.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if !eID1.Equal(endpointInfo1) {
		t.Errorf("retrieved endpointInfo is not equal to saved endpointInfo.")
	}

	eID2, err := store.GetFHIREndpointInfo(ctx, e2.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if !eID2.Equal(endpointInfo2) {
		t.Errorf("retrieved endpointInfo is not equal to saved endpointInfo.")
	}

	// update endpointInfo

	e1.HTTPResponse = 700

	err = store.UpdateFHIREndpointInfo(ctx, e1)
	if err != nil {
		t.Errorf("Error updating fhir endpointInfo: %s", err.Error())
	}

	e1, err = store.GetFHIREndpointInfo(ctx, endpointInfo1.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if e1.Equal(endpointInfo1) {
		t.Errorf("retrieved UPDATED endpointInfo is equal to original endpointInfo.")
	}
	if e1.UpdatedAt.Equal(e1.CreatedAt) {
		t.Errorf("UpdatedAt is not being properly set on update.")
	}

	// update with nil capability statement
	capStat := e1.CapabilityStatement
	e1.CapabilityStatement = nil

	err = store.UpdateFHIREndpointInfo(ctx, e1)
	if err != nil {
		t.Errorf("Error updating fhir endpointInfo: %s", err.Error())
	}

	e1, err = store.GetFHIREndpointInfo(ctx, endpointInfo1.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if e1.CapabilityStatement != nil {
		t.Errorf("Expected capability statement to be nil")
	}

	e1.CapabilityStatement = capStat

	// delete endpointInfos

	err = store.DeleteFHIREndpointInfo(ctx, endpointInfo1)
	if err != nil {
		t.Errorf("Error deleting fhir endpointInfo: %s", err.Error())
	}

	_, err = store.GetFHIREndpointInfo(ctx, endpointInfo1.ID) // ensure we deleted the entry
	if err == nil {
		t.Errorf("did not expected endpoint in db")
	}

	_, err = store.GetFHIREndpointInfo(ctx, endpointInfo2.ID) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("error retrieving endpointInfo2 after deleting endpointInfo1: %s", err.Error())
	}

	err = store.DeleteFHIREndpointInfo(ctx, endpointInfo2)
	if err != nil {
		t.Errorf("Error deleting fhir endpointInfo: %s", err.Error())
	}

	_, err = store.GetFHIREndpointInfo(ctx, endpointInfo2.ID) // ensure we deleted the entry
	if err == nil {
		t.Errorf("did not expected endpoint in db")
	}
}
