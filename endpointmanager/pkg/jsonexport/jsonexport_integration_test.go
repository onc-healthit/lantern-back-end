// +build integration

package jsonexport

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/smartparser"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/spf13/viper"
)

var store *postgresql.Store

var testEndpointOrganization = &endpointmanager.FHIREndpointOrganization{
	OrganizationName: "Test Org"}

var testEndpoint = endpointmanager.FHIREndpoint{
	URL:               "www.testURL.com",
	OrganizationList: []*endpointmanager.FHIREndpointOrganization{testEndpointOrganization},
	ListSource:        "Test List Source",
}

var testEndpointMetadata = endpointmanager.FHIREndpointMetadata{
	HTTPResponse:      200,
	SMARTHTTPResponse: 200,
	ResponseTime:      0.345,
	Availability:      1.00,
}

var testEndpointInfo = endpointmanager.FHIREndpointInfo{
	URL:        "www.testURL.com",
	TLSVersion: "TLS 1.3",
	MIMETypes:  []string{"application/fhir+json"},
	OperationResource: map[string][]string{
		"read":        []string{"AllergyIntolerance", "Conformance"},
		"search-type": []string{"AllergyIntolerance"},
	},
	Metadata: &testEndpointMetadata,
}

var firstEndpoint = testEndpointInfo
var secondEndpoint = testEndpointInfo

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

func Test_createJSON(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var actualNumEndptsStored int
	var jsonAsObj []jsonEntry

	ctx := context.Background()
	err := store.AddFHIREndpoint(ctx, &testEndpoint)
	th.Assert(t, err == nil, fmt.Sprintf("Error while adding a FHIR Endpoint. Error: %s", err))

	metadataID, err := store.AddFHIREndpointMetadata(ctx, firstEndpoint.Metadata)
	valResID1, err := store.AddValidationResult(ctx)
	th.Assert(t, err == nil, fmt.Sprintf("Error adding validation result ID: %s", err))
	firstEndpoint.ValidationID = valResID1
	err = store.AddFHIREndpointInfo(ctx, &firstEndpoint, metadataID)
	th.Assert(t, err == nil, fmt.Sprintf("Error while adding the FHIR Endpoint Info. Error: %s", err))

	secondEndpoint.ID = firstEndpoint.ID

	metadataID, err = store.AddFHIREndpointMetadata(ctx, secondEndpoint.Metadata)
	valResID2, err := store.AddValidationResult(ctx)
	th.Assert(t, err == nil, fmt.Sprintf("Error adding validation result ID: %s", err))
	secondEndpoint.ValidationID = valResID2
	err = store.UpdateFHIREndpointInfo(ctx, &secondEndpoint, metadataID)
	th.Assert(t, err == nil, fmt.Sprintf("Error while updating the FHIR Endpoint Info. Error: %s", err))

	rows := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info_history;")
	err = rows.Scan(&actualNumEndptsStored)
	th.Assert(t, err == nil, fmt.Sprintf("Error while getting number of endpoints from the history table. Error: %s", err))
	th.Assert(t, actualNumEndptsStored == 2, fmt.Sprintf("Expected 2 endpoints stored. Actually had %d endpoints stored.", actualNumEndptsStored))

	// Base case

	returnedJSON, err := createJSON(ctx, store, "30days")
	th.Assert(t, err == nil, fmt.Sprintf("Error returned from the createJSON function: %s", err))
	err = json.Unmarshal(returnedJSON, &jsonAsObj)
	th.Assert(t, err == nil, fmt.Sprintf("Error while unmarshalling the JSON. Error: %s", err))
	th.Assert(t, len(jsonAsObj) == 1, fmt.Sprintf("Expected 1 endpoints in JSON. Actually had %d endpoints stored.", len(jsonAsObj)))
	th.Assert(t, jsonAsObj[0].URL == "www.testURL.com", fmt.Sprintf("Expected URL to equal 'www.testURL.com'. Is actually '%s'.", jsonAsObj[0].URL))
	th.Assert(t, len(jsonAsObj[0].Operation) == 2, fmt.Sprintf("Expected 2 history values in JSON. Actually had %d endpoints stored.", len(jsonAsObj[0].Operation)))
}

