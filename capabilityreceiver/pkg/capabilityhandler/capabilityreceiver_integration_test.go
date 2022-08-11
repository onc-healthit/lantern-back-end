// +build integration

package capabilityhandler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var store *postgresql.Store

var testFhirEndpoint1 = &endpointmanager.FHIREndpoint{
	URL: "http://example.com/DTSU2/",
}
var testFhirEndpoint2 = &endpointmanager.FHIREndpoint{
	URL: "https://test-two.com",
}

var vendors []*endpointmanager.Vendor = []*endpointmanager.Vendor{
	{
		Name:          "Epic Systems Corporation",
		DeveloperCode: "A",
		CHPLID:        1,
	},
	{
		Name:          "Cerner Corporation",
		DeveloperCode: "B",
		CHPLID:        2,
	},
	{
		Name:          "Cerner Health Services, Inc.",
		DeveloperCode: "C",
		CHPLID:        3,
	},
}

func TestMain(m *testing.M) {
	var err error

	err = config.SetupConfigForTests()
	if err != nil {
		panic(err)
	}

	err = setup()
	if err != nil {
		panic(err)
	}

	hap := th.HostAndPort{Host: viper.GetString("dbhost"), Port: viper.GetString("dbport")}
	err = th.CheckResources(hap)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	teardown()
	os.Exit(code)
}

func Test_removeNoLongerExistingVersionsInfos(t *testing.T) {
	var ctx = context.Background()
	var endpointURL = "https://example.com"
	// add vendor that can later be referenced
	fakeVendor := &endpointmanager.Vendor{
		Name:          "Fake Vendor",
		DeveloperCode: "12345",
		CHPLID:        222,
	}
	var endpointMetadataDefaultRequest = &endpointmanager.FHIREndpointMetadata{
		URL:                  endpointURL,
		HTTPResponse:         200,
		Errors:               "",
		SMARTHTTPResponse:    0,
		Availability:         1.0,
		RequestedFhirVersion: "None"}

	var endpointMetadataRequestVersion1 = &endpointmanager.FHIREndpointMetadata{
		URL:                  endpointURL,
		HTTPResponse:         200,
		Errors:               "",
		SMARTHTTPResponse:    0,
		Availability:         1.0,
		RequestedFhirVersion: "None"}

	// endpointInfos
	var endpointInfoDefaultRequest = &endpointmanager.FHIREndpointInfo{
		URL:                   "https://example.com",
		VendorID:              fakeVendor.ID,
		TLSVersion:            "TLS 1.1",
		MIMETypes:             []string{"application/json+fhir"},
		CapabilityStatement:   nil,
		SMARTResponse:         nil,
		RequestedFhirVersion:  "None",
		CapabilityFhirVersion: "1.0.2",
		Metadata:              endpointMetadataDefaultRequest}

	var endpointInfo1RequestedVersion1 = &endpointmanager.FHIREndpointInfo{
		URL:                   "https://example.com",
		VendorID:              fakeVendor.ID,
		TLSVersion:            "TLS 1.1",
		MIMETypes:             []string{"application/json+fhir"},
		CapabilityStatement:   nil,
		SMARTResponse:         nil,
		RequestedFhirVersion:  "1.0.0",
		CapabilityFhirVersion: "1.0.2",
		Metadata:              endpointMetadataRequestVersion1}

	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	// add endpointInfos and Metadata
	metadataID, err := store.AddFHIREndpointMetadata(ctx, endpointInfoDefaultRequest.Metadata)
	if err != nil {
		t.Errorf("Error adding fhir endpointMetadata: %s", err.Error())
	}

	valResID, err := store.AddValidationResult(ctx)
	th.Assert(t, err == nil, fmt.Sprintf("Error adding validation result ID: %s", err))
	endpointInfoDefaultRequest.ValidationID = valResID

	err = store.AddFHIREndpointInfo(ctx, endpointInfoDefaultRequest, metadataID)
	if err != nil {
		t.Errorf("Error adding fhir endpointInfo: %s", err.Error())
	}

	metadataID, err = store.AddFHIREndpointMetadata(ctx, endpointInfo1RequestedVersion1.Metadata)
	if err != nil {
		t.Errorf("Error adding fhir endpointMetadata: %s", err.Error())
	}

	valResID, err = store.AddValidationResult(ctx)
	th.Assert(t, err == nil, fmt.Sprintf("Error adding validation result ID: %s", err))
	endpointInfo1RequestedVersion1.ValidationID = valResID

	err = store.AddFHIREndpointInfo(ctx, endpointInfo1RequestedVersion1, metadataID)
	if err != nil {
		t.Errorf("Error adding fhir endpointInfo: %s", err.Error())
	}

	var count int
	// check insertions in info table
	row := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info;")
	err = row.Scan(&count)
	if err != nil {
		t.Errorf("info count for insertions: %s", err.Error())
	}
	if count != 2 {
		t.Errorf("expected 2 entries in info table . Got %d.", count)
	}
	// check insertions in info history table
	row = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info_history;")
	err = row.Scan(&count)
	if err != nil {
		t.Errorf("info history count for insertions: %s", err.Error())
	}
	if count != 2 {
		t.Errorf("expected 2 entries in info history table . Got %d.", count)
	}

	// If the supported versions array matches all entries in the info table, nothing should be deleted
	removeNoLongerExistingVersionsInfos(ctx, store, "https://example.com", []string{"None", "1.0.0"})
	// make sure nothing was removed
	row = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info;")
	err = row.Scan(&count)
	if err != nil {
		t.Errorf("info count for insertions: %s", err.Error())
	}
	if count != 2 {
		t.Errorf("expected 2 entries in info table . Got %d.", count)
	}
	// check that info history table still only has 2 entries
	row = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info_history;")
	err = row.Scan(&count)
	if err != nil {
		t.Errorf("info history count for insertions: %s", err.Error())
	}
	if count != 2 {
		t.Errorf("expected 2 entries in info history table . Got %d.", count)
	}

	// This simulates a change in advertised supported versions from ["None","1.0.0"] to ["None","2.0.0"]
	// This means we expect the existing 1.0.0 entry to be removed
	removeNoLongerExistingVersionsInfos(ctx, store, "https://example.com", []string{"None", "2.0.0"})
	row = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info;")
	err = row.Scan(&count)
	if err != nil {
		t.Errorf("info count for insertions: %s", err.Error())
	}
	if count != 1 {
		t.Errorf("expected 1 entry in info table . Got %d.", count)
	}
	// Make sure the specific info entry was removed
	row = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info WHERE fhir_endpoints_info.requested_fhir_version='1.0.0';")
	err = row.Scan(&count)
	if err != nil {
		t.Errorf("info count for insertions: %s", err.Error())
	}
	if count != 0 {
		t.Errorf("expected 0 entries in info table . Got %d.", count)
	}
	// check deletion in info history table
	row = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info_history WHERE operation='D' AND requested_fhir_version='1.0.0';")
	err = row.Scan(&count)
	if err != nil {
		t.Errorf("info histry count for insertions: %s", err.Error())
	}
	// There should now be a deletion entry
	if count != 1 {
		t.Errorf("expected 1 entries in info table . Got %d.", count)
	}

}

