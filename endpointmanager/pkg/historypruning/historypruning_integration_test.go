// +build integration

package historypruning

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/smartparser"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/spf13/viper"
)

var store *postgresql.Store
var addFHIREndpointInfoHistoryStatement *sql.Stmt
var getIDStatement *sql.Stmt
var idCount int = 0

var testFhirEndpointInfo = endpointmanager.FHIREndpointInfo{
	URL:           "http://example.com/DTSU2/",
	MIMETypes:     []string{"application/json+fhir"},
	TLSVersion:    "TLS 1.2",
	SMARTResponse: nil,
	RequestedFhirVersion: "1.0.2",
}

var testFhirEndpointInfo2 = endpointmanager.FHIREndpointInfo{
	URL:           "http://otherexample.com/DTSU2/",
	MIMETypes:     []string{"application/json+fhir"},
	TLSVersion:    "TLS 1.2",
	SMARTResponse: nil,
	RequestedFhirVersion: "1.0.2",
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

func Test_PruneInfoHistory(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	setupCapabilityStatement(t, filepath.Join("../testdata", "cerner_capability_dstu2.json"))

	ctx := context.Background()

	var err error

	addFHIREndpointInfoHistoryStatement, err = store.DB.Prepare(`
	INSERT INTO fhir_endpoints_info_history (
		operation, 
		entered_at, 
		id, 
		url,
		tls_version,
		mime_types,
		smart_response, 
		capability_statement,
		requested_fhir_version)			
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`)
	th.Assert(t, err == nil, err)
	defer addFHIREndpointInfoHistoryStatement.Close()

	var testEndpointURL = testFhirEndpointInfo.URL

	// Add few days to the threshold to make sure date is older than a month
	threshold := viper.GetInt("pruning_threshold")
	pastDate := threshold + 3*(1440)
	var idExpectedArr []int
	var checkCorrectness bool

	var idActual int
	getIDStatement, err = store.DB.Prepare(`SELECT id FROM fhir_endpoints_info_history WHERE url = $1 ORDER BY entered_at ASC;`)
	th.Assert(t, err == nil, err)
	defer getIDStatement.Close()

	var count int
	ctStatement, err := store.DB.Prepare(`SELECT count(*) FROM fhir_endpoints_info_history WHERE url = $1;`)
	th.Assert(t, err == nil, err)
	defer ctStatement.Close()

	// Clear history table in database
	clearStatement, err := store.DB.Prepare(`DELETE FROM fhir_endpoints_info_history WHERE url = $1;`)
	th.Assert(t, err == nil, err)
	defer clearStatement.Close()
	_, err = clearStatement.ExecContext(ctx, testEndpointURL)
	th.Assert(t, err == nil, err)

	// Add two endpoint entries to info history table with current entered_at dates
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Call PruneInfoHistory function which will call the history pruning function
	PruneInfoHistory(ctx, store, false)

	// Info history table should have 2 entries as history pruning will not remove entries less than month old
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, testEndpointURL)
	th.Assert(t, err == nil, err)
	idExpectedArr = nil
	var idExpectedArr2 []int

	// Add four fhir endpoint info history entries with old entered at date, first and second to last I operations
	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "I")
	th.Assert(t, err == nil, err)

	idExpectedArr2 = append(idExpectedArr2, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo2, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "I")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo2, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo2, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	err = ctStatement.QueryRow(testFhirEndpointInfo2.URL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// PruneInfoHistory ignores current entry and prunes old repetitive info entries, keeping the oldest entry
	PruneInfoHistory(ctx, store, false)

	// Should be 2 entries as history pruning will not remove the I operation entries but will remove each of their duplicates
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, "Should have got 1, got "+strconv.Itoa(count))

	err = ctStatement.QueryRow(testFhirEndpointInfo2.URL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, "Should have got 1, got "+strconv.Itoa(count))

	// Ensure correct entries were left in the database
	checkCorrectness, err = checkCorrect(idExpectedArr, testEndpointURL)
	th.Assert(t, err == nil, err)
	th.Assert(t, checkCorrectness, "Unexpected entries kept in database")

	checkCorrectness, err = checkCorrect(idExpectedArr2, testFhirEndpointInfo2.URL)
	th.Assert(t, err == nil, err)
	th.Assert(t, checkCorrectness, "Unexpected entries kept in database")

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, testEndpointURL)
	th.Assert(t, err == nil, err)
	idExpectedArr = nil

	// Add three fhir endpoint info history entries with old entered at date
	expectedID := idCount
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// PruneInfoHistory ignores current entry and prunes old repetitive info entries, keeping the oldest entry
	PruneInfoHistory(ctx, store, false)

	// Should be 1 entry as history pruning will remove the two newest repetitive entries and keep oldest repetitive entry
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, "Should have got 1, got "+strconv.Itoa(count))

	// Check remaining entry is the oldest entry
	err = getIDStatement.QueryRow(testEndpointURL).Scan(&idActual)
	th.Assert(t, err == nil, err)
	th.Assert(t, expectedID == idActual, "Expected remaining entry to have id "+strconv.Itoa(expectedID)+" but instead it was "+strconv.Itoa(idActual))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, testEndpointURL)
	th.Assert(t, err == nil, err)

	// Add endpoint entries to info history table with old dates and non-modified capability statement date fields
	expectedID = idCount
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

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
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 4, "Should have got 4, got "+strconv.Itoa(count))

	// Call PruneInfoHistory function
	PruneInfoHistory(ctx, store, false)

	// Info history table should have only 1 entry as history pruning will remove all old entries if their capability statements only differ by date field and keep only oldest entry
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, "Should have got 1, got "+strconv.Itoa(count))

	// Check remaining entry is the oldest entry
	err = getIDStatement.QueryRow(testEndpointURL).Scan(&idActual)
	th.Assert(t, err == nil, err)
	th.Assert(t, expectedID == idActual, "Expected remaining entry to have id "+strconv.Itoa(expectedID)+" but instead it was "+strconv.Itoa(idActual))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, testEndpointURL)
	th.Assert(t, err == nil, err)
	idExpectedArr = nil

	// Add endpoint entries to info history table with old dates and non-modified capability statement date fields
	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

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
	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 4, "Should have got 4, got "+strconv.Itoa(count))

	// Call PruneInfoHistory function
	PruneInfoHistory(ctx, store, false)

	// Info history table should have 2 entries as history pruning will remove 1 entry with modified description and 1 entry without modified description, keeping the oldest of each
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Ensure correct entries were left in the database
	checkCorrectness, err = checkCorrect(idExpectedArr, testEndpointURL)
	th.Assert(t, err == nil, err)
	th.Assert(t, checkCorrectness, "Unexpected entries kept in database")

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, testEndpointURL)
	th.Assert(t, err == nil, err)
	idExpectedArr = nil

	// Add two endpoint entries to info history table with non-modified capability statement description fields
	testFhirEndpointInfo.CapabilityStatement = originalCapStat

	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Modify the description field of the capability statement
	testFhirEndpointInfo.CapabilityStatement = capStatDescription

	// Add one endpoint entries to info history table with old dates and modified capability statement description fields
	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure entry was added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with non-modified capability statement description fields
	testFhirEndpointInfo.CapabilityStatement = originalCapStat

	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 5, "Should have got 5, got "+strconv.Itoa(count))

	// Call PruneInfoHistory function
	PruneInfoHistory(ctx, store, false)

	// Info history table should have 3 entries as history pruning will remove 1 of the first two equal entries, will not remove the modified description entry in middle, and will remove 1 of the oldest non modified capability statements
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Ensure correct entries were left in the database
	checkCorrectness, err = checkCorrect(idExpectedArr, testEndpointURL)
	th.Assert(t, err == nil, err)
	th.Assert(t, checkCorrectness, "Unexpected entries kept in database")

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, testEndpointURL)
	th.Assert(t, err == nil, err)

	// Set capability statement equal to null
	testFhirEndpointInfo.CapabilityStatement = nil

	// Add two endpoint entries to info history table with old dates and null capability statement
	expectedID = idCount
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Call PruneInfoHistory function
	PruneInfoHistory(ctx, store, false)

	// Info history table should have 1 entries as history pruning will remove the newer null capability statement entry
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, "Should have got 1, got "+strconv.Itoa(count))

	// Check remaining entry is the oldest entry
	err = getIDStatement.QueryRow(testEndpointURL).Scan(&idActual)
	th.Assert(t, err == nil, err)
	th.Assert(t, expectedID == idActual, "Expected remaining entry to have id "+strconv.Itoa(expectedID)+" but instead it was "+strconv.Itoa(idActual))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, testEndpointURL)
	th.Assert(t, err == nil, err)
	idExpectedArr = nil

	// Add two endpoint entries to info history table with old dates and null capability statement
	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and non-null capability statement
	testFhirEndpointInfo.CapabilityStatement = originalCapStat

	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 4, "Should have got 4, got "+strconv.Itoa(count))

	// Call PruneInfoHistory function
	PruneInfoHistory(ctx, store, false)

	// Info history table should have 2 entry as history pruning will remove 1 of the non null capability statment entries and 1 of the null capability statement entries
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Ensure correct entries were left in the database
	checkCorrectness, err = checkCorrect(idExpectedArr, testEndpointURL)
	th.Assert(t, err == nil, err)
	th.Assert(t, checkCorrectness, "Unexpected entries kept in database")

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, testEndpointURL)
	th.Assert(t, err == nil, err)
	idExpectedArr = nil

	// Add two endpoint entries to info history table with old dates and null capability statement
	testFhirEndpointInfo.CapabilityStatement = nil

	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Add one endpoint entries to info history table with old dates and non-null capability statement
	testFhirEndpointInfo.CapabilityStatement = originalCapStat

	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure entry was added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and null capability statement
	testFhirEndpointInfo.CapabilityStatement = nil

	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 5, "Should have got 5, got "+strconv.Itoa(count))

	// Call PruneInfoHistory function
	PruneInfoHistory(ctx, store, false)

	// Info history table should have 3 entries as history pruning will remove 1 of the first two old null entries, it will not remove the non-null entry in middle, and it will remove 1 of the older null entries more entries
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Ensure correct entries were left in the database
	checkCorrectness, err = checkCorrect(idExpectedArr, testEndpointURL)
	th.Assert(t, err == nil, err)
	th.Assert(t, checkCorrectness, "Unexpected entries kept in database")

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, testEndpointURL)
	th.Assert(t, err == nil, err)
	idExpectedArr = nil

	// Add two endpoint entries to info history table with old dates
	testFhirEndpointInfo.CapabilityStatement = originalCapStat

	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and different MIME types
	testFhirEndpointInfo.MIMETypes = []string{"application/json+fhir", "application/fhir+json"}

	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 4, "Should have got 4, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and no MIME types
	testFhirEndpointInfo.MIMETypes = []string{}

	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 6, "Should have got 6, got "+strconv.Itoa(count))

	// Call PruneInfoHistory function
	PruneInfoHistory(ctx, store, false)

	// Info history table should have 3 entries as history pruning will keep one entry for each differing mime type
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Ensure correct entries were left in the database
	checkCorrectness, err = checkCorrect(idExpectedArr, testEndpointURL)
	th.Assert(t, err == nil, err)
	th.Assert(t, checkCorrectness, "Unexpected entries kept in database")

	testFhirEndpointInfo.MIMETypes = []string{"application/json+fhir"}

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, testEndpointURL)
	th.Assert(t, err == nil, err)
	idExpectedArr = nil

	// Add two endpoint entries to info history table with old dates
	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and different TLS version
	testFhirEndpointInfo.TLSVersion = "TLS 1.3"

	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 4, "Should have got 4, got "+strconv.Itoa(count))

	// Call PruneInfoHistory function
	PruneInfoHistory(ctx, store, false)

	// Info history table should have 2 entries one for each differing tls version
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Ensure correct entries were left in the database
	checkCorrectness, err = checkCorrect(idExpectedArr, testEndpointURL)
	th.Assert(t, err == nil, err)
	th.Assert(t, checkCorrectness, "Unexpected entries kept in database")

	testFhirEndpointInfo.TLSVersion = "TLS 1.2"

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, testEndpointURL)
	th.Assert(t, err == nil, err)
	idExpectedArr = nil

	// Add two endpoint entries to info history table with old dates
	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and different smart response
	smartResp, _ := smartparser.NewSMARTResp([]byte(
		`{
			"authorization_endpoint": "https://ehr.example.com/auth/authorize",
			"token_endpoint": "https://ehr.example.com/auth/token"
		}`))
	testFhirEndpointInfo.SMARTResponse = smartResp

	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 4, "Should have got 4, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and different smart response
	smartResp2, _ := smartparser.NewSMARTResp([]byte(
		`{
			"authorization_endpoint": "https://ehr.differentexample.com/auth/authorize",
			"token_endpoint": "https://ehr.example.com/auth/token"
		}`))
	testFhirEndpointInfo.SMARTResponse = smartResp2

	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 6, "Should have got 6, got "+strconv.Itoa(count))

	// Call PruneInfoHistory function
	PruneInfoHistory(ctx, store, false)

	// Info history table should have 3 entries one for each differing smart response
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Ensure correct entries were left in the database
	checkCorrectness, err = checkCorrect(idExpectedArr, testEndpointURL)
	th.Assert(t, err == nil, err)
	th.Assert(t, checkCorrectness, "Unexpected entries kept in database")

	testFhirEndpointInfo.SMARTResponse = nil

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, testEndpointURL)
	th.Assert(t, err == nil, err)
	idExpectedArr = nil
	
	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Change requested fhir version for same endpoint
	testFhirEndpointInfo.RequestedFhirVersion = "4.0.0"

	idExpectedArr = append(idExpectedArr, idCount)
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"), idCount, "U")
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Call PruneInfoHistory function
	PruneInfoHistory(ctx, store, false)

	// Info history table should have 2 entries as history pruning will keep both entries for an endpoint if their requested version differs
	err = ctStatement.QueryRow(testEndpointURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Ensure correct entries were left in the database
	checkCorrectness, err = checkCorrect(idExpectedArr, testEndpointURL)
	th.Assert(t, err == nil, err)
	th.Assert(t, checkCorrectness, "Unexpected entries kept in database")

	testFhirEndpointInfo.RequestedFhirVersion = "1.0.2"
	
}

