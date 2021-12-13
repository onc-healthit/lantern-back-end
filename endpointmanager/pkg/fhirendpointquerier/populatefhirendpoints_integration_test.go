// +build integration

package populatefhirendpoints

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fetcher"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/pkg/errors"

	"github.com/spf13/viper"

	"strings"
)

var store *postgresql.Store
var testEndpointEntry2 fetcher.EndpointEntry = fetcher.EndpointEntry{
	OrganizationNames:    []string{"Access Community Health Network"},
	FHIRPatientFacingURI: "https://eprescribing.accesscommunityhealth.net/FHIR/api/FHIR/DSTU2/",
	ListSource:           "epicList",
}
var testEndpointEntry3 fetcher.EndpointEntry = fetcher.EndpointEntry{
	OrganizationNames:    []string{"fakeOrganization"},
	FHIRPatientFacingURI: "http://example.com/DTSU2/",
	ListSource:           "Lantern",
	NPIIDs:               []string{"1"},
}
var testEndpointEntry4 fetcher.EndpointEntry = fetcher.EndpointEntry{
	OrganizationNames:    []string{"fakeOrganization2"},
	FHIRPatientFacingURI: "http://example.com/DTSU2/",
	ListSource:           "Lantern",
	NPIIDs:               []string{"2"},
}
var testFHIREndpoint2 endpointmanager.FHIREndpoint = endpointmanager.FHIREndpoint{
	OrganizationNames: []string{"Access Community Health Network"},
	URL:               "https://eprescribing.accesscommunityhealth.net/FHIR/api/FHIR/DSTU2/",
	ListSource:        "epicList",
}
var testFHIREndpoint3 endpointmanager.FHIREndpoint = endpointmanager.FHIREndpoint{
	OrganizationNames: []string{"fakeOrganization"},
	URL:               "http://example.com/DTSU2/",
	ListSource:        "Lantern",
	NPIIDs:            []string{"1"},
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

func Test_Integration_AddEndpointData(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	var actualNumEndptsStored int

	ctx := context.Background()
	expectedNumEndptsStored := 340

	var listOfEndpoints, listErr = fetcher.GetEndpointsFromFilepath("../../resources/EpicEndpointSources.json", "Epic", "Epic", "https://open.epic.com/MyApps/EndpointsJson")
	th.Assert(t, listErr == nil, "Endpoint List Parsing Error")

	err = AddEndpointData(ctx, store, &listOfEndpoints)
	th.Assert(t, err == nil, err)
	rows := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints;")
	err = rows.Scan(&actualNumEndptsStored)
	th.Assert(t, err == nil, err)
	th.Assert(t, actualNumEndptsStored >= expectedNumEndptsStored, fmt.Sprintf("Expected at least %d endpoints stored. Actually had %d endpoints stored.", expectedNumEndptsStored, actualNumEndptsStored))

	// based on this entry in the DB:
	// {
	//	"OrganizationName":"AdvantageCare Physicians",
	// 	"FHIRPatientFacingURI":"https://epwebapps.acpny.com/FHIRproxy/api/FHIR/DSTU2/"
	// }
	fhirEndpt, err := store.GetFHIREndpointUsingURLAndListSource(ctx, "https://epwebapps.acpny.com/FHIRproxy/api/FHIR/DSTU2/", "https://open.epic.com/MyApps/EndpointsJson")
	th.Assert(t, err == nil, err)
	th.Assert(t, fhirEndpt.URL == "https://epwebapps.acpny.com/FHIRproxy/api/FHIR/DSTU2/", "URL is not what was expected")
	th.Assert(t, helpers.StringArraysEqual(fhirEndpt.OrganizationNames, []string{"AdvantageCare Physicians"}), "Organization Name is not what was expected.")

	// Test that when updating endpoints from same listsource, old endpoints are removed based on update time
	// This endpoint list has 10 endpoints removed from it
	listOfEndpoints, listErr = fetcher.GetEndpointsFromFilepath("../../resources/EpicEndpointSources_1.json", "Epic", "Epic", "https://open.epic.com/MyApps/EndpointsJson")
	th.Assert(t, listErr == nil, "Endpoint List Parsing Error")
	err = AddEndpointData(ctx, store, &listOfEndpoints)
	th.Assert(t, err == nil, err)
	rows = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints;")
	err = rows.Scan(&actualNumEndptsStored)
	th.Assert(t, err == nil, err)
	th.Assert(t, actualNumEndptsStored >= expectedNumEndptsStored-10, fmt.Sprintf("Expected at least %d endpoints stored. Actually had %d endpoints stored.", expectedNumEndptsStored-10, actualNumEndptsStored))
	// This endpoint should be removed from table
	fhirEndpt, err = store.GetFHIREndpointUsingURLAndListSource(ctx, "https://epwebapps.acpny.com/FHIRproxy/api/FHIR/DSTU2/", "https://open.epic.com/MyApps/EndpointsJson")
	th.Assert(t, err == sql.ErrNoRows, err)
}

func Test_saveEndpointData(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	endpt := testEndpointEntry
	fhirEndpt := testFHIREndpoint
	endptLantern := testEndpointEntry3
	endptLantern2 := testEndpointEntry4
	fhirEndptLantern := testFHIREndpoint3
	var savedEndpt *endpointmanager.FHIREndpoint

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM fhir_endpoints;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()

	// check that nothing is stored and that saveEndpointData throws an error if the context is canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = saveEndpointData(ctx, store, &endpt)
	th.Assert(t, errors.Cause(err) == context.Canceled, "should have errored out with root cause that the context was canceled")

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 0, "should not have stored data")

	// reset context
	ctx = context.Background()

	// check that new item is stored
	err = saveEndpointData(ctx, store, &endpt)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")

	var endptID int
	err = store.DB.QueryRow("SELECT id FROM fhir_endpoints LIMIT 1;").Scan(&endptID)
	th.Assert(t, err == nil, err)
	savedEndpt, err = store.GetFHIREndpoint(ctx, endptID)
	th.Assert(t, err == nil, err)
	th.Assert(t, fhirEndpt.Equal(savedEndpt), "stored data does not equal expected store data")

	// check that an item with the same URL replaces item and merges the organization names lists
	endpt.OrganizationNames = []string{"AdvantageCare Physicians 2"}
	err = saveEndpointData(ctx, store, &endpt)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")

	err = store.DB.QueryRow("SELECT id FROM fhir_endpoints LIMIT 1;").Scan(&endptID)
	th.Assert(t, err == nil, err)
	savedEndpt, err = store.GetFHIREndpoint(ctx, endptID)
	th.Assert(t, err == nil, err)

	th.Assert(t, helpers.StringArraysEqual(savedEndpt.OrganizationNames, []string{"AdvantageCare Physicians", "AdvantageCare Physicians 2"}),
		fmt.Sprintf("stored data %v does not equal expected store data [AdvantageCare Physicians, AdvantageCare Physicians 2]", savedEndpt.OrganizationNames))

	// reset context
	ctx = context.Background()

	// reset values
	_, err = store.DB.Exec("DELETE FROM fhir_endpoints;")
	th.Assert(t, err == nil, err)

	// check that new item is stored
	err = saveEndpointData(ctx, store, &endptLantern)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")

	err = store.DB.QueryRow("SELECT id FROM fhir_endpoints LIMIT 1;").Scan(&endptID)
	th.Assert(t, err == nil, err)
	savedEndpt, err = store.GetFHIREndpoint(ctx, endptID)
	th.Assert(t, err == nil, err)
	th.Assert(t, fhirEndptLantern.Equal(savedEndpt), "stored data does not equal expected store data")

	// check that an lantern source endpoint with the same URL replaces item and merges the organization names lists and npi lists
	err = saveEndpointData(ctx, store, &endptLantern2)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")

	err = store.DB.QueryRow("SELECT id FROM fhir_endpoints LIMIT 1;").Scan(&endptID)
	th.Assert(t, err == nil, err)
	savedEndpt, err = store.GetFHIREndpoint(ctx, endptID)
	th.Assert(t, err == nil, err)

	th.Assert(t, helpers.StringArraysEqual(savedEndpt.OrganizationNames, []string{"fakeOrganization", "fakeOrganization2"}),
		fmt.Sprintf("stored data %v does not equal expected store data [fakeOrganization, fakeOrganization2]", savedEndpt.OrganizationNames))

	th.Assert(t, helpers.StringArraysEqual(savedEndpt.NPIIDs, []string{"1", "2"}),
		fmt.Sprintf("stored data %v does not equal expected store data [fakeOrganization, fakeOrganization2]", savedEndpt.OrganizationNames))

	// check that error adding to store throws error
	endpt = testEndpointEntry
	endpt.FHIRPatientFacingURI = "http://a-new-url.com/metadata/"
	endpt.ListSource = strings.Repeat("a", 510) // length is too long - causes an error on entry
	err = saveEndpointData(ctx, store, &endpt)
	th.Assert(t, err != nil, "expected error adding product")
}

