// +build integration

package populatefhirendpoints

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"
	"github.com/pkg/errors"

	"github.com/spf13/viper"

	"strings"
)

var store *postgresql.Store

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
	// 	"organization_name": "A Woman's Place",
	//	"fhir_version": "",
	// 	"authorization_standard": "",
	//	"location": null,
	// 	"capability_statement": null,
	// }
	fhirEndpt, err := store.GetFHIREndpointUsingURL(ctx, "https://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2/")
	th.Assert(t, err == nil, err)
	th.Assert(t, fhirEndpt.URL == "https://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2/", "URL is not what was expected")
	th.Assert(t, fhirEndpt.OrganizationName == "A Woman's Place, LLC", "Organization Name is not what was expected.")
	th.Assert(t, fhirEndpt.FHIRVersion == "", "Fhir Version is not what was expected")
	th.Assert(t, fhirEndpt.AuthorizationStandard == "", "Authorization Standard is not what was expected")
	th.Assert(t, fhirEndpt.Location == nil, "Location is not what was expected")
	th.Assert(t, fhirEndpt.CapabilityStatement == nil, "Capability Statement is not what was expected")
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

	// check that an item with the same URL replaces item
	endpt.OrganizationName = "A Woman's Place 2"
	err = saveEndpointData(ctx, store, &endpt)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")

	err = store.DB.QueryRow("SELECT id FROM fhir_endpoints LIMIT 1;").Scan(&endptID)
	th.Assert(t, err == nil, err)
	savedEndpt, err = store.GetFHIREndpoint(ctx, endptID)
	th.Assert(t, err == nil, err)

	th.Assert(t, savedEndpt.OrganizationName == "A Woman's Place 2",
		fmt.Sprintf("stored data %s does not equal expected store data 'A Woman's Place 2", savedEndpt.OrganizationName))

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
	// fhirEndpt := testFHIREndpoint
	// var savedEndpt *endpointmanager.FHIREndpoint

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
	storedEndpt, err := store.GetFHIREndpointUsingURL(ctx, endpt1.FHIRPatientFacingURI)
	th.Assert(t, err == nil, err)
	th.Assert(t, storedEndpt != nil, "Did not store first product as expected")
	storedEndpt, err = store.GetFHIREndpointUsingURL(ctx, endpt2.FHIRPatientFacingURI)
	th.Assert(t, err == nil, err)
	th.Assert(t, storedEndpt != nil, "Did not store first product as expected")

	// reset values
	_, err = store.DB.Exec("DELETE FROM fhir_endpoints;")
	th.Assert(t, err == nil, err)

	endpt2 = testEndpointEntry
	endpt2.OrganizationName = "New Name"
	listEndpoints = fetcher.ListOfEndpoints{Entries: []fetcher.EndpointEntry{endpt1, endpt2}}
	err = AddEndpointData(ctx, store, &listEndpoints)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not persist one product as expected")

	storedEndpt, err = store.GetFHIREndpointUsingURL(ctx, endpt1.FHIRPatientFacingURI)
	th.Assert(t, err == nil, err)
	th.Assert(t, storedEndpt.OrganizationName == endpt2.OrganizationName, "stored data does not equal expected store data")
}

func setup() error {
	var err error
	store, err = postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	return err
}

func teardown() {
	store.Close()
}