// AddFHIREndpointInfoHistory adds the FHIREndpointInfoHistory to the database.
func AddFHIREndpointInfoHistory(ctx context.Context, store *postgresql.Store, e endpointmanager.FHIREndpointInfo, createdAt string, id int, operation string) error {
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

	var smartResponseJSON []byte
	if e.SMARTResponse != nil {
		smartResponseJSON, err = e.SMARTResponse.GetJSON()
		if err != nil {
			return err
		}
	} else {
		smartResponseJSON = []byte("null")
	}

	_, err = addFHIREndpointInfoHistoryStatement.ExecContext(ctx,
		operation,
		createdAt,
		id,
		e.URL,
		e.TLSVersion,
		pq.Array(e.MIMETypes),
		smartResponseJSON,
		capabilityStatementJSON,
		e.RequestedFhirVersion)
	if err != nil {
		return err
	}

	idCount++

	return err
}

func checkCorrect(idArr []int, testEndpointURL string) (bool, error) {
	rows, err := getIDStatement.Query(testEndpointURL)

	if err != nil {
		return false, err
	}

	indexCount := 0
	var idActual int

	for rows.Next() {
		err = rows.Scan(&idActual)
		if err != nil {
			return false, err
		}

		if idArr[indexCount] != idActual {
			return false, nil
		}

		indexCount++
	}

	if indexCount < len(idArr) {
		return false, nil
	}

	return true, nil
}

func setupCapabilityStatement(t *testing.T, path string) {
	// capability statement
	csJSON, err := ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)
	testFhirEndpointInfo.CapabilityStatement = cs
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
