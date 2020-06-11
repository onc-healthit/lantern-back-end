// +build integration

package populatefhirendpoints

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"
	"github.com/pkg/errors"

	"github.com/spf13/viper"

	"strings"
)

var store *postgresql.Store
var testEndpointEntry2 fetcher.EndpointEntry = fetcher.EndpointEntry{
	OrganizationNames:    []string{"Access Community Health Network"},
	FHIRPatientFacingURI: "https://eprescribing.accesscommunityhealth.net/FHIR/api/FHIR/DSTU2/",
	ListSource:           "CareEvolution",
}
var testFHIREndpoint2 endpointmanager.FHIREndpoint = endpointmanager.FHIREndpoint{
	OrganizationNames: []string{"Access Community Health Network"},
	URL:               "https://eprescribing.accesscommunityhealth.net/FHIR/api/FHIR/DSTU2/",
	ListSource:        "CareEvolution",
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

	var listOfEndpoints, listErr = fetcher.GetEndpointsFromFilepath("../../../networkstatsquerier/resources/EndpointSources.json", "CareEvolution")
	th.Assert(t, listErr == nil, "Endpoint List Parsing Error")

	err = AddEndpointData(ctx, store, &listOfEndpoints)
	th.Assert(t, err == nil, err)
	rows := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints;")
	err = rows.Scan(&actualNumEndptsStored)
	th.Assert(t, err == nil, err)
	th.Assert(t, actualNumEndptsStored >= expectedNumEndptsStored, fmt.Sprintf("Expected at least %d endpoints stored. Actually had %d endpoints stored.", expectedNumEndptsStored, actualNumEndptsStored))

	// based on this entry in the DB:
	// {
	//	"url": "https://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2/",
	// 	"organization_names": {"A Woman's Place"}
	// }
	fhirEndpt, err := store.GetFHIREndpointUsingURLAndListSource(ctx, "https://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2/", "CareEvolution")
	th.Assert(t, err == nil, err)
	th.Assert(t, fhirEndpt.URL == "https://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2/", "URL is not what was expected")
	th.Assert(t, helpers.StringArraysEqual(fhirEndpt.OrganizationNames, []string{"A Woman's Place, LLC"}), "Organization Name is not what was expected.")

	// Test that when updating endpoints from same listsource, old endpoints are removed based on update time
	// This endpoint list has 10 endpoints removed from it
	listOfEndpoints, listErr = fetcher.GetEndpointsFromFilepath("../../../networkstatsquerier/resources/EndpointSources_1.json", "CareEvolution")
	th.Assert(t, listErr == nil, "Endpoint List Parsing Error")

	err = AddEndpointData(ctx, store, &listOfEndpoints)
	th.Assert(t, err == nil, err)
	rows = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints;")
	err = rows.Scan(&actualNumEndptsStored)
	th.Assert(t, err == nil, err)
	th.Assert(t, actualNumEndptsStored >= expectedNumEndptsStored-10, fmt.Sprintf("Expected at least %d endpoints stored. Actually had %d endpoints stored.", expectedNumEndptsStored-10, actualNumEndptsStored))
	// This endpoint should be removed from table
	fhirEndpt, err = store.GetFHIREndpointUsingURLAndListSource(ctx, "https://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2/", "CareEvolution")
	th.Assert(t, err == sql.ErrNoRows, err)	
}

func Test_saveEndpointData(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	endpt := testEndpointEntry
	fhirEndpt := testFHIREndpoint
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
	endpt.OrganizationNames = []string{"A Woman's Place 2"}
	err = saveEndpointData(ctx, store, &endpt)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")

	err = store.DB.QueryRow("SELECT id FROM fhir_endpoints LIMIT 1;").Scan(&endptID)
	th.Assert(t, err == nil, err)
	savedEndpt, err = store.GetFHIREndpoint(ctx, endptID)
	th.Assert(t, err == nil, err)

	th.Assert(t, helpers.StringArraysEqual(savedEndpt.OrganizationNames, []string{"A Woman's Place", "A Woman's Place 2"}),
		fmt.Sprintf("stored data %v does not equal expected store data [A Woman's Place, A Woman's Place 2]", savedEndpt.OrganizationNames))

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
	// endpt1 has organization name "A Woman's Place"
	listEndpoints = fetcher.ListOfEndpoints{Entries: []fetcher.EndpointEntry{endpt1, endpt2}}
	err = AddEndpointData(ctx, store, &listEndpoints)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not persist one product as expected")

	storedEndpt, err = store.GetFHIREndpointUsingURLAndListSource(ctx, endpt1.FHIRPatientFacingURI, endpt1.ListSource)
	th.Assert(t, err == nil, err)
	th.Assert(t, helpers.StringArraysEqual(storedEndpt.OrganizationNames, []string{"A Woman's Place", "New Name"}),
		fmt.Sprintf("stored data '%v' does not equal expected store data '%v'", storedEndpt.OrganizationNames, endpt2.OrganizationNames))

	endpt2 = testEndpointEntry2
	listEndpoints = fetcher.ListOfEndpoints{Entries: []fetcher.EndpointEntry{endpt2}}
	err = AddEndpointData(ctx, store, &listEndpoints)
	th.Assert(t, err == nil, err)
	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "expected one endpoint after update")
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
	endpt2 := testFHIREndpoint2
	
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

	// Add second endpoint
	err = store.AddFHIREndpoint(ctx, &endpt2)
	th.Assert(t, err == nil, err)
	err = store.DB.QueryRow(query_str).Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not persist second endpoint as expected")
	savedEndpt, err = store.GetFHIREndpointUsingURLAndListSource(ctx, endpt2.URL, endpt2.ListSource)
	th.Assert(t, err == nil, err)
	th.Assert(t, endpt2.Equal(savedEndpt), "stored data does not equal expected store data")

	// Check that first endpoint is removed based on update time
	err = removeOldEndpoints(ctx, store, savedEndpt.UpdatedAt, endpt2.ListSource)
	th.Assert(t, err == nil, err)
	_, err = store.GetFHIREndpointUsingURLAndListSource(ctx, endpt1.URL, endpt1.ListSource)
	th.Assert(t, err == sql.ErrNoRows, "Expected endpoint to removed")
}

func setup() error {
	var err error
	store, err = postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	return err
}

func teardown() {
	store.Close()
}