func Test_AddEndpointData(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	endpt1 := testEndpointEntry
	endpt2 := testEndpointEntry

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM fhir_endpoints;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()

	endpt2.FHIRPatientFacingURI = "http://a-new-url.com/metadata/"
	listEndpoints := fetcher.ListOfEndpoints{Entries: []fetcher.EndpointEntry{endpt1, endpt2}}
	expectedEndptsStored := 2

	// check that nothing is stored and that AddEndpointData throws an error if the context is canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = AddEndpointData(ctx, store, &listEndpoints)
	th.Assert(t, errors.Cause(err) == context.Canceled, "should have errored out with root cause that the context was canceled")

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 0, "should not have stored data")

	// reset context
	ctx = context.Background()

	// check basic functionality
	err = AddEndpointData(ctx, store, &listEndpoints)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)

	th.Assert(t, ct == expectedEndptsStored, fmt.Sprintf("Expected %d products stored. Actually had %d products stored.", expectedEndptsStored, ct))
	storedEndpt, err := store.GetFHIREndpointUsingURLAndListSource(ctx, endpt1.FHIRPatientFacingURI, endpt1.ListSource)
	th.Assert(t, err == nil, err)
	th.Assert(t, storedEndpt != nil, "Did not store first product as expected")
	storedEndpt, err = store.GetFHIREndpointUsingURLAndListSource(ctx, endpt2.FHIRPatientFacingURI, endpt2.ListSource)
	th.Assert(t, err == nil, err)
	th.Assert(t, storedEndpt != nil, "Did not store second product as expected")

	// reset values
	_, err = store.DB.Exec("DELETE FROM fhir_endpoints;")
	th.Assert(t, err == nil, err)

	endpt2 = testEndpointEntry
	endpt2.OrganizationNames = []string{"New Name"}
	// endpt1 and endpt2 identical other than organization name.
	// endpt1 has organization name "AdvantageCare Physicians"
	listEndpoints = fetcher.ListOfEndpoints{Entries: []fetcher.EndpointEntry{endpt1, endpt2}}
	err = AddEndpointData(ctx, store, &listEndpoints)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not persist one product as expected")

	storedEndpt, err = store.GetFHIREndpointUsingURLAndListSource(ctx, endpt1.FHIRPatientFacingURI, endpt1.ListSource)
	th.Assert(t, err == nil, err)
	th.Assert(t, helpers.StringArraysEqual(storedEndpt.OrganizationNames, []string{"AdvantageCare Physicians", "New Name"}),
		fmt.Sprintf("stored data '%v' does not equal expected store data '%v'", storedEndpt.OrganizationNames, endpt2.OrganizationNames))

	endpt2 = testEndpointEntry2
	listEndpoints = fetcher.ListOfEndpoints{Entries: []fetcher.EndpointEntry{endpt2}}
	err = AddEndpointData(ctx, store, &listEndpoints)
	th.Assert(t, err == nil, err)
	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, fmt.Sprintf("expected one endpoint after update, got %v", ct))
	storedEndpt, err = store.GetFHIREndpointUsingURLAndListSource(ctx, endpt1.FHIRPatientFacingURI, endpt1.ListSource)
	th.Assert(t, err == sql.ErrNoRows, "Endpoint should be deleted")
	storedEndpt, err = store.GetFHIREndpointUsingURLAndListSource(ctx, endpt2.FHIRPatientFacingURI, endpt2.ListSource)
	th.Assert(t, helpers.StringArraysEqual(storedEndpt.OrganizationNames, []string{"Access Community Health Network"}),
		fmt.Sprintf("stored data '%v' does not equal expected store data '%v'", storedEndpt.OrganizationNames, endpt2.OrganizationNames))
}

