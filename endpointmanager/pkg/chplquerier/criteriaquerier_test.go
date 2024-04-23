package chplquerier

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/errors"
	logtest "github.com/sirupsen/logrus/hooks/test"

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

func Test_makeCHPLCriteriaURL(t *testing.T) {

	// basic test

	apiKey := viper.GetString("chplapikey")
	viper.Set("chplapikey", "tmp_api_key")
	defer viper.Set("chplapikey", apiKey)

	expected := "https://chpl.healthit.gov/rest/certification-criteria?api_key=tmp_api_key"

	actualURL, err := makeCHPLCriteriaURL()
	th.Assert(t, err == nil, err)

	actual := actualURL.String()
	th.Assert(t, expected == actual, fmt.Sprintf("Expected %s to equal %s.", actual, expected))

	// test empty api key

	viper.Set("chplapikey", "")
	_, err = makeCHPLCriteriaURL()
	th.Assert(t, err != nil, "Expected to return an error due to the api key not being set")

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

	critListJSON := `[
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
		}
	]
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

	var Results = []chplCertCriteria{expectedCrit1, expectedCrit2}

	ctx = context.Background()
	critList, err := convertCriteriaJSONToObj(ctx, []byte(critListJSON))
	th.Assert(t, err == nil, err)
	th.Assert(t, critList != nil, "Expected results field to be filled out for criteria list.")
	th.Assert(t, len(critList) == len(Results), fmt.Sprintf("Number of criteria is %d. Should be %d.", len(critList), len(Results)))

	for i, crit := range critList {
		th.Assert(t, crit == Results[i], "Expected parsed criteria to equal expected criteria.")
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

	// mock JSON includes 182 criteria entries
	expectedCriteriaReceived := 182

	tc, err = basicTestCriteriaClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx = context.Background()

	critJSON, err := getCriteriaJSON(ctx, &(tc.Client), "")
	th.Assert(t, err == nil, err)

	// convert received JSON so we can count the number of entries received
	criteria, err := convertCriteriaListJSONToObj(ctx, critJSON)
	th.Assert(t, err == nil, err)
	actualCriteriaReceived := len(criteria)
	th.Assert(t, actualCriteriaReceived == expectedCriteriaReceived, fmt.Sprintf("Expected to receive %d criteria. Actually received %d criteria.", expectedCriteriaReceived, actualCriteriaReceived))

	// test context ended.
	// also checks what happens when an http request fails

	hook := logtest.NewGlobal()
	expectedErr := "Got error:\nmaking the GET request to the CHPL server failed: Get \"https://chpl.healthit.gov/rest/certification-criteria?api_key=tmp_api_key\": context canceled"

	tc, err = basicTestCriteriaClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _ = getCriteriaJSON(ctx, &(tc.Client), "")
	// expect presence of a log message
	found := false
	for i := range hook.Entries {
		if strings.Contains(hook.Entries[i].Message, expectedErr) {
			found = true
			break
		}
	}
	th.Assert(t, found, "expected an error to be logged")

	// test http status != 200

	expectedErr = "Got error:\nCHPL request responded with status: 404 Not Found\n\nfrom URL: https://chpl.healthit.gov/rest/certification-criteria?api_key=tmp_api_key"

	tc = th.NewTestClientWith404()
	defer tc.Close()

	ctx = context.Background()

	_, _ = getCriteriaJSON(ctx, &(tc.Client), "")
	// expect presence of a log message
	found = false
	for i := range hook.Entries {
		if strings.Contains(hook.Entries[i].Message, expectedErr) {
			found = true
			break
		}
	}
	th.Assert(t, found, "expected 404 error to be logged")
	// test error on URL creation

	chplDomainOrig := chplDomain
	chplDomain = "http://%41:8080/" // invalid domain
	defer func() { chplDomain = chplDomainOrig }()

	tc, err = basicTestCriteriaClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx = context.Background()

	_, err = getCriteriaJSON(ctx, &(tc.Client), "")
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
