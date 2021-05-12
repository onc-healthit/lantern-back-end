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
var workerDur int
var numWorkers int

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

var testMetadata = endpointmanager.FHIREndpointMetadata{
	URL:               "http://example.com/DTSU2/",
	HTTPResponse:      200,
	Errors:            "Smart Response Failed",
	ResponseTime:      0.8,
	SMARTHTTPResponse: 400,
}

var testMetadata2 = endpointmanager.FHIREndpointMetadata{
	URL:               "http://example.com/DTSU2/",
	HTTPResponse:      200,
	Errors:            "Smart Response Failed",
	ResponseTime:      1.0,
	SMARTHTTPResponse: 0,
}

var vendors []*endpointmanager.Vendor = []*endpointmanager.Vendor{
	{
		Name:          "Epic Systems Corporation",
		DeveloperCode: "A",
		CHPLID:        1,
	},
	{
		Name:          "Cerner Corporation",
		DeveloperCode: "B",
		CHPLID:        2,
	},
	{
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

	numWorkers = viper.GetInt("export_numworkers")
	workerDur = viper.GetInt("export_duration")

	addFHIREndpointInfoHistoryStatement, err = store.DB.Prepare(`
	INSERT INTO fhir_endpoints_info_history (
		operation, 
		updated_at, 
		id, 
		url,
		tls_version,
		mime_types,
		vendor_id,
		capability_statement,
		capability_fhir_version)			
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`)
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

	entries, err := CreateArchive(ctx, store, formatToday, formatTomorrow, numWorkers, workerDur)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(entries) == 1, fmt.Sprintf("length of entries should have been 1, is instead %d", len(entries)))
	th.Assert(t, entries[0].NumberOfUpdates == 0, fmt.Sprintf("There should have been no updates, instead there were %d updates", entries[0].NumberOfUpdates))
	th.Assert(t, len(entries[0].TLSVersion) == 2, fmt.Sprintf("TLS first and last should exist, is instead %+v", entries[0].TLSVersion))
	th.Assert(t, entries[0].TLSVersion["first"] == nil, fmt.Sprint("TLS first should have been nil"))
	th.Assert(t, entries[0].TLSVersion["last"] == nil, fmt.Sprint("TLS last should have been nil"))
	th.Assert(t, entries[0].HTTPResponse == nil, fmt.Sprintf("HTTP Response should be nil, is instead %+v", entries[0].HTTPResponse))

	// Add Metadata for Endpoint
	_, err = store.AddFHIREndpointMetadata(ctx, &testMetadata)
	th.Assert(t, err == nil, err)

	// Metadata should exist without impacting the history fields

	entries, err = CreateArchive(ctx, store, formatToday, formatTomorrow, numWorkers, workerDur)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(entries) == 1, fmt.Sprintf("length of entries should have been 1, is instead %d", len(entries)))
	th.Assert(t, entries[0].NumberOfUpdates == 0, fmt.Sprintf("There should have been no updates, instead there were %d updates", entries[0].NumberOfUpdates))
	th.Assert(t, len(entries[0].HTTPResponse) == 1, fmt.Sprintf("HTTP Response length should be 1, is instead %d", len(entries[0].HTTPResponse)))

	// Add 1 endpoint and make sure values are correct

	err = addFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Format("2006-01-02 15:04:05.000000000"), idCount, "I", 1)
	th.Assert(t, err == nil, err)
	err = ctStatement.QueryRow(testFhirEndpointInfo.URL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, fmt.Sprintf("Should have got 1, intead got %d", count))

	entries, err = CreateArchive(ctx, store, formatToday, formatTomorrow, numWorkers, workerDur)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(entries) == 1, fmt.Sprintf("length of entries should have been 1, is instead %d", len(entries)))
	th.Assert(t, entries[0].NumberOfUpdates == 1, fmt.Sprintf("only 1 update should have been registered, instead there were %d updates", entries[0].NumberOfUpdates))
	th.Assert(t, len(entries[0].TLSVersion) == 2, fmt.Sprintf("TLS first and last should exist, is instead %+v", entries[0].TLSVersion))
	th.Assert(t, entries[0].TLSVersion["first"] == "TLS 1.2", fmt.Sprintf("TLS first should have been TLS 1.2, is instead %s", entries[0].TLSVersion["first"]))
	th.Assert(t, entries[0].TLSVersion["last"] == nil, fmt.Sprint("TLS last should have been nil"))
	th.Assert(t, len(entries[0].SmartHTTPResponse) == 1, fmt.Sprintf("SMART HTTP Response length should be 1, is instead %d", len(entries[0].SmartHTTPResponse)))
	th.Assert(t, entries[0].SmartHTTPResponse[0].ResponseCode == 400, fmt.Sprintf("SMART HTTP Response Code should be 400, is instead %d", entries[0].SmartHTTPResponse[0].ResponseCode))
	th.Assert(t, entries[0].SmartHTTPResponse[0].ResponseCount == 1, fmt.Sprintf("SMART HTTP Response Count should be 1, is instead %d", entries[0].SmartHTTPResponse[0].ResponseCount))

	// Add 2nd endpoint (with same data) and make sure values are correct

	err = addFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Format("2006-01-02 15:04:05.000000000"), idCount, "U", 1)
	th.Assert(t, err == nil, err)
	err = ctStatement.QueryRow(testFhirEndpointInfo.URL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, fmt.Sprintf("Should have got 2, intead got %d", count))

	entries, err = CreateArchive(ctx, store, formatToday, formatTomorrow, numWorkers, workerDur)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(entries) == 1, fmt.Sprintf("length of entries should have been 1, is instead %d", len(entries)))
	th.Assert(t, entries[0].NumberOfUpdates == 2, fmt.Sprintf("2 updates should have been registered, instead there were %d updates", entries[0].NumberOfUpdates))
	th.Assert(t, len(entries[0].TLSVersion) == 2, fmt.Sprintf("TLS first and last should exist, is instead %+v", entries[0].TLSVersion))
	th.Assert(t, entries[0].TLSVersion["first"] == "TLS 1.2", fmt.Sprintf("TLS first should have been TLS 1.2, is instead %s", entries[0].TLSVersion["first"]))
	th.Assert(t, entries[0].TLSVersion["last"] == nil, fmt.Sprintf("TLS last should have been nil since it should have been the same value, it is instead %s", entries[0].TLSVersion["last"]))
	th.Assert(t, entries[0].Operation["first"] == "I", fmt.Sprintf("First operation should have been 'I', is instead %s", entries[0].Operation["first"]))
	th.Assert(t, entries[0].Operation["last"] == "U", fmt.Sprintf("Last operation should have been 'U', is instead %s", entries[0].Operation["last"]))
	th.Assert(t, len(entries[0].Errors) == 1, fmt.Sprintf("Errors length should be 1, is instead %d", len(entries[0].Errors)))
	th.Assert(t, entries[0].Errors[0].Error == "Smart Response Failed", fmt.Sprintf("Errors should include 'Smart Response Failed', is instead %s", entries[0].Errors[0].Error))
	th.Assert(t, entries[0].Errors[0].ErrorCount == 1, fmt.Sprintf("Error Count should be 1, is instead %d", entries[0].Errors[0].ErrorCount))

	// Add 3rd endpoint (with different data) and make sure values are correct

	err = addFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo2, time.Now().Format("2006-01-02 15:04:05.000000000"), idCount, "U", 2)
	th.Assert(t, err == nil, err)
	err = ctStatement.QueryRow(testFhirEndpointInfo.URL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 3, fmt.Sprintf("Should have got 3, intead got %d", count))

	entries, err = CreateArchive(ctx, store, formatToday, formatTomorrow, numWorkers, workerDur)
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

	entries, err = CreateArchive(ctx, store, formatTomorrow, twoDays, numWorkers, workerDur)
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

	// If there is no capability statement, FHIRVersion should be null instead of empty string

	emptyCap := []byte("null")
	cs, err := capabilityparser.NewCapabilityStatement(emptyCap)
	th.Assert(t, err == nil, err)
	testFhirEndpointInfo.CapabilityStatement = cs
	testFhirEndpointInfo.CapabilityFhirVersion = ""

	err = addFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Format("2006-01-02 15:04:05.000000000"), idCount, "U", 1)
	th.Assert(t, err == nil, err)
	err = ctStatement.QueryRow(testFhirEndpointInfo.URL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 1, fmt.Sprintf("Should have got 1, intead got %d", count))

	resultCh2 := make(chan Result)
	jobArgs2 := make(map[string]interface{})
	jobArgs2["historyArgs"] = historyArgs{
		fhirURL:   "http://example.com/DTSU2/",
		dateStart: formatToday,
		dateEnd:   formatTomorrow,
		store:     store,
		result:    resultCh2,
	}

	go getHistory(ctx, &jobArgs2)

	for res := range resultCh2 {
		th.Assert(t, res.URL == "http://example.com/DTSU2/", fmt.Sprintf("Expected URL to equal 'http://example.com/DTSU2/'. Is actually '%s'.", res.URL))
		th.Assert(t, res.Summary.NumberOfUpdates == 1, fmt.Sprintf("1 update should have been registered, instead there were %d updates", res.Summary.NumberOfUpdates))
		th.Assert(t, res.Summary.FHIRVersion["first"] == nil, fmt.Sprintf("FHIR Version first should have been nil, is instead %s", res.Summary.FHIRVersion["first"]))
		close(resultCh2)
	}

	// Base Case

	setupCapabilityStatement(t, filepath.Join("../testdata", "cerner_capability_dstu2.json"))
	err = addFHIREndpointInfoHistory(ctx, store, testFhirEndpointInfo, time.Now().Format("2006-01-02 15:04:05.000000000"), idCount, "U", 1)
	th.Assert(t, err == nil, err)
	err = ctStatement.QueryRow(testFhirEndpointInfo.URL).Scan(&count)
	th.Assert(t, err == nil, err)
	th.Assert(t, count == 2, fmt.Sprintf("Should have got 2, intead got %d", count))

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
		th.Assert(t, res.Summary.NumberOfUpdates == 2, fmt.Sprintf("2 updates should have been registered, instead there were %d updates", res.Summary.NumberOfUpdates))
		th.Assert(t, res.Summary.TLSVersion["first"] == "TLS 1.2", fmt.Sprintf("TLS first should have been TLS 1.2, is instead %s", res.Summary.TLSVersion["first"]))
		th.Assert(t, res.Summary.TLSVersion["last"] == nil, fmt.Sprintf("TLS last should have been nil, it is instead %s", res.Summary.TLSVersion["last"]))
		th.Assert(t, res.Summary.FHIRVersion["last"] == "1.0.2", fmt.Sprintf("FHIR Version last should have been 1.0.2, is instead %+v", res.Summary.FHIRVersion["last"]))
		close(resultCh)
	}

	// If the args are not properly formatted

	jobArgs3 := make(map[string]interface{})
	jobArgs3["historyArgs"] = map[string]interface{}{
		"nonsense": 1,
	}

	err = getHistory(ctx, &jobArgs3)
	th.Assert(t, err != nil, fmt.Sprint("Malformed arguments should have thrown error."))

	// If the URL does not exist, return default data

	resultCh4 := make(chan Result)
	jobArgs4 := make(map[string]interface{})
	jobArgs4["historyArgs"] = historyArgs{
		fhirURL:   "thisurldoesntexist.com",
		dateStart: formatToday,
		dateEnd:   formatTomorrow,
		store:     store,
		result:    resultCh4,
	}

	go getHistory(ctx, &jobArgs4)
	for res := range resultCh4 {
		th.Assert(t, res.Summary.NumberOfUpdates == 0, fmt.Sprintf("Expected 0 entries in history table. Actually had %d entries.", res.Summary.NumberOfUpdates))
		th.Assert(t, res.URL == "thisurldoesntexist.com", fmt.Sprintf("Expected URL to equal 'thisurldoesntexist.com'. Is actually '%s'.", res.URL))
		th.Assert(t, res.Summary.TLSVersion["first"] == nil, fmt.Sprint("TLS first should have been nil"))
		th.Assert(t, res.Summary.TLSVersion["last"] == nil, fmt.Sprint("TLS last should have been nil"))
		close(resultCh4)
	}
}

