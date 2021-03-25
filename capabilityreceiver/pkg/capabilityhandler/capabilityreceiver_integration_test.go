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
	&endpointmanager.Vendor{
		Name:          "Epic Systems Corporation",
		DeveloperCode: "A",
		CHPLID:        1,
	},
	&endpointmanager.Vendor{
		Name:          "Cerner Corporation",
		DeveloperCode: "B",
		CHPLID:        2,
	},
	&endpointmanager.Vendor{
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

func Test_saveMsgInDB(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	setupCapabilityStatement(t, filepath.Join("../../testdata", "cerner_capability_dstu2.json"))

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM fhir_endpoints_info;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()

	args := make(map[string]interface{})
	args["store"] = store
	args["chplMatchFile"] = "../../testdata/test_chpl_product_mapping.json"

	ctx := context.Background()

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
	args["ctx"] = testCtx
	cancel()
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, errors.Cause(err) == context.Canceled, fmt.Sprintf("should have errored out with root cause that the context was canceled, instead was %s", err))

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 0, "should not have stored data")

	// reset context
	args["ctx"] = context.Background()

	// check that new item is stored
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, errors.Wrap(err, "error"))

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")

	storedEndpt, err := store.GetFHIREndpointInfoUsingURL(ctx, testFhirEndpoint1.URL)
	storedEndpt.Validation.Results = []endpointmanager.Rule{storedEndpt.Validation.Results[0]}
	th.Assert(t, err == nil, err)
	th.Assert(t, expectedEndpt.Equal(storedEndpt), "stored data does not equal expected store data")

	// check that endpoint availability was updated
	var http_200_ct int
	var http_all_ct int
	var endpt_availability_ct int
	query_str := "SELECT http_200_count, http_all_count from fhir_endpoints_availability WHERE url=$1;"
	ct_availability_str := "SELECT COUNT(*) from fhir_endpoints_availability;"

	err = store.DB.QueryRow(ct_availability_str).Scan(&endpt_availability_ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, endpt_availability_ct == 1, "endpoint availability should have 1 endpoint")
	err = store.DB.QueryRow(query_str, testFhirEndpoint1.URL).Scan(&http_200_ct, &http_all_ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, http_all_ct == 1, "endpoint should have http return count of 1")
	th.Assert(t, http_200_ct == 1, "endpoint should have http 200 return count of 1")

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

	storedEndpt, err = store.GetFHIREndpointInfoUsingURL(ctx, testFhirEndpoint2.URL)
	storedEndpt.Validation.Results = []endpointmanager.Rule{storedEndpt.Validation.Results[0]}
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

	storedEndpt, err = store.GetFHIREndpointInfoUsingURL(ctx, testFhirEndpoint1.URL)
	th.Assert(t, err == nil, err)
	th.Assert(t, storedEndpt.TLSVersion == "TLS 1.3", "The TLS Version was not updated")
	th.Assert(t, storedEndpt.Metadata.HTTPResponse == 404, "The http response was not updated")

	// check that availability is updated
	err = store.DB.QueryRow(query_str, testFhirEndpoint1.URL).Scan(&http_200_ct, &http_all_ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, http_all_ct == 2, "http all count should have been incremented to 2, was %d")
	th.Assert(t, storedEndpt.Metadata.Availability == 0.5, "endpoint availability should have been updated to 0.5")

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

	storedEndpt, err = store.GetFHIREndpointInfoUsingURL(ctx, testFhirEndpoint1.URL)
	th.Assert(t, err == nil, err)

	store.DB.QueryRow(historySQLStatement, storedEndpt.URL).Scan(&updatedAt)
	oldUpdateAt := updatedAt
	oldMetadataID := storedEndpt.Metadata.ID
	oldMetadataUpdatedAt := storedEndpt.Metadata.UpdatedAt

	// Try to update with exact same values
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not store data as expected")
	storedEndpt, err = store.GetFHIREndpointInfoUsingURL(ctx, testFhirEndpoint1.URL)
	th.Assert(t, err == nil, err)
	th.Assert(t, storedEndpt.Metadata.ID != oldMetadataID, "The selective update should have still updated the old endpoint info metadata id")
	th.Assert(t, !storedEndpt.Metadata.UpdatedAt.Equal(oldMetadataUpdatedAt), "The selective update should have still updated the old endpoint metadata updated at time")

	store.DB.QueryRow(historySQLStatement, storedEndpt.URL).Scan(&updatedAt)
	th.Assert(t, updatedAt.Equal(oldUpdateAt), "The selective update should not have updated the old endpoint updated at time in the history table")

	oldMetadataID = storedEndpt.Metadata.ID
	oldMetadataUpdatedAt = storedEndpt.Metadata.UpdatedAt

	// Try to update with exact same values besides metadata
	queueTmp["responseTime"] = 0.3456
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not store data as expected")

	storedEndpt, err = store.GetFHIREndpointInfoUsingURL(ctx, testFhirEndpoint1.URL)
	th.Assert(t, err == nil, err)
	th.Assert(t, storedEndpt.Metadata.ID != oldMetadataID, "The selective update should have still updated the old endpoint info metadata id")
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
