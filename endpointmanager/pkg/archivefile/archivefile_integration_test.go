// +build integration

package archivefile

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/spf13/viper"
)

var store *postgresql.Store
var addFHIREndpointInfoHistoryStatement *sql.Stmt
var addVendorStatement *sql.Stmt
var addFHIREndpointStatement *sql.Stmt
var getIDStatement *sql.Stmt
var ctStatement *sql.Stmt
var idCount int = 0

var testFhirEndpointInfo = endpointmanager.FHIREndpointInfo{
	URL:        "http://example.com/DTSU2/",
	MIMETypes:  []string{"application/json+fhir"},
	TLSVersion: "TLS 1.2",
}

var testFhirEndpointInfo2 = endpointmanager.FHIREndpointInfo{
	URL:        "http://example.com/DTSU2/",
	MIMETypes:  []string{"application/fhir+json"},
	TLSVersion: "TLS 1.3",
}

var testFhirEndpoint = endpointmanager.FHIREndpoint{
	URL:               "http://example.com/DTSU2/",
	OrganizationNames: []string{"Org 1"},
	ListSource:        "http://cerner.com/dstu2",
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

	addFHIREndpointInfoHistoryStatement, err = store.DB.Prepare(`
	INSERT INTO fhir_endpoints_info_history (
		operation, 
		updated_at, 
		id, 
		url,
		tls_version,
		mime_types,
		vendor_id,
		capability_statement)			
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`)
	if err != nil {
		panic(err)
	}
	defer addFHIREndpointInfoHistoryStatement.Close()

	ctStatement, err = store.DB.Prepare(`SELECT count(*) FROM fhir_endpoints_info_history WHERE url = $1;`)
	if err != nil {
		panic(err)
	}
	defer ctStatement.Close()

	addVendorStatement, err = store.DB.Prepare(`
	INSERT INTO vendors (
		name,
		id,
		developer_code,
		chpl_id)
	VALUES ($1, $2, $3, $4)`)
	if err != nil {
		panic(err)
	}
	defer addVendorStatement.Close()

	code := m.Run()

	teardown()
	os.Exit(code)
}

