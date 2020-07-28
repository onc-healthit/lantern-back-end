package chplquerier

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"

	"github.com/spf13/viper"
)

var testCHPLCrit = chplCertCriteria{
	ID:                     44,
	Number:                 "170.315 (f)(2)",
	Title:                  "Transmission to Public Health Agencies - Syndromic Surveillance",
	CertificationEditionID: 3,
	CertificationEdition:   "2015",
	Description:            "Syndromic Surveillance",
	Removed:                false,
}

var testCrit = endpointmanager.CertificationCriteria{
	CertificationID:        44,
	CertificationNumber:    "170.315 (f)(2)",
	Title:                  "Transmission to Public Health Agencies - Syndromic Surveillance",
	CertificationEditionID: 3,
	CertificationEdition:   "2015",
	Description:            "Syndromic Surveillance",
	Removed:                false,
}

func Test_makeCHPLCriteriaURL(t *testing.T) {

	// basic test

	apiKey := viper.GetString("chplapikey")
	viper.Set("chplapikey", "tmp_api_key")
	defer viper.Set("chplapikey", apiKey)

	expected := "https://chpl.healthit.gov/rest/data/certification-criteria?api_key=tmp_api_key"

	actualURL, err := makeCHPLCriteriaURL()
	th.Assert(t, err == nil, err)

	actual := actualURL.String()
	th.Assert(t, expected == actual, fmt.Sprintf("Expected %s to equal %s.", actual, expected))

	// test empty api key

	viper.Set("chplapikey", "")
	actualURL, err = makeCHPLCriteriaURL()
	th.Assert(t, err != nil, fmt.Sprintf("Expected to return an error due to the api key not being set"))

	// test invalid domain and error handling

	chplDomainOrig := chplDomain
	chplDomain = "http://%41:8080/" // invalid domain
	defer func() { chplDomain = chplDomainOrig }()

	_, err = makeCHPLCriteriaURL()
	switch errors.Cause(err).(type) {
	case *url.Error:
		// ok
	default:
		t.Fatal("Expected url error")
	}
}

func Test_convertCriteriaJSONToObj(t *testing.T) {
	var ctx context.Context
	var err error

	// basic test

	critListJSON := `{
		"criteria": [
		{
			"id": 44,
			"number": "170.315 (f)(2)",
			"title": "Transmission to Public Health Agencies - Syndromic Surveillance",
			"certificationEditionId": 3,
			"certificationEdition": "2015",
			"description": "Syndromic Surveillance",
			"removed": false
		},
		{
			"id": 64,
            "number": "170.314 (a)(4)",
            "title": "Vital signs, body mass index, and growth Charts",
            "certificationEditionId": 2,
            "certificationEdition": "2014",
            "description": "Vital signs",
            "removed": false
		}]}
		`

	expectedCrit1 := testCHPLCrit

	expectedCrit2 := chplCertCriteria{
		ID:                     64,
		Number:                 "170.314 (a)(4)",
		Title:                  "Vital signs, body mass index, and growth Charts",
		CertificationEditionID: 2,
		CertificationEdition:   "2014",
		Description:            "Vital signs",
		Removed:                false,
	}

	expectedCritList := chplCertifiedCriteriaList{
		Results: []chplCertCriteria{expectedCrit1, expectedCrit2},
	}

	ctx = context.Background()
	critList, err := convertCriteriaJSONToObj(ctx, []byte(critListJSON))
	th.Assert(t, err == nil, err)
	th.Assert(t, critList.Results != nil, "Expected results field to be filled out for criteria list.")
	th.Assert(t, len(critList.Results) == len(expectedCritList.Results), fmt.Sprintf("Number of criteria is %d. Should be %d.", len(critList.Results), len(expectedCritList.Results)))

	for i, prod := range critList.Results {
		th.Assert(t, prod == expectedCritList.Results[i], "Expected parsed criteria to equal expected criteria.")
	}

	// test with canceled context

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = convertCriteriaJSONToObj(ctx, []byte(critListJSON))
	th.Assert(t, errors.Cause(err) == context.Canceled, "Expected canceled context error")

	// test with malformed JSON

	ctx = context.Background()
	malformedJSON := `
		"asdf": [
		{}]}
		`

	_, err = convertCriteriaJSONToObj(ctx, []byte(malformedJSON))
	switch errors.Cause(err).(type) {
	case *json.SyntaxError:
		// ok
	default:
		t.Fatal("Expected JSON syntax error")
	}
}

func Test_getCriteriaJSON(t *testing.T) {
	var err error
	var tc *th.TestClient
	var ctx context.Context

	apiKey := viper.GetString("chplapikey")
	viper.Set("chplapikey", "tmp_api_key")
	defer viper.Set("chplapikey", apiKey)

	// basic test

	// mock JSON includes 36 criteria entries
	expectedCriteriaReceived := 36

	tc, err = basicTestCriteriaClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx = context.Background()

	critJSON, err := getCriteriaJSON(ctx, &(tc.Client))
	th.Assert(t, err == nil, err)

	// convert received JSON so we can count the number of entries received
	criteria, err := convertCriteriaJSONToObj(ctx, critJSON)
	th.Assert(t, err == nil, err)
	actualCriteriaReceived := len(criteria.Results)
	th.Assert(t, actualCriteriaReceived == expectedCriteriaReceived, fmt.Sprintf("Expected to receive %d criteria. Actually received %d criteria.", expectedCriteriaReceived, actualCriteriaReceived))

	// test context ended.
	// also checks what happens when an http request fails

	tc, err = basicTestCriteriaClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = getCriteriaJSON(ctx, &(tc.Client))
	switch reqErr := errors.Cause(err).(type) {
	case *url.Error:
		th.Assert(t, reqErr.Err == context.Canceled, "Expected error stating that context was canceled")
	default:
		t.Fatal("Expected context canceled error")
	}

	// test http status != 200

	tc = th.NewTestClientWith404()
	defer tc.Close()

	ctx = context.Background()

	_, err = getCriteriaJSON(ctx, &(tc.Client))
	th.Assert(t, err.Error() == "CHPL request responded with status: 404 Not Found", "expected response error specifying response code")

	// test error on URL creation

	chplDomainOrig := chplDomain
	chplDomain = "http://%41:8080/" // invalid domain
	defer func() { chplDomain = chplDomainOrig }()

	tc, err = basicTestCriteriaClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx = context.Background()

	_, err = getCriteriaJSON(ctx, &(tc.Client))
	switch errors.Cause(err).(type) {
	case *url.Error:
		// ok
	default:
		t.Fatal("Expected url error")
	}
}

func basicTestCriteriaClient() (*th.TestClient, error) {

	path := filepath.Join("testdata", "chpl_criteria.json")
	okResponse, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tc := th.NewTestClientWithResponse(okResponse)

	return tc, nil
}
