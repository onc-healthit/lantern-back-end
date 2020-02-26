// +build integration

package populatefhirendpoints_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	fhirquerier "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fhirendpointquerier"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"

	"github.com/spf13/viper"
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

	var listOfEndpoints, listErr = fetcher.GetListOfEndpoints("../../../networkstatsquerier/resources/EndpointSources.json")
	th.Assert(t, listErr == nil, "Endpoint List Parsing Error")

	err = fhirquerier.AddEndpointData(ctx, store, &listOfEndpoints)
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

func setup() error {
	var err error
	store, err = postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	return err
}

func teardown() {
	store.Close()
}