func Test_CreateArchive(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	var count int
	ctx := context.Background()
	setupCapabilityStatement(t, filepath.Join("../testdata", "cerner_capability_dstu2.json"))

	// populate vendors
	for _, vendor := range vendors {
		_, err = addVendorStatement.ExecContext(ctx, vendor.Name, vendor.CHPLID, vendor.DeveloperCode, vendor.CHPLID)
		th.Assert(t, err == nil, err)
	}

	// Add FHIR Endpoint
	err = store.AddFHIREndpoint(ctx, &testFhirEndpoint)
	th.Assert(t, err == nil, err)

	// Get today and tomorrow's date
	today := time.Now()
	formatTomorrow := today.Add(time.Hour * 24).Format("2006-01-02")
	formatToday := today.Format("2006-01-02")

	// Empty test, come back to this

	entries, err := CreateArchive(ctx, store, formatToday, formatTomorrow)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(entries) == 1, fmt.Sprintf("length of entries should have been 1, is instead %d", len(entries)))
	th.Assert(t, entries[0].NumberOfUpdates == 0, fmt.Sprintf("There should have been no updates, instead there were %d updates", entries[0].NumberOfUpdates))
	th.Assert(t, len(entries[0].TLSVersion) == 2, fmt.Sprintf("TLS first and last should exist, is instead %+v", entries[0].TLSVersion))
	th.Assert(t, entries[0].TLSVersion["first"] == nil, fmt.Sprint("TLS first should have been nil"))
	th.Assert(t, entries[0].TLSVersion["last"] == nil, fmt.Sprint("TLS last should have been nil"))

	// Add 1 endpoint and make sure values are correct

	err = addFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Format("2006-01-02 15:04:05.000000000"), idCount, "I", 1)
	th.Assert(t, err == nil, err)
	err = ctStatement.QueryRow(testFhirEndpointInfo.URL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, fmt.Sprintf("Should have got 1, intead got %d", count))

	entries, err = CreateArchive(ctx, store, formatToday, formatTomorrow)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(entries) == 1, fmt.Sprintf("length of entries should have been 1, is instead %d", len(entries)))
	th.Assert(t, entries[0].NumberOfUpdates == 1, fmt.Sprintf("only 1 update should have been registered, instead there were %d updates", entries[0].NumberOfUpdates))
	th.Assert(t, len(entries[0].TLSVersion) == 2, fmt.Sprintf("TLS first and last should exist, is instead %+v", entries[0].TLSVersion))
	th.Assert(t, entries[0].TLSVersion["first"] == "TLS 1.2", fmt.Sprintf("TLS first should have been TLS 1.2, is instead %s", entries[0].TLSVersion["first"]))
	th.Assert(t, entries[0].TLSVersion["last"] == nil, fmt.Sprint("TLS last should have been nil"))

	// Add 2nd endpoint (with same data) and make sure values are correct

	err = addFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Format("2006-01-02 15:04:05.000000000"), idCount, "U", 1)
	th.Assert(t, err == nil, err)
	err = ctStatement.QueryRow(testFhirEndpointInfo.URL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, fmt.Sprintf("Should have got 2, intead got %d", count))

	entries, err = CreateArchive(ctx, store, formatToday, formatTomorrow)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(entries) == 1, fmt.Sprintf("length of entries should have been 1, is instead %d", len(entries)))
	th.Assert(t, entries[0].NumberOfUpdates == 2, fmt.Sprintf("2 updates should have been registered, instead there were %d updates", entries[0].NumberOfUpdates))
	th.Assert(t, len(entries[0].TLSVersion) == 2, fmt.Sprintf("TLS first and last should exist, is instead %+v", entries[0].TLSVersion))
	th.Assert(t, entries[0].TLSVersion["first"] == "TLS 1.2", fmt.Sprintf("TLS first should have been TLS 1.2, is instead %s", entries[0].TLSVersion["first"]))
	th.Assert(t, entries[0].TLSVersion["last"] == nil, fmt.Sprintf("TLS last should have been nil since it should have been the same value, it is instead %s", entries[0].TLSVersion["last"]))
	th.Assert(t, entries[0].Operation["first"] == "I", fmt.Sprintf("First operation should have been 'I', is instead %s", entries[0].Operation["first"]))
	th.Assert(t, entries[0].Operation["last"] == "U", fmt.Sprintf("Last operation should have been 'U', is instead %s", entries[0].Operation["last"]))

	// Add 3rd endpoint (with different data) and make sure values are correct

	err = addFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo2, time.Now().Format("2006-01-02 15:04:05.000000000"), idCount, "U", 2)
	th.Assert(t, err == nil, err)
	err = ctStatement.QueryRow(testFhirEndpointInfo.URL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, fmt.Sprintf("Should have got 3, intead got %d", count))

	entries, err = CreateArchive(ctx, store, formatToday, formatTomorrow)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(entries) == 1, fmt.Sprintf("length of entries should have been 1, is instead %d", len(entries)))
	th.Assert(t, entries[0].NumberOfUpdates == 3, fmt.Sprintf("3 updates should have been registered, instead there were %d updates", entries[0].NumberOfUpdates))
	th.Assert(t, entries[0].TLSVersion["first"] == "TLS 1.2", fmt.Sprintf("TLS first should have been TLS 1.2, is instead %s", entries[0].TLSVersion["first"]))
	th.Assert(t, entries[0].TLSVersion["last"] == "TLS 1.3", fmt.Sprintf("TLS last should have been TLS 1.3, it is instead %s", entries[0].TLSVersion["last"]))
	th.Assert(t, helpers.StringArraysEqual(entries[0].MIMETypes.First, testFhirEndpointInfo.MIMETypes), fmt.Sprintf("Mime types are not correct, are instead %+v", entries[0].MIMETypes.First))
	th.Assert(t, helpers.StringArraysEqual(entries[0].MIMETypes.Last, testFhirEndpointInfo2.MIMETypes), fmt.Sprintf("Mime types are not correct, are instead %+v", entries[0].MIMETypes.Last))
	th.Assert(t, entries[0].Vendor["first"] == "Epic Systems Corporation", fmt.Sprintf("Vendor name should equal 'Epic Systems Corporation', is instead %s", entries[0].Vendor["first"]))
	th.Assert(t, entries[0].Vendor["last"] == "Cerner Corporation", fmt.Sprintf("Vendor name should equal 'Cerner Corporation', is instead %s", entries[0].Vendor["last"]))

	// Get empty info if the date range is tomorrow & the day after that

	twoDays := today.Add(time.Hour * 48).Format("2006-01-02")

	entries, err = CreateArchive(ctx, store, formatTomorrow, twoDays)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(entries) == 1, fmt.Sprintf("length of entries should have been 1, is instead %d", len(entries)))
	th.Assert(t, entries[0].NumberOfUpdates == 0, fmt.Sprintf("There should have been no updates, instead there were %d updates", entries[0].NumberOfUpdates))
	th.Assert(t, len(entries[0].TLSVersion) == 2, fmt.Sprintf("TLS first and last should exist, is instead %+v", entries[0].TLSVersion))
	th.Assert(t, entries[0].TLSVersion["first"] == nil, fmt.Sprint("TLS first should have been nil"))
	th.Assert(t, entries[0].TLSVersion["last"] == nil, fmt.Sprint("TLS last should have been nil"))
}

