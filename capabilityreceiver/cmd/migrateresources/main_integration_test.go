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

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/spf13/viper"
)

var store *postgresql.Store

var capStat1 capabilityparser.CapabilityStatement
var capStat2 capabilityparser.CapabilityStatement

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

func Test_updateOperationResource(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	setupCapabilityStatements(t, filepath.Join("../../testdata", "cerner_capability_dstu2.json"), filepath.Join("../../testdata", "epic_capability_dstu2.json"))
	ctx := context.Background()

	addFHIREndpointInfoHistoryStatement, err := store.DB.Prepare(`
		INSERT INTO fhir_endpoints_info_history (
			url,
			capability_statement,
			entered_at
		)
		VALUES ($1, $2, $3)`)
	th.Assert(t, err == nil, "addFHIREndpointInfoHistoryStatement should not have returned an error")
	defer addFHIREndpointInfoHistoryStatement.Close()

	getFHIREndpointInfoHistoryStatement := `
		SELECT entered_at, operation_resource
		FROM fhir_endpoints_info_history
		WHERE url=$1
	`
	// Put two FHIR endpoints in the history table
	firstTime := time.Now()
	url1 := "www.testurl.com/cerner/DSTU2"
	addFHIREndpointInfoHistoryStatement.ExecContext(ctx, url1, capStat1, firstTime)

	secondTime := time.Now()
	url2 := "www.testurl.com/epic/DSTU2"
	addFHIREndpointInfoHistoryStatement.ExecContext(ctx, url2, capStat2, secondTime)

	resultCh := make(chan Result)

	// Check that data only updates the first URL
	defaultArgs := make(map[string]interface{})
	defaultArgs["historyArgs"] = historyArgs{
		fhirURL: url1,
		store:   store,
		result:  resultCh,
	}

	go updateOperationResource(ctx, &defaultArgs)
	for res := range resultCh {
		th.Assert(t, res.URL == url1, fmt.Sprintf("Returned result URL is not equal to %s, is instead %s", url1, res.URL))
		close(resultCh)
	}

	historyRows, err := store.DB.QueryContext(ctx, getFHIREndpointInfoHistoryStatement, url1)
	th.Assert(t, err == nil, fmt.Sprintf("error getting data from fhir_endpoints_info_history: %s", err))
	// Loop through the rows
	defer historyRows.Close()
	count := 0
	for historyRows.Next() {
		th.Assert(t, count <= 1, fmt.Sprintf("should only be one item in the database for this URL"))
		var receivedTime time.Time
		var operationResource map[string][]string
		err = historyRows.Scan(&receivedTime, &operationResource)
		th.Assert(t, err == nil, fmt.Sprintf("Error while scanning the rows of the history table for URL %s. Error: %s", url1, err))
		th.Assert(t, receivedTime == firstTime, fmt.Sprintf("The time was updated to %+v", receivedTime))
		th.Assert(t, len(operationResource) == 2, fmt.Sprintf("The number of operation resources should have been 2. Is instead %d", len(operationResource)))
		th.Assert(t, operationResource["read"] != nil, "There should be a read resource defined, instead is nil")
		th.Assert(t, operationResource["search-type"] != nil, "There should be a search-type resource defined, instead is nil")
		th.Assert(t, len(operationResource["read"]) == 25, fmt.Sprintf("The number of operation resources for read should have been 25. Is instead %d", len(operationResource["read"])))
		th.Assert(t, len(operationResource["search-type"]) == 23, fmt.Sprintf("The number of operation resources should have been 23. Is instead %d", len(operationResource["search-type"])))
		count++
	}

	// Add another instance of the second URL
	thirdTime := time.Now()
	addFHIREndpointInfoHistoryStatement.ExecContext(ctx, url2, capStat2, thirdTime)

	// Make sure all instances of that are updated
	resultCh2 := make(chan Result)

	// Check that data only updates the first URL
	defaultArgs2 := make(map[string]interface{})
	defaultArgs2["historyArgs"] = historyArgs{
		fhirURL: url2,
		store:   store,
		result:  resultCh2,
	}

	go updateOperationResource(ctx, &defaultArgs2)
	for res := range resultCh2 {
		th.Assert(t, res.URL == url2, fmt.Sprintf("Returned result URL is not equal to %s, is instead %s", url1, res.URL))
		close(resultCh2)
	}
	historyRows, err = store.DB.QueryContext(ctx, getFHIREndpointInfoHistoryStatement, url2)
	th.Assert(t, err == nil, fmt.Sprintf("error getting data from fhir_endpoints_info_history: %s", err))
	// Loop through the rows
	defer historyRows.Close()
	count = 0
	for historyRows.Next() {
		th.Assert(t, count <= 2, fmt.Sprintf("should be two items in the database for this URL"))
		var receivedTime time.Time
		var operationResource map[string][]string
		err = historyRows.Scan(&receivedTime, &operationResource)
		th.Assert(t, err == nil, fmt.Sprintf("Error while scanning the rows of the history table for URL %s. Error: %s", url1, err))
		th.Assert(t, len(operationResource) == 2, fmt.Sprintf("The number of operation resources should have been 2. Is instead %d", len(operationResource)))
		th.Assert(t, operationResource["read"] != nil, "There should be a read resource defined, instead is nil")
		th.Assert(t, operationResource["search-type"] != nil, "There should be a search-type resource defined, instead is nil")
		th.Assert(t, len(operationResource["read"]) == 16, fmt.Sprintf("The number of operation resources for read should have been 16. Is instead %d", len(operationResource["read"])))
		th.Assert(t, len(operationResource["search-type"]) == 16, fmt.Sprintf("The number of operation resources should have been 16. Is instead %d", len(operationResource["search-type"])))
		count++
	}
}

func setupCapabilityStatements(t *testing.T, path1 string, path2 string) {
	// capability statement
	csJSON, err := ioutil.ReadFile(path1)
	th.Assert(t, err == nil, err)
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)
	capStat1 = cs

	csJSON2, err := ioutil.ReadFile(path2)
	th.Assert(t, err == nil, err)
	cs2, err := capabilityparser.NewCapabilityStatement(csJSON2)
	th.Assert(t, err == nil, err)
	capStat2 = cs2
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
