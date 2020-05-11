// +build integration

package nppesquerier_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"strconv"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/nppesquerier"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/pkg/errors"
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

	ctx := context.Background()

	parsed_orgs, err := nppesquerier.ParseAndStoreNPIFile(ctx, "testdata/npidata_pfile_fixture.csv", store)
	if err != nil {
		t.Errorf("Error Parsing NPI File: %s", err.Error())
	}
	// Assert expected number of orgs are parsed out of fixture file
	if parsed_orgs != 3 {
		t.Errorf("Expected number or parsed orgs to be %d, got %d", 3, parsed_orgs)
	}
	// Assert NPI orgs were successfully parsed out of fixture file
	org1, err := store.GetNPIOrganizationByNPIID(ctx, "1497758544")
	if org1 == nil {
		t.Errorf("Error Retriving Parsed NPI Org")
	}
	if err != nil {
		t.Errorf("Error Retriving Parsed NPI Org: %s", err.Error())
	}
	org2, err := store.GetNPIOrganizationByNPIID(ctx, "1023011178")
	if org2 == nil {
		t.Errorf("Error Retriving Parsed NPI Org")
	}
	if err != nil {
		t.Errorf("Error Retriving Parsed NPI Org: %s", err.Error())
	}
	org3, err := store.GetNPIOrganizationByNPIID(ctx, "1023011079")
	if org3 == nil {
		t.Errorf("Error Retriving Parsed NPI Org")
	}
	if err != nil {
		t.Errorf("Error Retriving Parsed NPI Org: %s", err.Error())
	}
}

func Test_ParseAndStoreNPIFileContext(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	// Note: it's possible that on a particularly slow or fast machine, this time deadline won't work.
	// need to set a deadline rather than call cancel so we get through the read of the csv file.
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Millisecond))
	defer cancel()

	added, err := nppesquerier.ParseAndStoreNPIFile(ctx, "testdata/npidata_pfile_fixture.csv", store)
	th.Assert(t, errors.Cause(err) == context.DeadlineExceeded, fmt.Sprintf("Expected canceled context error %+v. Got %+v\n", context.DeadlineExceeded, errors.Cause(err)))
	th.Assert(t, added >= 0, "expected items added to be zero or more after context deadline met")
}

func Test_ParseAndStoreNPIContactFile(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	ctx := context.Background()

	parsed_orgs, err := nppesquerier.ParseAndStoreNPIContactsFile(ctx, "testdata/npi_contact_file.csv", store)
	if err != nil {
		t.Errorf("Error Parsing NPI File: %s", err.Error())
	}
	th.Assert(t, parsed_orgs == 212, "Incorrect number of FHIR_URL entries parsed out of npi file " + strconv.Itoa(parsed_orgs))

	// Assert NPI orgs were successfully parsed out of fixture file
	bad_url_contact, err := store.GetNPIContactByNPIID(ctx, "1346747029")
	if err != nil {
		t.Errorf("Error Retriving Parsed NPI Contact: %s", err.Error())
	}
	th.Assert(t, bad_url_contact != nil, "Unable to find npi 1346747029 in npi_contacts")
	th.Assert(t, bad_url_contact.ValidURL == false, "Invalid URL marked as valid")
	th.Assert(t, bad_url_contact.Endpoint == "HIPPA and Health IT", "Unexpected Endpoint information for entry")

	good_url_contact, err := store.GetNPIContactByNPIID(ctx, "1760025803")
	if err != nil {
		t.Errorf("Error Retriving Parsed NPI Contact: %s", err.Error())
	}
	th.Assert(t, good_url_contact != nil, "Unable to find npi 1760025803 in npi_contacts")
	th.Assert(t, good_url_contact.ValidURL == true, "Valid URL marked as invalid")
	th.Assert(t, good_url_contact.Endpoint == "https://dpc.cms.gov/api/v1", "Unexpected Endpoint information for entry")

}

func Test_ParseAndStoreNPIContactFileContext(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	// Note: it's possible that on a particularly slow or fast machine, this time deadline won't work.
	// need to set a deadline rather than call cancel so we get through the read of the csv file.
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Millisecond))
	defer cancel()

	added, err := nppesquerier.ParseAndStoreNPIContactsFile(ctx, "testdata/npi_contact_file.csv", store)
	th.Assert(t, errors.Cause(err) == context.DeadlineExceeded, fmt.Sprintf("Expected canceled context error %+v. Got %+v\n", context.DeadlineExceeded, errors.Cause(err)))
	th.Assert(t, added >= 0, "expected items added to be zero or more after context deadline met")
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
