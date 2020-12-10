// +build integration

package capabilityhandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var store *postgresql.Store

var addFHIREndpointInfoHistoryStatement *sql.Stmt

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
	expectedEndpt = testFhirEndpointInfo
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
	th.Assert(t, storedEndpt.HTTPResponse == 404, "The http response was not updated")

	// check that availability is updated
	err = store.DB.QueryRow(query_str, testFhirEndpoint1.URL).Scan(&http_200_ct, &http_all_ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, http_all_ct == 2, "http all count should have been incremented to 2")
	th.Assert(t, storedEndpt.Availability == 0.5, "endpoint availability should have been updated to 0.5")

	queueTmp["tlsVersion"] = "TLS 1.2" // resetting value
	queueTmp["httpResponse"] = 200

	// check that error adding to store throws error
	queueTmp["url"] = "https://a-new-url.com"
	queueTmp["tlsVersion"] = strings.Repeat("a", 510) // too long. causes db error

	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err != nil, "expected error adding product")

	addFHIREndpointInfoHistoryStatement, err = store.DB.Prepare(`
	INSERT INTO fhir_endpoints_info_history (
		operation, 
		entered_at, 
		id, 
		url, 
		capability_statement)			
	VALUES ($1, $2, $3, $4, $5);`)
	th.Assert(t, err == nil, err)
	defer addFHIREndpointInfoHistoryStatement.Close()

	// resetting values
	queueTmp["url"] = "http://example.com/DTSU2/"
	queueTmp["tlsVersion"] = "TLS 1.2"

	historyUrl := "http://example.com/DTSU2/"
	// reset context
	ctx = context.Background()
	args["ctx"] = ctx

	// Add few days to the threshold to make sure date is older than a month
	threshold := viper.GetInt("pruning_threshold") + 3*(1440)

	currentTime := time.Now()
	pastTime := currentTime.Add(time.Duration((-1)*threshold) * time.Minute)

	// Clear history table in database
	clearStatement, err := store.DB.Prepare(`DELETE FROM fhir_endpoints_info_history WHERE url = $1;`)
	th.Assert(t, err == nil, err)
	defer clearStatement.Close()
	_, err = clearStatement.ExecContext(ctx, historyUrl)
	th.Assert(t, err == nil, err)

	// Add fhir endpoint info history entry with old entered at date
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, pastTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	var count int
	ctStatement, err := store.DB.Prepare(`SELECT count(*) FROM fhir_endpoints_info_history WHERE url = $1;`)
	th.Assert(t, err == nil, err)
	defer ctStatement.Close()

	// Ensure entry was added to info history table correctly
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, "Should have got 1, got "+strconv.Itoa(count))

	// Add a second old info history entry
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, pastTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Save message in DB stores a new entry in endpoint info history table and prunes old entries
	queueTmp = testQueueMsg
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	// Should only be one entry as history pruning will remove the two old entries
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, "Should have got 1, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyUrl)
	th.Assert(t, err == nil, err)

	// Add two endpoint entries to info history table with current entered_at dates
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, currentTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, currentTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Call saveMsgInDB function which will call the history pruning function
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	// Info history table should have 3 entries as history pruning will not remove entries less than month old
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyUrl)
	th.Assert(t, err == nil, err)

	// Modify the date field of the capability statement
	originalCapStat := testFhirEndpointInfo.CapabilityStatement
	cs := testFhirEndpointInfo.CapabilityStatement
	var csInt map[string]interface{}
	csJSON, err := cs.GetJSON()
	th.Assert(t, err == nil, err)
	err = json.Unmarshal(csJSON, &csInt)
	th.Assert(t, err == nil, err)

	csInt["date"] = "2010-01-03 15:04:05"
	capStatDate, err := capabilityparser.NewCapabilityStatementFromInterface(csInt)
	th.Assert(t, err == nil, err)
	testFhirEndpointInfo.CapabilityStatement = capStatDate

	// Add two endpoint entries to info history table with old dates and modified capability statement date fields
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, pastTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, pastTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Call saveMsgInDB function which will call the history pruning function
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	// Info history table should have only 1 entry as history pruning will remove old entries if their capability statements only differ by date field
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, "Should have got 1, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyUrl)
	th.Assert(t, err == nil, err)

	// Add two endpoint entries to info history table with current dates and modified capability statement date fields
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, currentTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, currentTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Call saveMsgInDB function which will call the history pruning function
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	// Info history table should have 3 entries as history pruning will not remove entries less than month old
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyUrl)
	th.Assert(t, err == nil, err)

	// Modify the description field of the capability statement
	testFhirEndpointInfo.CapabilityStatement = originalCapStat
	cs = testFhirEndpointInfo.CapabilityStatement
	csJSON, err = cs.GetJSON()
	th.Assert(t, err == nil, err)
	err = json.Unmarshal(csJSON, &csInt)
	th.Assert(t, err == nil, err)

	csInt["description"] = "This is a new description for the capability statement"
	capStatDescription, err := capabilityparser.NewCapabilityStatementFromInterface(csInt)
	th.Assert(t, err == nil, err)
	testFhirEndpointInfo.CapabilityStatement = capStatDescription

	// Add two endpoint entries to info history table with old dates and modified capability statement description fields
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, pastTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, pastTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Call saveMsgInDB function which will call the history pruning function
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	// Info history table should have 3 entries as history pruning will not remove old entries if their capability statements differ by field other than date field
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyUrl)
	th.Assert(t, err == nil, err)

	// Set capability statement equal to null
	testFhirEndpointInfo.CapabilityStatement = nil

	// Add two endpoint entries to info history table with old dates and null capability statement
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, pastTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, pastTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Call saveMsgInDB function which will call the history pruning function
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	// Info history table should have 3 entries as history pruning will not remove old entries if their capability statements are null but new capability statement is not null
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyUrl)
	th.Assert(t, err == nil, err)

	// Set msg queue capability statement equal to null
	queueTmp["capabilityStatement"] = nil
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)

	// Add two endpoint entries to info history table with old dates and null capability statement
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, pastTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, pastTime.Format("2006-01-02 15:04:05"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Call saveMsgInDB function which will call the history pruning function
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)

	// Info history table should have 1 entry as history pruning will remove old entries if both their capability statements and new capability statement null
	err = ctStatement.QueryRow(historyUrl).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, "Should have got 1, got "+strconv.Itoa(count))

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

// AddFHIREndpointInfoHistory adds the FHIREndpointInfoHistory to the database.
func AddFHIREndpointInfoHistory(ctx context.Context, store *postgresql.Store, e endpointmanager.FHIREndpointInfo, created_at string) error {
	var err error
	var capabilityStatementJSON []byte

	if e.CapabilityStatement != nil {
		capabilityStatementJSON, err = e.CapabilityStatement.GetJSON()
		if err != nil {
			return err
		}
	} else {
		capabilityStatementJSON = []byte("null")
	}
	_, err = addFHIREndpointInfoHistoryStatement.ExecContext(ctx,
		"U",
		created_at,
		123,
		e.URL,
		capabilityStatementJSON)
	if err != nil {
		return err
	}
	return err
}