func Test_getHistory(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	var count int
	ctx := context.Background()
	setupCapabilityStatement(t, filepath.Join("../testdata", "cerner_capability_dstu2.json"))

	// populate vendors
	for _, vendor := range vendors {
		_, err = addVendorStatement.ExecContext(ctx, vendor.Name, vendor.CHPLID, vendor.DeveloperCode, vendor.CHPLID)
		th.Assert(t, err == nil, err)
	}

	// Add FHIR Endpoint
	err = store.AddFHIREndpoint(ctx, &testFhirEndpoint)
	th.Assert(t, err == nil, err)

	// Get today and tomorrow's date
	today := time.Now()
	formatTomorrow := today.Add(time.Hour * 24).Format("2006-01-02")
	formatToday := today.Format("2006-01-02")

	err = addFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Format("2006-01-02 15:04:05.000000000"), idCount, "U", 1)
	th.Assert(t, err == nil, err)
	err = ctStatement.QueryRow(testFhirEndpointInfo.URL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, fmt.Sprintf("Should have got 1, intead got %d", count))

	// Base Case

	resultCh := make(chan Result)
	jobArgs := make(map[string]interface{})
	jobArgs["historyArgs"] = historyArgs{
		fhirURL:   "http://example.com/DTSU2/",
		dateStart: formatToday,
		dateEnd:   formatTomorrow,
		store:     store,
		result:    resultCh,
	}

	go getHistory(ctx, &jobArgs)

	for res := range resultCh {
		th.Assert(t, res.URL == "http://example.com/DTSU2/", fmt.Sprintf("Expected URL to equal 'http://example.com/DTSU2/'. Is actually '%s'.", res.URL))
		th.Assert(t, res.Summary.NumberOfUpdates == 1, fmt.Sprintf("1 update should have been registered, instead there were %d updates", res.Summary.NumberOfUpdates))
		th.Assert(t, res.Summary.TLSVersion["first"] == "TLS 1.2", fmt.Sprintf("TLS first should have been TLS 1.2, is instead %s", res.Summary.TLSVersion["first"]))
		th.Assert(t, res.Summary.TLSVersion["last"] == nil, fmt.Sprintf("TLS last should have been nil, it is instead %s", res.Summary.TLSVersion["last"]))
		close(resultCh)
	}

	// If the args are not properly formatted

	jobArgs2 := make(map[string]interface{})
	jobArgs2["historyArgs"] = map[string]interface{}{
		"nonsense": 1,
	}

	err = getHistory(ctx, &jobArgs2)
	th.Assert(t, err != nil, fmt.Sprint("Malformed arguments should have thrown error."))

	// If the URL does not exist, return default data

	resultCh3 := make(chan Result)
	jobArgs3 := make(map[string]interface{})
	jobArgs3["historyArgs"] = historyArgs{
		fhirURL:   "thisurldoesntexist.com",
		dateStart: formatToday,
		dateEnd:   formatTomorrow,
		store:     store,
		result:    resultCh3,
	}

	go getHistory(ctx, &jobArgs3)
	for res := range resultCh3 {
		th.Assert(t, res.Summary.NumberOfUpdates == 0, fmt.Sprintf("Expected 0 entries in history table. Actually had %d entries.", res.Summary.NumberOfUpdates))
		th.Assert(t, res.URL == "thisurldoesntexist.com", fmt.Sprintf("Expected URL to equal 'thisurldoesntexist.com'. Is actually '%s'.", res.URL))
		th.Assert(t, res.Summary.TLSVersion["first"] == nil, fmt.Sprint("TLS first should have been nil"))
		th.Assert(t, res.Summary.TLSVersion["last"] == nil, fmt.Sprint("TLS last should have been nil"))
		close(resultCh3)
	}
}

func setupCapabilityStatement(t *testing.T, path string) {
	// capability statement
	csJSON, err := ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)
	testFhirEndpointInfo.CapabilityStatement = cs
}

// addFHIREndpointInfoHistory adds the FHIREndpointInfoHistory to the database.
func addFHIREndpointInfoHistory(ctx context.Context,
	store *postgresql.Store,
	e endpointmanager.FHIREndpointInfo,
	updatedAt string,
	id int,
	operation string,
	vendorID int) error {

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
		operation,
		updatedAt,
		id,
		e.URL,
		e.TLSVersion,
		pq.Array(e.MIMETypes),
		vendorID,
		capabilityStatementJSON)
	if err != nil {
		return err
	}

	idCount++

	return err
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