func Test_getMetadata(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()
	setupCapabilityStatement(t, filepath.Join("../testdata", "cerner_capability_dstu2.json"))

	// Get today and tomorrow's date
	today := time.Now()
	formatTomorrow := today.Add(time.Hour * 24).Format("2006-01-02")
	formatToday := today.Format("2006-01-02")

	// Add Metadata for Endpoint
	_, err = store.AddFHIREndpointMetadata(ctx, &testMetadata)
	th.Assert(t, err == nil, err)

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

	go getMetadata(ctx, &jobArgs)

	for res := range resultCh {
		th.Assert(t, res.URL == "http://example.com/DTSU2/", fmt.Sprintf("Expected URL to equal 'http://example.com/DTSU2/'. Is actually '%s'.", res.URL))
		th.Assert(t, len(res.Summary.SmartHTTPResponse) == 1, fmt.Sprintf("There should be 1 entry for the SMART HTTP Response, is instead %d", len(res.Summary.SmartHTTPResponse)))
		th.Assert(t, res.Summary.SmartHTTPResponse[0].ResponseCode == 400, fmt.Sprintf("SMART HTTP Response Code should be 400, is instead %d", res.Summary.SmartHTTPResponse[0].ResponseCode))
		th.Assert(t, res.Summary.SmartHTTPResponse[0].ResponseCount == 1, fmt.Sprintf("SMART HTTP Response Count should be 1, is instead %d", res.Summary.SmartHTTPResponse[0].ResponseCount))
		close(resultCh)
	}

	// Add 2nd Metadata for Endpoint
	_, err = store.AddFHIREndpointMetadata(ctx, &testMetadata2)
	th.Assert(t, err == nil, err)

	resultCh2 := make(chan Result)
	jobArgs2 := make(map[string]interface{})
	jobArgs2["historyArgs"] = historyArgs{
		fhirURL:   "http://example.com/DTSU2/",
		dateStart: formatToday,
		dateEnd:   formatTomorrow,
		store:     store,
		result:    resultCh2,
	}

	go getMetadata(ctx, &jobArgs2)

	for res := range resultCh2 {
		th.Assert(t, res.URL == "http://example.com/DTSU2/", fmt.Sprintf("Expected URL to equal 'http://example.com/DTSU2/'. Is actually '%s'.", res.URL))
		th.Assert(t, len(res.Summary.SmartHTTPResponse) == 2, fmt.Sprintf("SMART HTTP Response should have 2 entries, instead has %d", len(res.Summary.SmartHTTPResponse)))
		th.Assert(t, len(res.Summary.HTTPResponse) == 1, fmt.Sprintf("HTTP Response should have 1 entry, instead has %d", len(res.Summary.HTTPResponse)))
		th.Assert(t, res.Summary.HTTPResponse[0].ResponseCode == 200, fmt.Sprintf("HTTP Response Code should be 200, is instead %d", res.Summary.HTTPResponse[0].ResponseCode))
		th.Assert(t, res.Summary.HTTPResponse[0].ResponseCount == 2, fmt.Sprintf("HTTP Response Count should be 2, is instead %d", res.Summary.HTTPResponse[0].ResponseCount))
		th.Assert(t, len(res.Summary.Errors) == 1, fmt.Sprintf("Errors should have 1 entry, instead has %d", len(res.Summary.Errors)))
		th.Assert(t, res.Summary.ResponseTimeSecond == 0.9, fmt.Sprintf("HTTP Response Code should be 0.9, the median of [0.8, 1.0], is instead %f", res.Summary.ResponseTimeSecond))
		close(resultCh2)
	}

	// Add 3nd Metadata for Endpoint
	_, err = store.AddFHIREndpointMetadata(ctx, &testMetadata)
	th.Assert(t, err == nil, err)

	resultCh3 := make(chan Result)
	jobArgs3 := make(map[string]interface{})
	jobArgs3["historyArgs"] = historyArgs{
		fhirURL:   "http://example.com/DTSU2/",
		dateStart: formatToday,
		dateEnd:   formatTomorrow,
		store:     store,
		result:    resultCh3,
	}

	go getMetadata(ctx, &jobArgs3)

	for res := range resultCh3 {
		th.Assert(t, res.URL == "http://example.com/DTSU2/", fmt.Sprintf("Expected URL to equal 'http://example.com/DTSU2/'. Is actually '%s'.", res.URL))
		th.Assert(t, len(res.Summary.SmartHTTPResponse) == 2, fmt.Sprintf("SMART HTTP Response should have 2 entries, instead has %d", len(res.Summary.SmartHTTPResponse)))
		th.Assert(t, len(res.Summary.HTTPResponse) == 1, fmt.Sprintf("HTTP Response should have 1 entry, instead has %d", len(res.Summary.HTTPResponse)))
		th.Assert(t, res.Summary.HTTPResponse[0].ResponseCode == 200, fmt.Sprintf("HTTP Response Code should be 200, is instead %d", res.Summary.HTTPResponse[0].ResponseCode))
		th.Assert(t, res.Summary.HTTPResponse[0].ResponseCount == 3, fmt.Sprintf("HTTP Response Count should be 2, is instead %d", res.Summary.HTTPResponse[0].ResponseCount))
		th.Assert(t, len(res.Summary.Errors) == 1, fmt.Sprintf("Errors should have 1 entry, instead has %d", len(res.Summary.Errors)))
		th.Assert(t, res.Summary.ResponseTimeSecond == 0.8, fmt.Sprintf("HTTP Response Code should be 0.8, the median of [0.8, 0.8, 1.0], is instead %f", res.Summary.ResponseTimeSecond))
		close(resultCh3)
	}

	// If the args are not properly formatted

	jobArgs4 := make(map[string]interface{})
	jobArgs4["historyArgs"] = map[string]interface{}{
		"nonsense": 1,
	}

	err = getMetadata(ctx, &jobArgs4)
	th.Assert(t, err != nil, fmt.Sprint("Malformed arguments should have thrown error."))

	// If the URL does not exist, return default data

	resultCh5 := make(chan Result)
	jobArgs5 := make(map[string]interface{})
	jobArgs5["historyArgs"] = historyArgs{
		fhirURL:   "thisurldoesntexist.com",
		dateStart: formatToday,
		dateEnd:   formatTomorrow,
		store:     store,
		result:    resultCh5,
	}

	go getMetadata(ctx, &jobArgs5)
	for res := range resultCh5 {
		th.Assert(t, len(res.Summary.HTTPResponse) == 0, fmt.Sprintf("HTTP Response should have 0 entries, instead has %d", len(res.Summary.HTTPResponse)))
		th.Assert(t, len(res.Summary.SmartHTTPResponse) == 0, fmt.Sprintf("SMART HTTP Response should have 0 entries, instead has %d", len(res.Summary.SmartHTTPResponse)))
		th.Assert(t, len(res.Summary.Errors) == 0, fmt.Sprintf("Errors should have 0 entries, instead has %d", len(res.Summary.Errors)))
		th.Assert(t, res.Summary.ResponseTimeSecond == nil, fmt.Sprintf("ResponseTimeSecond should be 0, instead is %f", res.Summary.ResponseTimeSecond))
		close(resultCh5)
	}
}

func setupCapabilityStatement(t *testing.T, path string) {
	// capability statement
	csJSON, err := ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)
	testFhirEndpointInfo.CapabilityStatement = cs
	fhirVersion, err := cs.GetFHIRVersion()
	testFhirEndpointInfo.CapabilityFhirVersion = fhirVersion
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
		capabilityStatementJSON,
		e.CapabilityFhirVersion)
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
