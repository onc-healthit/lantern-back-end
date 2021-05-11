// +build integration

package main

import (
	"context"
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
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/spf13/viper"
)

var store *postgresql.Store

var capStat1 []byte
var capStat2 []byte

var testMetadata1 = endpointmanager.FHIREndpointMetadata{
	HTTPResponse:      200,
	SMARTHTTPResponse: 200,
}
var testMetadata2 = endpointmanager.FHIREndpointMetadata{
	HTTPResponse:      400,
	SMARTHTTPResponse: 200,
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

func Test_addToValidationTable(t *testing.T) {
	// _, _ = th.IntegrationDBTestSetup(t, store.DB)
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	setupCapabilityStatements(t, filepath.Join("../../testdata", "cerner_capability_dstu2.json"), filepath.Join("../../testdata", "epic_capability_dstu2.json"))
	ctx := context.Background()

	addFHIREndpointInfoStatement := `
		INSERT INTO fhir_endpoints_info_history (
			url,
			operation,
			capability_statement,
			tls_version,
			mime_types,
			metadata_id,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	getFHIREndpointInfoStatement := `
		SELECT updated_at, validation_result_id
		FROM fhir_endpoints_info_history
		WHERE url=$1
	`

	getValidationResultStatement := `
		SELECT COUNT(*)
		FROM validation_results
		WHERE id=$1
	`

	getValidationStatement := `
		SELECT COUNT(*)
		FROM validations
		WHERE validation_result_id=$1
	`

	// Add metadata
	metadataID1, err := store.AddFHIREndpointMetadata(ctx, &testMetadata1)
	th.Assert(t, err == nil, fmt.Sprintf("Error while adding first metadata object: %s", err))
	metadataID2, err := store.AddFHIREndpointMetadata(ctx, &testMetadata2)
	th.Assert(t, err == nil, fmt.Sprintf("Error while adding first metadata object: %s", err))

	// Put two FHIR endpoints in the history table
	tlsVersion := "1.3"
	mimeTypes := []string{"application/json+fhir"}
	firstTime := time.Now().UTC().Round(time.Microsecond)
	url1 := "www.testurl.com/cerner/DSTU2"
	_, err = store.DB.ExecContext(ctx, addFHIREndpointInfoStatement, url1, "I", capStat1, tlsVersion, pq.Array(mimeTypes), metadataID1, firstTime)
	th.Assert(t, err == nil, fmt.Sprintf("Error when adding to the database %s", err))

	secondTime := time.Now().UTC().Round(time.Microsecond)
	url2 := "www.testurl.com/epic/DSTU2"
	_, err = store.DB.ExecContext(ctx, addFHIREndpointInfoStatement, url2, "I", capStat2, tlsVersion, pq.Array(mimeTypes), metadataID2, secondTime)
	th.Assert(t, err == nil, fmt.Sprintf("Error when adding to the database again %s", err))

	resultCh := make(chan Result)

	// Check that data only updates the first URL
	defaultArgs := make(map[string]interface{})
	defaultArgs["historyArgs"] = historyArgs{
		fhirURL:   url1,
		store:     store,
		result:    resultCh,
		isHistory: true,
	}

	go addToValidationTable(ctx, &defaultArgs)
	for res := range resultCh {
		th.Assert(t, res.URL == url1, fmt.Sprintf("Returned result URL is not equal to %s, is instead %s", url1, res.URL))
		close(resultCh)
	}

	historyRows, err := store.DB.QueryContext(ctx, getFHIREndpointInfoStatement, url1)
	th.Assert(t, err == nil, fmt.Sprintf("error getting data from fhir_endpoints_info: %s", err))
	// Check to make sure exactly 1 ID was updated
	count := 0
	var valID int
	for historyRows.Next() {
		th.Assert(t, count < 1, fmt.Sprintf("should only be one item in the database for this URL"))
		var receivedTime time.Time
		err = historyRows.Scan(&receivedTime, &valID)
		th.Assert(t, err == nil, fmt.Sprintf("Error while scanning the rows of the history table for URL %s. Error: %s", url1, err))
		th.Assert(t, receivedTime.Equal(firstTime), fmt.Sprintf("The time was updated to %+v from %+v", receivedTime, firstTime))
		count++
	}
	th.Assert(t, count == 1, "should be one item in the database, instead is 0")

	// Make sure that ID was added to the validation results table
	valResRows, err := store.DB.QueryContext(ctx, getValidationResultStatement, valID)
	th.Assert(t, err == nil, fmt.Sprintf("error getting data from validation_results: %s", err))
	defer valResRows.Close()

	for valResRows.Next() {
		var valResCount int
		err = valResRows.Scan(&valResCount)
		th.Assert(t, valResCount == 1, fmt.Sprintf("for URL %s, there should be one row with id %d", url1, valID))
		count++
	}

	// Make sure that 5 entries were added to the validation table with that ID
	valRows, err := store.DB.QueryContext(ctx, getValidationStatement, valID)
	th.Assert(t, err == nil, fmt.Sprintf("error getting data from validations: %s", err))
	defer valRows.Close()

	for valRows.Next() {
		var valCount int
		err = valRows.Scan(&valCount)
		th.Assert(t, valCount == 5, fmt.Sprintf("there should be 5 entries in the validations table with id %d, instead there are %d", valID, valCount))
		count++
	}

	// Add another instance of the second URL
	thirdTime := time.Now().UTC().Round(time.Microsecond)
	_, err = store.DB.ExecContext(ctx, addFHIREndpointInfoStatement, url2, "U", capStat2, tlsVersion, pq.Array(mimeTypes), metadataID2, thirdTime)
	th.Assert(t, err == nil, fmt.Sprintf("Error when adding to the database third time %s", err))

	// Make sure all instances of that are updated
	resultCh2 := make(chan Result)

	// Check that data only updates the second URL
	defaultArgs2 := make(map[string]interface{})
	defaultArgs2["historyArgs"] = historyArgs{
		fhirURL:   url2,
		store:     store,
		result:    resultCh2,
		isHistory: true,
	}

	go addToValidationTable(ctx, &defaultArgs2)
	for res := range resultCh2 {
		th.Assert(t, res.URL == url2, fmt.Sprintf("Returned result URL is not equal to %s, is instead %s", url1, res.URL))
		close(resultCh2)
	}
	historyRows, err = store.DB.QueryContext(ctx, getFHIREndpointInfoStatement, url2)
	th.Assert(t, err == nil, fmt.Sprintf("error getting data from fhir_endpoints_info: %s", err))
	// Check that both entries with url2 have been updated and that they don't have the same validation result ID
	count = 0
	var firstValID int
	for historyRows.Next() {
		th.Assert(t, count < 2, fmt.Sprintf("should be two items in the database for this URL"))
		var receivedTime time.Time
		var currentValID int
		err = historyRows.Scan(&receivedTime, &currentValID)
		th.Assert(t, err == nil, fmt.Sprintf("Error while scanning the rows of the history table for URL %s. Error: %s", url1, err))
		th.Assert(t, currentValID != 0, fmt.Sprintf("The validation ID was not set for this history table entry %s", url1))
		if count == 1 {
			th.Assert(t, currentValID != firstValID, fmt.Sprintf("The second ID should not be equal to the first, %d, %d", currentValID, firstValID))
		} else if count == 0 {
			firstValID = currentValID
		}
		count++
	}
	th.Assert(t, count == 2, "should be two items in the database, instead is 1")

	// Just check the first validation ID to make sure it's in validation_results
	valResRows, err = store.DB.QueryContext(ctx, getValidationResultStatement, firstValID)
	th.Assert(t, err == nil, fmt.Sprintf("error getting data from validation_results: %s", err))
	defer valResRows.Close()

	for valResRows.Next() {
		var valResCount int
		err = valResRows.Scan(&valResCount)
		th.Assert(t, valResCount == 1, fmt.Sprintf("for URL %s, there should be one row with id %d", url1, firstValID))
		count++
	}

	// Then check that it's 5 validation entries were added to the validation table
	valRows, err = store.DB.QueryContext(ctx, getValidationStatement, valID)
	th.Assert(t, err == nil, fmt.Sprintf("error getting data from validations: %s", err))
	defer valRows.Close()

	for valRows.Next() {
		var valCount int
		err = valRows.Scan(&valCount)
		th.Assert(t, valCount == 5, fmt.Sprintf("there should only be 5 entries in the validations table with id %d, instead there are %d", valID, valCount))
		count++
	}
}

func setupCapabilityStatements(t *testing.T, path1 string, path2 string) {
	// capability statement
	csJSON, err := ioutil.ReadFile(path1)
	th.Assert(t, err == nil, err)
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)
	capStat1, err = cs.GetJSON()
	th.Assert(t, err == nil, err)

	csJSON2, err := ioutil.ReadFile(path2)
	th.Assert(t, err == nil, err)
	cs2, err := capabilityparser.NewCapabilityStatement(csJSON2)
	th.Assert(t, err == nil, err)
	capStat2, err = cs2.GetJSON()
	th.Assert(t, err == nil, err)
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
