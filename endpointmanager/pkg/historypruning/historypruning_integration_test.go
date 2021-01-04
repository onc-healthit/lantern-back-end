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
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/spf13/viper"
)

var store *postgresql.Store
var addFHIREndpointInfoHistoryStatement *sql.Stmt

var testFhirEndpointInfo = endpointmanager.FHIREndpointInfo{
	URL:           "http://example.com/DTSU2/",
	MIMETypes:     []string{"application/json+fhir"},
	TLSVersion:    "TLS 1.2",
	SMARTResponse: nil,
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

func Test_historypruning(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	setupCapabilityStatement(t, filepath.Join("../testdata", "cerner_capability_dstu2.json"))

	ctx := context.Background()

	thresholdInt := viper.GetInt("pruning_threshold")
	threshold := strconv.Itoa(thresholdInt)
	//queryInterval := strconv.Itoa(thresholdInt + (2 * viper.GetInt("capquery_qryintvl")))
	queryInterval := ""
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
		capability_statement)			
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`)
	th.Assert(t, err == nil, err)
	defer addFHIREndpointInfoHistoryStatement.Close()

	var historyURL = "http://example.com/DTSU2/"
	// reset context
	ctx = context.Background()

	// Add few days to the threshold to make sure date is older than a month
	pastDate := thresholdInt + 3*(1440)

	var Timestamp time.Time
	expectedTimeStatement, err := store.DB.Prepare(`SELECT entered_at FROM fhir_endpoints_info_history WHERE url = $1;`)
	th.Assert(t, err == nil, err)
	defer expectedTimeStatement.Close()

	var count int
	ctStatement, err := store.DB.Prepare(`SELECT count(*) FROM fhir_endpoints_info_history WHERE url = $1;`)
	th.Assert(t, err == nil, err)
	defer ctStatement.Close()

	// Clear history table in database
	clearStatement, err := store.DB.Prepare(`DELETE FROM fhir_endpoints_info_history WHERE url = $1;`)
	th.Assert(t, err == nil, err)
	defer clearStatement.Close()
	_, err = clearStatement.ExecContext(ctx, historyURL)
	th.Assert(t, err == nil, err)

	// Add three fhir endpoint info history entries with old entered at date
	oldestTime := time.Now().Add(time.Duration((-1)*pastDate) * time.Minute).Format("2006-01-02 15:04:05.000000000")
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, oldestTime)
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// HistoryPruningCheck ignores current entry and prunes old repetitive info entries, keeping the oldest entry
	HistoryPruningCheck(ctx, store, threshold, queryInterval)

	// Should be 1 entry as history pruning will remove the two newest repetitive entries and keep oldest repetitive entry
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, "Should have got 1, got "+strconv.Itoa(count))

	// Check remaining entry is the oldest entry
	err = expectedTimeStatement.QueryRow(historyURL).Scan(&Timestamp)
	th.Assert(t, err == nil, err)
	th.Assert(t, Timestamp.Format("2006-01-02 15:04:05.000000000") == oldestTime, "Expected remaining entry "+Timestamp.Format("2006-01-02 15:04:05.000000000")+" to be the oldest repeated entry "+oldestTime)

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyURL)
	th.Assert(t, err == nil, err)

	// Add two endpoint entries to info history table with current entered_at dates
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Call HistoryPruningCheck function which will call the history pruning function
	HistoryPruningCheck(ctx, store, threshold, queryInterval)

	// Info history table should have 2 entries as history pruning will not remove entries less than month old
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyURL)
	th.Assert(t, err == nil, err)

	// Add endpoint entries to info history table with old dates and non-modified capability statement date fields
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
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
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 4, "Should have got 4, got "+strconv.Itoa(count))

	// Call HistoryPruningCheck function
	HistoryPruningCheck(ctx, store, threshold, queryInterval)

	// Info history table should have only 1 entry as history pruning will remove all old entries if their capability statements only differ by date field and keep only oldest entry
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, "Should have got 1, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyURL)
	th.Assert(t, err == nil, err)

	// Add endpoint entries to info history table with old dates and non-modified capability statement date fields
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
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
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 4, "Should have got 4, got "+strconv.Itoa(count))

	// Call HistoryPruningCheck function
	HistoryPruningCheck(ctx, store, threshold, queryInterval)

	// Info history table should have 2 entries as history pruning will remove 1 entry with modified description and 1 entry without modified description, keeping the oldest of each
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyURL)
	th.Assert(t, err == nil, err)

	// Add two endpoint entries to info history table with non-modified capability statement description fields
	testFhirEndpointInfo.CapabilityStatement = originalCapStat

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Modify the description field of the capability statement
	testFhirEndpointInfo.CapabilityStatement = capStatDescription

	// Add one endpoint entries to info history table with old dates and modified capability statement description fields
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure entry was added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with non-modified capability statement description fields
	testFhirEndpointInfo.CapabilityStatement = originalCapStat

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 5, "Should have got 5, got "+strconv.Itoa(count))

	// Call HistoryPruningCheck function
	HistoryPruningCheck(ctx, store, threshold, queryInterval)

	// Info history table should have 3 entries as history pruning will remove 1 of the first two equal entries, will not remove the modified description entry in middle, and will remove 1 of the oldest non modified capability statements
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyURL)
	th.Assert(t, err == nil, err)

	// Set capability statement equal to null
	testFhirEndpointInfo.CapabilityStatement = nil

	// Add two endpoint entries to info history table with old dates and null capability statement
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Call HistoryPruningCheck function
	HistoryPruningCheck(ctx, store, threshold, queryInterval)

	// Info history table should have 1 entries as history pruning will remove the newer null capability statement entry
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, "Should have got 1, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyURL)
	th.Assert(t, err == nil, err)

	// Add two endpoint entries to info history table with old dates and null capability statement
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and non-null capability statement
	testFhirEndpointInfo.CapabilityStatement = originalCapStat

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 4, "Should have got 4, got "+strconv.Itoa(count))

	// Call HistoryPruningCheck function
	HistoryPruningCheck(ctx, store, threshold, queryInterval)

	// Info history table should have 2 entry as history pruning will remove 1 of the non null capability statment entries and 1 of the null capability statement entries
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyURL)
	th.Assert(t, err == nil, err)

	// Add two endpoint entries to info history table with old dates and null capability statement
	testFhirEndpointInfo.CapabilityStatement = nil

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Add one endpoint entries to info history table with old dates and non-null capability statement
	testFhirEndpointInfo.CapabilityStatement = originalCapStat

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure entry was added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and null capability statement
	testFhirEndpointInfo.CapabilityStatement = nil

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 5, "Should have got 5, got "+strconv.Itoa(count))

	// Call HistoryPruningCheck function
	HistoryPruningCheck(ctx, store, threshold, queryInterval)

	// Info history table should have 3 entries as history pruning will remove 1 of the first two old null entries, it will not remove the non-null entry in middle, and it will remove 1 of the older null entries more entries
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyURL)
	th.Assert(t, err == nil, err)

	// Add two endpoint entries to info history table with old dates
	testFhirEndpointInfo.CapabilityStatement = originalCapStat

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and different MIME types
	testFhirEndpointInfo.MIMETypes = []string{"application/json+fhir", "application/fhir+json"}

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 4, "Should have got 4, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and no MIME types
	testFhirEndpointInfo.MIMETypes = []string{}

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 6, "Should have got 6, got "+strconv.Itoa(count))

	// Call HistoryPruningCheck function
	HistoryPruningCheck(ctx, store, threshold, queryInterval)

	// Info history table should have 3 entries as history pruning will keep one entry for each differing mime type
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

	testFhirEndpointInfo.MIMETypes = []string{"application/json+fhir"}

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyURL)
	th.Assert(t, err == nil, err)

	// Add two endpoint entries to info history table with old dates
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and different TLS version
	testFhirEndpointInfo.TLSVersion = "TLS 1.3"

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 4, "Should have got 4, got "+strconv.Itoa(count))

	// Call HistoryPruningCheck function
	HistoryPruningCheck(ctx, store, threshold, queryInterval)

	// Info history table should have 2 entries one for each differing tls version
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	testFhirEndpointInfo.TLSVersion = "TLS 1.2"

	// Clear history table in database
	_, err = clearStatement.ExecContext(ctx, historyURL)
	th.Assert(t, err == nil, err)

	// Add two endpoint entries to info history table with old dates
	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure both entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, "Should have got 2, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and different smart response
	smartResp, _ := capabilityparser.NewSMARTResp([]byte(
		`{
			"authorization_endpoint": "https://ehr.example.com/auth/authorize",
			"token_endpoint": "https://ehr.example.com/auth/token"
		}`))
	testFhirEndpointInfo.SMARTResponse = smartResp

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 4, "Should have got 4, got "+strconv.Itoa(count))

	// Add two endpoint entries to info history table with old dates and different smart response
	smartResp2, _ := capabilityparser.NewSMARTResp([]byte(
		`{
			"authorization_endpoint": "https://ehr.differentexample.com/auth/authorize",
			"token_endpoint": "https://ehr.example.com/auth/token"
		}`))
	testFhirEndpointInfo.SMARTResponse = smartResp2

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	err = AddFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Add(time.Duration((-1)*pastDate)*time.Minute).Format("2006-01-02 15:04:05.000000000"))
	th.Assert(t, err == nil, err)

	// Ensure all entries were added to info history table correctly
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 6, "Should have got 6, got "+strconv.Itoa(count))

	// Call HistoryPruningCheck function
	HistoryPruningCheck(ctx, store, threshold, queryInterval)

	// Info history table should have 3 entries one for each differing smart response
	err = ctStatement.QueryRow(historyURL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, "Should have got 3, got "+strconv.Itoa(count))

}

// AddFHIREndpointInfoHistory adds the FHIREndpointInfoHistory to the database.
func AddFHIREndpointInfoHistory(ctx context.Context, store *postgresql.Store, e endpointmanager.FHIREndpointInfo, createdAt string) error {
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
		"U",
		createdAt,
		123,
		e.URL,
		e.TLSVersion,
		pq.Array(e.MIMETypes),
		smartResponseJSON,
		capabilityStatementJSON)
	if err != nil {
		return err
	}
	return err
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