func Test_getHistory(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var actualNumEndptsStored int

	ctx := context.Background()
	err := store.AddFHIREndpoint(ctx, &testEndpoint)
	th.Assert(t, err == nil, err)

	metadataID, err := store.AddFHIREndpointMetadata(ctx, firstEndpoint.Metadata)
	valResID1, err := store.AddValidationResult(ctx)
	th.Assert(t, err == nil, fmt.Sprintf("Error adding validation result ID: %s", err))
	firstEndpoint.ValidationID = valResID1
	err = store.AddFHIREndpointInfo(ctx, &firstEndpoint, metadataID)
	th.Assert(t, err == nil, err)

	rows := store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info_history;")
	err = rows.Scan(&actualNumEndptsStored)
	th.Assert(t, err == nil, err)
	th.Assert(t, actualNumEndptsStored == 1, fmt.Sprintf("Expected 1 endpoint stored. Actually had %d endpoints stored.", actualNumEndptsStored))

	// Base case

	resultCh := make(chan Result)
	jobArgs := make(map[string]interface{})
	jobArgs["historyArgs"] = historyArgs{
		fhirURL: "www.testURL.com",
		store:   store,
		result:  resultCh,
		exportType: "30days",
	}

	go getHistory(ctx, &jobArgs)

	for res := range resultCh {
		th.Assert(t, len(res.Rows) == 1, fmt.Sprintf("Expected 1 entry in history table. Actually had %d entries.", len(res.Rows)))
		th.Assert(t, res.URL == "www.testURL.com", fmt.Sprintf("Expected URL to equal 'www.testURL.com'. Is actually '%s'.", res.URL))
		th.Assert(t, res.Rows[0].TLSVersion == "TLS 1.3", fmt.Sprintf("Should be the current entry in the fhir_endpoints_info table. %+v", res.Rows[0].TLSVersion))
		close(resultCh)
	}

	// base case with export type equal to month

	// LANTERN-726: (Special Case) Subtract the day by 1 if it is the 31st of a month for accurate calculation of oldDate
	now := time.Now()
	if now.Day() == 31 {
		now = now.AddDate(0, 0, -1)
	}

	// Subtract the day again by 2 so that the oldDate is Feb 28 if the current date is March 31.
	if now.Month() == time.March {
		now = now.AddDate(0, 0, -2)
	}

	oldDate := now.AddDate(0, -1, 0).Format("2006-01-02 15:04:05.000000000")

	updateEndpointInfoHistoryDate := `
	UPDATE fhir_endpoints_info_history
	SET
		updated_at = $1
	WHERE url = $2 AND validation_result_id = $3;`

	_, err = store.DB.ExecContext(ctx, updateEndpointInfoHistoryDate, oldDate, firstEndpoint.URL, firstEndpoint.ValidationID)
	th.Assert(t, err == nil, fmt.Sprintf("Error when updating updated at time for endpoint in the history table %s", err))

	rows = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info_history;")
	err = rows.Scan(&actualNumEndptsStored)
	th.Assert(t, err == nil, err)
	th.Assert(t, actualNumEndptsStored == 1, fmt.Sprintf("Expected 1 endpoints stored. Actually had %d endpoints stored.", actualNumEndptsStored))

	resultChMonth := make(chan Result)
	jobArgsMonth := make(map[string]interface{})
	jobArgsMonth["historyArgs"] = historyArgs{
		fhirURL: "www.testURL.com",
		store:   store,
		result:  resultChMonth,
		exportType: "month",
	}

	go getHistory(ctx, &jobArgsMonth)

	for res := range resultChMonth {
		th.Assert(t, len(res.Rows) == 1, fmt.Sprintf("Expected 1 entry in history table. Actually had %d entries.", len(res.Rows)))
		th.Assert(t, res.URL == "www.testURL.com", fmt.Sprintf("Expected URL to equal 'www.testURL.com'. Is actually '%s'.", res.URL))
		th.Assert(t, res.Rows[0].TLSVersion == "TLS 1.3", fmt.Sprintf("Should be the current entry in the fhir_endpoints_info table. %+v", res.Rows[0].TLSVersion))
		close(resultChMonth)
	}

	// base case with export type equal to all


	_, err = store.DB.Exec("DElETE FROM fhir_endpoints_info;")
	th.Assert(t, err == nil, err)

	metadataID2, err := store.AddFHIREndpointMetadata(ctx, secondEndpoint.Metadata)
	valResID2, err := store.AddValidationResult(ctx)
	th.Assert(t, err == nil, fmt.Sprintf("Error adding validation result ID: %s", err))
	secondEndpoint.ValidationID = valResID2
	err = store.AddFHIREndpointInfo(ctx, &secondEndpoint, metadataID2)
	th.Assert(t, err == nil, err)

	rows = store.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info_history;")
	err = rows.Scan(&actualNumEndptsStored)
	th.Assert(t, err == nil, err)
	// 3 endpoints should be stored since 1 entry will also be added for deleting from the fhir_endpoints_info table
	th.Assert(t, actualNumEndptsStored == 3, fmt.Sprintf("Expected 3 endpoints stored. Actually had %d endpoints stored.", actualNumEndptsStored))


	resultChAll := make(chan Result)
	jobArgsAll := make(map[string]interface{})
	jobArgsAll["historyArgs"] = historyArgs{
		fhirURL: "www.testURL.com",
		store:   store,
		result:  resultChAll,
		exportType: "all",
	}

	go getHistory(ctx, &jobArgsAll)

	for res := range resultChAll {
		th.Assert(t, len(res.Rows) == 3, fmt.Sprintf("Expected 3 entries in history table. Actually had %d entries.", len(res.Rows)))
		th.Assert(t, res.URL == "www.testURL.com", fmt.Sprintf("Expected URL to equal 'www.testURL.com'. Is actually '%s'.", res.URL))
		th.Assert(t, res.Rows[0].TLSVersion == "TLS 1.4", fmt.Sprintf("Should be the current entry in the fhir_endpoints_info table. %+v", res.Rows[0].TLSVersion))
		close(resultChAll)
	}

	// If the args are not properly formatted

	jobArgs2 := make(map[string]interface{})
	jobArgs2["historyArgs"] = map[string]interface{}{
		"nonsense": 1,
	}

	err = getHistory(ctx, &jobArgs2)
	th.Assert(t, err != nil, fmt.Sprint("Malformed arguments should have thrown error."))

	// If the URL does not exist, return an empty array

	resultCh3 := make(chan Result)
	jobArgs3 := make(map[string]interface{})
	jobArgs3["historyArgs"] = historyArgs{
		fhirURL: "thisurldoesntexist.com",
		store:   store,
		result:  resultCh3,
		exportType: "30days",
	}

	go getHistory(ctx, &jobArgs3)
	for res := range resultCh3 {
		th.Assert(t, len(res.Rows) == 0, fmt.Sprintf("Expected 0 entries in history table. Actually had %d entries.", len(res.Rows)))
		th.Assert(t, res.URL == "thisurldoesntexist.com", fmt.Sprintf("Expected URL to equal 'thisurldoesntexist.com'. Is actually '%s'.", res.URL))
		close(resultCh3)
	}
}

func setup() error {
	var err error
	store, err = postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	if err != nil {
		return err
	}
	// add test capability statement and smart response
	capStat, _ := capabilityparser.NewCapabilityStatement([]byte(`
	{
		"fhirVersion": "4.0.1",
		"kind": "instance"
	}`))
	smartResp, _ := smartparser.NewSMARTResp([]byte(
		`{
			"authorization_endpoint": "https://ehr.example.com/auth/authorize",
			"token_endpoint": "https://ehr.example.com/auth/token"
		}`))
	firstEndpoint.CapabilityStatement = capStat
	firstEndpoint.SMARTResponse = smartResp
	secondEndpoint.CapabilityStatement = capStat
	secondEndpoint.SMARTResponse = smartResp
	secondEndpoint.TLSVersion = "TLS 1.4"

	return nil
}

func teardown() {
	store.Close()
}
