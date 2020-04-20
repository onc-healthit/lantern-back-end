// +build integration

package capabilityhandler

import (
	"context"
	"os"
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

var hitps []*endpointmanager.HealthITProduct = []*endpointmanager.HealthITProduct{
	&endpointmanager.HealthITProduct{
		Name:                 "EpicCare Ambulatory Base",
		Version:              "February 2020",
		Developer:            "Epic Systems Corporation",
		CertificationStatus:  "Active",
		CertificationDate:    time.Date(2020, 2, 20, 0, 0, 0, 0, time.UTC),
		CertificationEdition: "2015",
		CHPLID:               "15.04.04.1447.Epic.AM.13.1.200220",
		APIURL:               "https://open.epic.com/Interface/FHIR",
	},
	&endpointmanager.HealthITProduct{
		Name:                 "PowerChart (Clinical)",
		Version:              "2018.01",
		Developer:            "Cerner Corporation",
		CertificationStatus:  "Active",
		CertificationDate:    time.Date(2018, 7, 27, 0, 0, 0, 0, time.UTC),
		CertificationEdition: "2015",
		CHPLID:               "15.04.04.1221.Powe.18.03.1.180727",
		APIURL:               "http://fhir.cerner.com/authorization/",
	},
	&endpointmanager.HealthITProduct{
		Name:                 "Health Services Analytics",
		Version:              "8.00 SP1-SP5",
		Developer:            "Cerner Health Services, Inc.",
		CertificationStatus:  "Withdrawn by Developer",
		CertificationDate:    time.Date(2017, 12, 5, 0, 0, 0, 0, time.UTC),
		CertificationEdition: "2014",
		CHPLID:               "14.07.07.1222.HEA5.03.01.1.171205",
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

func Test_saveMsgInDB(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	setupCapabilityStatement(t)

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM fhir_endpoints;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()

	args := make(map[string]interface{})
	args["store"] = store

	ctx := context.Background()

	// populate healthit products
	for _, hitp := range hitps {
		err = store.AddHealthITProduct(ctx, hitp)
	}

	expectedEndpt := testFhirEndpoint
	expectedEndpt.Vendor = "Cerner Corporation"
	queueTmp := testQueueMsg

	queueMsg, err := convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)

	// check that nothing is stored and that saveMsgInDB throws an error if the context is canceled
	testCtx, cancel := context.WithCancel(context.Background())
	args["ctx"] = testCtx
	cancel()
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, errors.Cause(err) == context.Canceled, "should have errored out with root cause that the context was canceled")

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 0, "should not have stored data")

	// reset context
	args["ctx"] = context.Background()

	// check that new item is stored
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")

	storedEndpt, err := store.GetFHIREndpointUsingURL(ctx, expectedEndpt.URL)
	th.Assert(t, err == nil, err)
	th.Assert(t, expectedEndpt.Equal(storedEndpt), "stored data does not equal expected store data")

	// check that a second new item is stored
	queueTmp["url"] = "https://test-two.com"
	expectedEndpt.URL = "https://test-two.com"
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "there should be two endpoints in the database")

	storedEndpt, err = store.GetFHIREndpointUsingURL(ctx, expectedEndpt.URL)
	th.Assert(t, err == nil, err)
	th.Assert(t, expectedEndpt.Equal(storedEndpt), "the second endpoint data does not equal expected store data")
	expectedEndpt = testFhirEndpoint
	queueTmp["url"] = "http://example.com/DTSU2/metadata"

	// check that an item with the same URL updates the endpoint in the database
	queueTmp["tlsVersion"] = "TLS 1.3"
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not store data as expected")

	storedEndpt, err = store.GetFHIREndpointUsingURL(ctx, expectedEndpt.URL)
	th.Assert(t, err == nil, err)
	th.Assert(t, storedEndpt.TLSVersion == "TLS 1.3", "The TLS Version was not updated")

	queueTmp["tlsVersion"] = "TLS 1.2" // resetting value

	// check that error adding to store throws error
	queueTmp["url"] = "https://a-new-url.com"
	queueTmp["tlsVersion"] = strings.Repeat("a", 510) // too long. causes db error

	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err != nil, "expected error adding product")

	// resetting values
	queueTmp["url"] = "http://example.com/DTSU2/metadata"
	queueTmp["tlsVersion"] = "TLS 1.2"
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