func Test_RemoveOldEndpoints(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	endpt1 := testFHIREndpoint
	endpt2 := testFHIREndpoint
	endpt3 := testFHIREndpoint2

	ctx := context.Background()

	query_str := "SELECT COUNT(*) FROM fhir_endpoints;"
	var ct int
	// Add first endpoint
	err = store.AddFHIREndpoint(ctx, &endpt1)
	th.Assert(t, err == nil, err)
	err = store.DB.QueryRow(query_str).Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not persist endpoint as expected")
	var endptID int
	err = store.DB.QueryRow("SELECT id FROM fhir_endpoints LIMIT 1;").Scan(&endptID)
	th.Assert(t, err == nil, err)
	savedEndpt, err := store.GetFHIREndpoint(ctx, endptID)
	th.Assert(t, err == nil, err)
	th.Assert(t, endpt1.Equal(savedEndpt), "stored data does not equal expected store data")

	// Add endpoint with same url but different listsource
	endpt2.ListSource = "test"
	err = store.AddFHIREndpoint(ctx, &endpt2)
	err = store.DB.QueryRow(query_str).Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not persist second endpoint as expected")

	endptInfo := endpointmanager.FHIREndpointInfo{
		URL:                  endpt2.URL,
		RequestedFhirVersion: "None",
		Metadata: &endpointmanager.FHIREndpointMetadata{
			URL:                  endpt2.URL,
			HTTPResponse:         200,
			RequestedFhirVersion: "None",
		},
	}
	metadataID, err := store.AddFHIREndpointMetadata(ctx, endptInfo.Metadata)
	valResID1, err := store.AddValidationResult(ctx)
	th.Assert(t, err == nil, fmt.Sprintf("Error adding validation result ID: %s", err))
	endptInfo.ValidationID = valResID1
	err = store.AddFHIREndpointInfo(ctx, &endptInfo, metadataID)
	th.Assert(t, err == nil, err)

	// Add third endpoint
	err = store.AddFHIREndpoint(ctx, &endpt3)
	th.Assert(t, err == nil, err)
	err = store.DB.QueryRow(query_str).Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 3, "did not persist third endpoint as expected")
	savedEndpt, err = store.GetFHIREndpointUsingURLAndListSource(ctx, endpt3.URL, endpt3.ListSource)
	th.Assert(t, err == nil, err)
	th.Assert(t, endpt3.Equal(savedEndpt), "stored data does not equal expected store data")

	err = RemoveOldEndpoints(ctx, store, savedEndpt.UpdatedAt, endpt3.ListSource)
	th.Assert(t, err == nil, err)
	err = store.DB.QueryRow(query_str).Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "Expected only one endpoint to deleted")
	// Check that first endpoint is removed based on update time
	_, err = store.GetFHIREndpointUsingURLAndListSource(ctx, endpt1.URL, endpt1.ListSource)
	th.Assert(t, err == sql.ErrNoRows, "Expected endpoint to removed")
	// Check that second endpoint still exist
	_, err = store.GetFHIREndpointUsingURLAndListSource(ctx, endpt2.URL, endpt2.ListSource)
	th.Assert(t, err == nil, "Endpoint should still exist from different listsource")
	// Test that endpoint is not removed from fhir_endpoints_info because it still exist in
	// fhir_endpoints but from different listsource
	FHIREndpointInfo, err := store.GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx, endpt2.URL, "None")
	th.Assert(t, err == nil, "Expected endpoint to still persist in fhir_endpoints_info")
	// Test that endpoint is not removed from fhir_endpoints_metadata because it still exist in
	// fhir_endpoints but from different listsource
	_, err = store.GetFHIREndpointMetadata(ctx, FHIREndpointInfo.Metadata.ID)
	th.Assert(t, err == nil, "Expected endpoint to still persist in fhir_endpoints_metadata")

	// reset values
	_, err = store.DB.Exec("DELETE FROM fhir_endpoints_info;")
	th.Assert(t, err == nil, err)
	_, err = store.DB.Exec("DELETE FROM fhir_endpoints;")
	th.Assert(t, err == nil, err)
	_, err = store.DB.Exec("DELETE FROM fhir_endpoints_metadata;")
	th.Assert(t, err == nil, err)

	endptInfo2 := endpointmanager.FHIREndpointInfo{
		URL:                  endpt1.URL,
		RequestedFhirVersion: "1.0.2",
		Metadata: &endpointmanager.FHIREndpointMetadata{
			URL:                  endpt1.URL,
			HTTPResponse:         200,
			RequestedFhirVersion: "1.0.2",
		},
	}

	// Add one endpoint
	err = store.AddFHIREndpoint(ctx, &endpt1)
	th.Assert(t, err == nil, err)
	err = store.DB.QueryRow(query_str).Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not persist first endpoint as expected")
	savedEndpt, err = store.GetFHIREndpointUsingURLAndListSource(ctx, endpt1.URL, endpt1.ListSource)
	th.Assert(t, err == nil, err)
	th.Assert(t, endpt1.Equal(savedEndpt), "stored data does not equal expected store data")

	metadataID, err = store.AddFHIREndpointMetadata(ctx, endptInfo2.Metadata)
	valResID2, err := store.AddValidationResult(ctx)
	th.Assert(t, err == nil, fmt.Sprintf("Error adding validation result ID: %s", err))
	endptInfo2.ValidationID = valResID2
	err = store.AddFHIREndpointInfo(ctx, &endptInfo2, metadataID)
	th.Assert(t, err == nil, err)

	endptInfo2.RequestedFhirVersion = "4.0.0"
	endptInfo2.Metadata.RequestedFhirVersion = "4.0.0"
	valResID2, err = store.AddValidationResult(ctx)
	th.Assert(t, err == nil, fmt.Sprintf("Error adding validation result ID: %s", err))
	endptInfo2.ValidationID = valResID2

	metadataID, err = store.AddFHIREndpointMetadata(ctx, endptInfo2.Metadata)
	err = store.AddFHIREndpointInfo(ctx, &endptInfo2, metadataID)
	th.Assert(t, err == nil, err)

	endptInfo2.RequestedFhirVersion = "1.0.2"
	endptInfo2.Metadata.RequestedFhirVersion = "1.0.2"

	err = RemoveOldEndpoints(ctx, store, savedEndpt.UpdatedAt.Add(time.Hour*1), endpt1.ListSource)
	th.Assert(t, err == nil, err)
	err = store.DB.QueryRow(query_str).Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 0, "Expected the one endpoint to be deleted")
	// Test that the two endpoints are removed from fhir_endpoints_info even though they have different
	// requested fhir versions
	err = store.DB.QueryRow("SELECT count(*) FROM fhir_endpoints_info").Scan(&ct)
	th.Assert(t, ct == 0, "Expected both endpoints to be removed from fhir endpoint info table")

	// Test that endpoints are not removed from fhir_endpoints_metadata since removing an endpoint from the info table
	// should not remove it from the metadata table
	err = store.DB.QueryRow("SELECT count(*) FROM fhir_endpoints_metadata").Scan(&ct)
	th.Assert(t, ct == 2, "Expected both endpoints to still be in the FHIR endpoint metadata table")

}

func setup() error {
	var err error
	store, err = postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	return err
}

func teardown() {
	store.Close()
}
