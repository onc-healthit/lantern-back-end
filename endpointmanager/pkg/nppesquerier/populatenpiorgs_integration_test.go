// +build integration

package nppesquerier_test

import (
	"os"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/nppesquerier"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
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

func Test_ParseAndStoreNPIFile(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	parsed_orgs, err := nppesquerier.ParseAndStoreNPIFile("testdata/npidata_pfile_fixture.csv", store)
	if err != nil {
		t.Errorf("Error Parsing NPI File: %s", err.Error())
	}
	// Assert expected number of orgs are parsed out of fixture file
	if parsed_orgs != 3 {
		t.Errorf("Expected number or parsed orgs to be %d, got %d", 3, parsed_orgs)
	}
	// Assert NPI orgs were successfully parsed out of fixture file
	org1, err := store.GetNPIOrganizationByNPIID("1497758544")
	if org1 == nil {
		t.Errorf("Error Retriving Parsed NPI Org")
	}
	if err != nil {
		t.Errorf("Error Retriving Parsed NPI Org: %s", err.Error())
	}
	org2, err := store.GetNPIOrganizationByNPIID("1023011178")
	if org2 == nil {
		t.Errorf("Error Retriving Parsed NPI Org")
	}
	if err != nil {
		t.Errorf("Error Retriving Parsed NPI Org: %s", err.Error())
	}
	org3, err := store.GetNPIOrganizationByNPIID("1023011079")
	if org3 == nil {
		t.Errorf("Error Retriving Parsed NPI Org")
	}
	if err != nil {
		t.Errorf("Error Retriving Parsed NPI Org: %s", err.Error())
	}
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
