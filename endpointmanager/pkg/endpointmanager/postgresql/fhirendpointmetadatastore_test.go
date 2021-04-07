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

func Test_PersistFHIREndpointMetadata(t *testing.T) {
	SetupStore()
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

	var endpointMetadata1 = &endpointmanager.FHIREndpointMetadata{
		URL:               "example.com/FHIR/DSTU2/",
		HTTPResponse:      200,
		Errors:            "Example Error",
		SMARTHTTPResponse: 0,
		Availability:      1.0}

	var endpointMetadata2 = &endpointmanager.FHIREndpointMetadata{
		URL:               "other.example.com/FHIR/DSTU2/",
		HTTPResponse:      404,
		Errors:            "Example Error 2",
		SMARTHTTPResponse: 0,
		Availability:      0}

	// endpointInfos
	var endpointInfo1 = &endpointmanager.FHIREndpointInfo{
		URL:                   "example.com/FHIR/DSTU2/",
		TLSVersion:            "TLS 1.1",
		MIMETypes:             []string{"application/json+fhir"},
		CapabilityStatement:   cs,
		RequestedFhirVersion:  "",
		CapabilityFhirVersion: "1.0.2",
		SMARTResponse:         nil,
		Metadata:              endpointMetadata1}
	var endpointInfo2 = &endpointmanager.FHIREndpointInfo{
		URL:                   "other.example.com/FHIR/DSTU2/",
		TLSVersion:            "TLS 1.2",
		RequestedFhirVersion:  "",
		CapabilityFhirVersion: "",
		MIMETypes:             []string{"application/fhir+json"},
		Metadata:              endpointMetadata2}

	// add endpointMetadata

	metadataID1, err := store.AddFHIREndpointMetadata(ctx, endpointMetadata1)
	if err != nil {
		t.Errorf("Error adding fhir endpointInfo: %s", err.Error())
	}

	metadataID2, err := store.AddFHIREndpointMetadata(ctx, endpointMetadata2)
	if err != nil {
		t.Errorf("Error adding fhir endpointInfo: %+v", err)
	}

	// retrieve endpointMetadata

	m1, err := store.GetFHIREndpointMetadata(ctx, metadataID1)
	if err != nil {
		t.Errorf("Error getting fhir endpoint Metadata: %s", err.Error())
	}
	if !m1.Equal(endpointMetadata1) {
		t.Errorf("retrieved endpointMetadata is not equal to saved endpointMetadata.")
	}

	m2, err := store.GetFHIREndpointMetadata(ctx, metadataID2)
	if err != nil {
		t.Errorf("Error getting fhir endpoint Metadata: %s", err.Error())
	}
	if !m2.Equal(endpointMetadata2) {
		t.Errorf("retrieved endpointMetadata is not equal to saved endpointMetadata.")
	}

	// add endpointMetadata and endpointInfo
	metadataID, err := store.AddFHIREndpointMetadata(ctx, endpointInfo1.Metadata)
	if err != nil {
		t.Errorf("Error adding fhir endpointInfo: %s", err.Error())
	}
	err = store.AddFHIREndpointInfo(ctx, endpointInfo1, metadataID)
	if err != nil {
		t.Errorf("Error adding fhir endpointInfo: %s", err.Error())
	}

	metadataID, err = store.AddFHIREndpointMetadata(ctx, endpointInfo2.Metadata)
	if err != nil {
		t.Errorf("Error adding fhir endpointInfo: %s", err.Error())
	}
	err = store.AddFHIREndpointInfo(ctx, endpointInfo2, metadataID)
	if err != nil {
		t.Errorf("Error adding fhir endpointInfo: %+v", err)
	}

	// retrieve endpointInfos

	e1, err := store.GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx, endpointInfo1.URL, endpointInfo1.RequestedFhirVersion)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if !e1.Equal(endpointInfo1) {
		t.Errorf("retrieved endpointInfo is not equal to saved endpointInfo.")
	}

	e2, err := store.GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx, endpointInfo2.URL, endpointInfo2.RequestedFhirVersion)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if !e2.Equal(endpointInfo2) {
		t.Errorf("retrieved endpointInfo is not equal to saved endpointInfo.")
	}

	// retrieve endpointMetadata

	m1, err = store.GetFHIREndpointMetadata(ctx, e1.Metadata.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpoint Metadata: %s", err.Error())
	}
	if !m1.Equal(endpointMetadata1) {
		t.Errorf("retrieved endpointMetadata is not equal to saved endpointMetadata.")
	}

	m2, err = store.GetFHIREndpointMetadata(ctx, e2.Metadata.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpoint Metadata: %s", err.Error())
	}
	if !m2.Equal(endpointMetadata2) {
		t.Errorf("retrieved endpointMetadata is not equal to saved endpointMetadata.")
	}

	// update endpoint info metadata id

	endpointInfo1.Metadata.HTTPResponse = 700

	metadataID1, err = store.AddFHIREndpointMetadata(ctx, endpointInfo1.Metadata)
	if err != nil {
		t.Errorf("Error adding fhir endpointMetadata: %s", err.Error())
	}

	err = store.UpdateMetadataIDInfo(ctx, metadataID1, endpointInfo1.ID)
	if err != nil {
		t.Errorf("Error updating fhir endpointInfo metadata ID: %s", err.Error())
	}

	e1, err = store.GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx, endpointInfo1.URL, endpointInfo1.RequestedFhirVersion)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}
	if e1.Metadata.HTTPResponse != 700 {
		t.Errorf("retrieved endpointInfo does not have updated HTTP Response.")
	}

	// check history table
	var count int

	// check insertions
	rows := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info_history WHERE id=$1 AND operation='I';", endpointInfo1.ID)
	err = rows.Scan(&count)
	if err != nil {
		t.Errorf("history count for insertions: %s", err.Error())
	}
	if count != 1 {
		t.Errorf("expected 1 insertion for endpointInfo1. Got %d.", count)
	}

	// check no updates after updating info metadata ID
	rows = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info_history WHERE id=$1 AND operation='U';", endpointInfo1.ID)
	err = rows.Scan(&count)
	if err != nil {
		t.Errorf("history count for updates: %s", err.Error())
	}
	if count != 0 {
		t.Errorf("expected 0 updates for endpointInfo1 in history table. Got %d.", count)
	}

	// check metadata table
	rows = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_metadata WHERE url=$1;", endpointInfo1.URL)
	err = rows.Scan(&count)
	if err != nil {
		t.Errorf("metadata count for insertions: %s", err.Error())
	}
	if count != 3 {
		t.Errorf("expected 3 insertions in metadata table for endpointInfo1 URL. Got %d.", count)
	}

	rows = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_metadata WHERE url=$1 AND http_response = 700;", endpointInfo1.URL)
	err = rows.Scan(&count)
	if err != nil {
		t.Errorf("metadata count for insertions: %s", err.Error())
	}
	if count != 1 {
		t.Errorf("expected 1 insertion in metadata table for endpointInfo1 URL with HTTP response 700. Got %d.", count)
	}

	rows = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_metadata WHERE url=$1 AND http_response = 200;", endpointInfo1.URL)
	err = rows.Scan(&count)
	if err != nil {
		t.Errorf("metadata count for insertions: %s", err.Error())
	}
	if count != 2 {
		t.Errorf("expected 2 insertions in metadata table for endpointInfo1 URL with HTTP response 200. Got %d.", count)
	}

	// update endpoint info metadata id

	endpointInfo1.Metadata.HTTPResponse = 404

	metadataID, err = store.AddFHIREndpointMetadata(ctx, endpointInfo1.Metadata)
	if err != nil {
		t.Errorf("Error adding update to fhir endpointMetadata: %s", err.Error())
	}

	err = store.UpdateFHIREndpointInfo(ctx, endpointInfo1, metadataID)
	if err != nil {
		t.Errorf("Error updating fhir endpointInfo: %s", err.Error())
	}

	e1, err = store.GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx, endpointInfo1.URL, endpointInfo1.RequestedFhirVersion)
	if err != nil {
		t.Errorf("Error getting fhir endpointInfo: %s", err.Error())
	}

	if e1.Metadata.HTTPResponse != 404 {
		t.Errorf("retrieved endpointInfo does not have updated HTTP Response.")
	}

	// check there is an update in history table
	rows = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info_history WHERE id=$1 AND operation='U';", endpointInfo1.ID)
	err = rows.Scan(&count)
	if err != nil {
		t.Errorf("history count for updates: %s", err.Error())
	}
	if count != 1 {
		t.Errorf("expected 1 update for endpointInfo1 in history table. Got %d.", count)
	}

}