func Test_saveMsgInDB(t *testing.T) {
	err := setup()
	if err != nil {
		panic(err)
	}
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	setupCapabilityStatement(t, filepath.Join("../../testdata", "cerner_capability_dstu2.json"))

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM fhir_endpoints_info;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()
	ctx := context.Background()

	args := make(map[string]interface{})
	args["queryArgs"] = capStatQueryArgs{
		store:         store,
		ctx:           ctx,
		chplMatchFile: "../../testdata/test_chpl_product_mapping.json",
		chplEndpointListInfoFile: "../../testdata/test_chpl_products_info.json",
	}

	// populate vendors
	for _, vendor := range vendors {
		err = store.AddVendor(ctx, vendor)
		th.Assert(t, err == nil, err)
	}

	// add fhir endpoint with url
	err = store.AddFHIREndpoint(ctx, testFhirEndpoint1)
	th.Assert(t, err == nil, err)
	err = store.AddFHIREndpoint(ctx, testFhirEndpoint2)
	th.Assert(t, err == nil, err)

	expectedEndpt := testFhirEndpointInfo
	expectedEndpt.VendorID = vendors[1].ID // "Cerner Corporation"
	expectedEndpt.URL = testFhirEndpoint1.URL
	expectedMetadata := testFhirEndpointMetadata
	expectedEndpt.Metadata = &expectedMetadata
	queueTmp := testQueueMsg

	queueMsg, err := convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)

	// check that nothing is stored and that saveMsgInDB throws an error if the context is canceled
	testCtx, cancel := context.WithCancel(context.Background())
	args["queryArgs"] = capStatQueryArgs{
		store:         store,
		ctx:           testCtx,
		chplMatchFile: "../../testdata/test_chpl_product_mapping.json",
		chplEndpointListInfoFile: "../../testdata/test_chpl_products_info.json",
	}
	cancel()
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, errors.Cause(err) == context.Canceled, fmt.Sprintf("should have errored out with root cause that the context was canceled, instead was %s and %s", err, errors.Cause(err)))

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 0, "should not have stored data")

	// reset context
	args["queryArgs"] = capStatQueryArgs{
		store:         store,
		ctx:           context.Background(),
		chplMatchFile: "../../testdata/test_chpl_product_mapping.json",
		chplEndpointListInfoFile: "../../testdata/test_chpl_products_info.json",
	}
	// check that new item is stored
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, errors.Wrap(err, "error"))

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")

	// Get validation result ID, there should only be one ID
	var valID1 int
	valResRows := store.DB.QueryRow("SELECT id FROM validation_results")
	err = valResRows.Scan(&valID1)
	th.Assert(t, err == nil, err)
	expectedEndpt.ValidationID = valID1

	storedEndpt, err := store.GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx, testFhirEndpoint1.URL, "None")
	th.Assert(t, err == nil, err)
	th.Assert(t, expectedEndpt.Equal(storedEndpt), "stored data does not equal expected store data")

	// check that endpoint availability was updated
	var http_200_ct int
	var http_all_ct int
	var endpt_availability_ct int
	query_str := "SELECT http_200_count, http_all_count from fhir_endpoints_availability WHERE url=$1 AND requested_fhir_version='None';"
	ct_availability_str := "SELECT COUNT(*) from fhir_endpoints_availability;"

	err = store.DB.QueryRow(ct_availability_str).Scan(&endpt_availability_ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, endpt_availability_ct == 1, "endpoint availability should have 1 endpoint")
	err = store.DB.QueryRow(query_str, testFhirEndpoint1.URL).Scan(&http_200_ct, &http_all_ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, http_all_ct == 1, "endpoint should have http return count of 1")
	th.Assert(t, http_200_ct == 1, "endpoint should have http 200 return count of 1")

	// check that the validation table entries exist
	var validationCount int
	valResRows = store.DB.QueryRow("SELECT COUNT(*) FROM validations WHERE validation_result_id=$1", valID1)
	err = valResRows.Scan(&validationCount)
	th.Assert(t, err == nil, err)
	th.Assert(t, validationCount == 7, fmt.Sprintf("Should be 7 validation entries for ID %d, is instead %d", valID1, validationCount))

	// check that a second new item is stored
	queueTmp["url"] = "https://test-two.com"
	expectedEndpt.URL = testFhirEndpoint2.URL
	expectedEndpt.Metadata.URL = testFhirEndpoint2.URL
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "there should be two endpoints in the database")

	// Get validation result ID for second item
	var valID2 int
	valResRows = store.DB.QueryRow("SELECT id FROM validation_results ORDER BY id DESC LIMIT 1")
	err = valResRows.Scan(&valID2)
	th.Assert(t, err == nil, err)
	expectedEndpt.ValidationID = valID2

	storedEndpt, err = store.GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx, testFhirEndpoint2.URL, "None")
	th.Assert(t, err == nil, err)
	th.Assert(t, expectedEndpt.Equal(storedEndpt), "the second endpoint data does not equal expected store data")
	expectedEndpt.URL = testFhirEndpoint1.URL
	queueTmp["url"] = "http://example.com/DTSU2/"

	// check that a second endpoint also added to availability table
	err = store.DB.QueryRow(ct_availability_str).Scan(&endpt_availability_ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, endpt_availability_ct == 2, "endpoint availability should have 2 endpoints")
	err = store.DB.QueryRow(query_str, testFhirEndpoint2.URL).Scan(&http_200_ct, &http_all_ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, http_all_ct == 1, "endpoint should http return count of 1")
	th.Assert(t, http_200_ct == 1, "endpoint should have http 200 return count of 1")

	// check that the validation table entries exist
	valResRows = store.DB.QueryRow("SELECT COUNT(*) FROM validations WHERE validation_result_id=$1", valID2)
	err = valResRows.Scan(&validationCount)
	th.Assert(t, err == nil, err)
	th.Assert(t, validationCount == 7, fmt.Sprintf("Should be 7 validation entries for ID %d, is instead %d", valID2, validationCount))

	// check that an item with the same URL updates the endpoint in the database
	queueTmp["tlsVersion"] = "TLS 1.3"
	queueTmp["httpResponse"] = 404
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not store data as expected")

	storedEndpt, err = store.GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx, testFhirEndpoint1.URL, "None")
	th.Assert(t, err == nil, err)
	th.Assert(t, storedEndpt.TLSVersion == "TLS 1.3", "The TLS Version was not updated")
	th.Assert(t, storedEndpt.Metadata.HTTPResponse == 404, "The http response was not updated")

	// check that availability is updated
	err = store.DB.QueryRow(query_str, testFhirEndpoint1.URL).Scan(&http_200_ct, &http_all_ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, http_all_ct == 2, "http all count should have been incremented to 2, was %d")
	th.Assert(t, storedEndpt.Metadata.Availability == 0.5, "endpoint availability should have been updated to 0.5")

	// Check that the updated entry has new validation ID
	var valID3 int
	valResRows = store.DB.QueryRow("SELECT id FROM validation_results ORDER BY id DESC LIMIT 1")
	err = valResRows.Scan(&valID3)
	th.Assert(t, err == nil, err)
	th.Assert(t, valID2 != valID3, "No new validation ID was added to the validation_results table")

	queueTmp["tlsVersion"] = "TLS 1.2" // resetting value
	queueTmp["httpResponse"] = 200

	// check that error adding to store throws error
	queueTmp["url"] = "https://a-new-url.com"
	queueTmp["tlsVersion"] = strings.Repeat("a", 510) // too long. causes db error

	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err != nil, "expected error adding product")

	// resetting values
	queueTmp["url"] = "http://example.com/DTSU2/"
	queueTmp["tlsVersion"] = "TLS 1.2"

	// test selective update

	historySQLStatement := "SELECT updated_at FROM fhir_endpoints_info_history WHERE URL = $1 ORDER BY updated_at DESC LIMIT 1"
	var updatedAt time.Time

	// Update endpoint back to original values
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not store data as expected")

	storedEndpt, err = store.GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx, testFhirEndpoint1.URL, "None")
	th.Assert(t, err == nil, err)

	store.DB.QueryRow(historySQLStatement, storedEndpt.URL).Scan(&updatedAt)
	oldUpdateAt := updatedAt
	oldMetadataID := storedEndpt.Metadata.ID
	oldMetadataUpdatedAt := storedEndpt.Metadata.UpdatedAt
	oldValidationID := storedEndpt.ValidationID

	// Try to update with exact same values
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not store data as expected")
	storedEndpt, err = store.GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx, testFhirEndpoint1.URL, "None")
	th.Assert(t, err == nil, err)
	th.Assert(t, storedEndpt.Metadata.ID != oldMetadataID, "The selective update should have still updated the old endpoint info metadata id")
	th.Assert(t, storedEndpt.ValidationID == oldValidationID, fmt.Sprintf("The selective update should not have updated the old endpoint validation id for same values, %+v, %d", storedEndpt, oldValidationID))
	th.Assert(t, !storedEndpt.Metadata.UpdatedAt.Equal(oldMetadataUpdatedAt), "The selective update should have still updated the old endpoint metadata updated at time")

	store.DB.QueryRow(historySQLStatement, storedEndpt.URL).Scan(&updatedAt)
	th.Assert(t, updatedAt.Equal(oldUpdateAt), fmt.Sprintf("The selective update should not have updated the old endpoint updated at time in the history table, \n current: %s, \n old: %s", updatedAt, oldUpdateAt))

	oldMetadataID = storedEndpt.Metadata.ID
	oldMetadataUpdatedAt = storedEndpt.Metadata.UpdatedAt
	oldValidationID = storedEndpt.ValidationID

	// Try to update with exact same values besides metadata
	queueTmp["responseTime"] = 0.3456
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not store data as expected")

	storedEndpt, err = store.GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx, testFhirEndpoint1.URL, "None")
	th.Assert(t, err == nil, err)
	th.Assert(t, storedEndpt.Metadata.ID != oldMetadataID, "The selective update should have still updated the old endpoint info metadata id")
	th.Assert(t, storedEndpt.ValidationID == oldValidationID, "The selective update should not have updated the old endpoint validation id for same values minus metadata")
	th.Assert(t, !storedEndpt.Metadata.UpdatedAt.Equal(oldMetadataUpdatedAt), "The selective update should have still updated the old endpoint metadata updated at time")

	store.DB.QueryRow(historySQLStatement, storedEndpt.URL).Scan(&updatedAt)
	th.Assert(t, updatedAt.Equal(oldUpdateAt), "The selective update should not have updated the old endpoint updated at time in the history table")

	queueTmp["responseTime"] = 0.1234

}

func setup() error {
	var err error
	store, err = postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	if err != nil {
		return err
	}

	return nil
}

func teardown() {
	store.Close()
}
