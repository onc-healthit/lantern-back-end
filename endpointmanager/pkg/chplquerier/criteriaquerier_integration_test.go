// +build integration

package chplquerier

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/pkg/errors"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/viper"
)

var testCrit = endpointmanager.CertificationCriteria{
	CertificationID:        44,
	CertificationNumber:    "170.315 (f)(2)",
	Title:                  "Transmission to Public Health Agencies - Syndromic Surveillance",
	CertificationEditionID: 3,
	CertificationEdition:   "2015",
	Description:            "Syndromic Surveillance",
	Removed:                false,
}

func Test_persistCriteria(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	var ctx context.Context
	var cancel context.CancelFunc

	criteria := testCHPLCrit
	crit := testCrit

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM certification_criteria;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()

	// check that ended context when no element in store fails as expected
	ctx, cancel = context.WithCancel(context.Background())
	cancel()
	err = persistCriteria(ctx, store, &criteria)
	th.Assert(t, errors.Cause(err) == context.Canceled, "should have errored out with root cause that the context was canceled")

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 0, "should not have stored data")

	// reset context
	ctx = context.Background()

	err = persistCriteria(ctx, store, &criteria)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")
	storedCrit, err := store.GetCriteriaByCertificationID(ctx, 44)
	th.Assert(t, err == nil, err)
	th.Assert(t, crit.Equal(storedCrit), fmt.Errorf("stored data does not equal expected store data,stored: %+v,expected: %+v", storedCrit, crit))

	// check value is updated in db
	criteria = testCHPLCrit
	criteria.Title = "new title"
	err = persistCriteria(ctx, store, &criteria)
	th.Assert(t, err == nil, "did not expect error adding criteria")

	expectedCrit := testCrit
	expectedCrit.Title = "new title"
	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not update already stored data")
	storedCrit, err = store.GetCriteriaByCertificationID(ctx, 44)
	th.Assert(t, err == nil, err)
	th.Assert(t, expectedCrit.Equal(storedCrit), fmt.Errorf("stored data does not equal expected store data,stored: %+v,expected: %+v", storedCrit, expectedCrit))

	// check that error adding to store throws error
	criteria = testCHPLCrit
	criteria.ID = 45
	criteria.Number = strings.Repeat("a", 510) // name too long. throw db error.
	err = persistCriteria(ctx, store, &criteria)
	th.Assert(t, err != nil, "expected error adding criteria")

	// check that error updating to store throws error
	criteria = testCHPLCrit
	criteria.CertificationEdition = strings.Repeat("a", 510) // name too long. throw db error.
	err = persistCriteria(ctx, store, &criteria)
	th.Assert(t, err != nil, "expected error updating criteria")
}

func Test_persistCriterias(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	ctx := context.Background()

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM certification_criteria;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()

	// standard persist

	crit1 := testCHPLCrit
	crit2 := testCHPLCrit
	crit2.ID = 46

	critList := chplCertifiedCriteriaList{Results: []chplCertCriteria{crit1, crit2}}

	err = persistCriterias(ctx, store, &critList)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not persist two criteria as expected")

	_, err = store.GetCriteriaByCertificationID(ctx, crit1.ID)
	th.Assert(t, err == nil, "Did not store first criteria as expected")
	_, err = store.GetCriteriaByCertificationID(ctx, crit2.ID)
	th.Assert(t, err == nil, "Did not store second criteria as expected")

	// persist when context has ended

	_, err = store.DB.Exec("DELETE FROM certification_criteria;") // reset values
	th.Assert(t, err == nil, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	crit2 = testCHPLCrit
	crit2.ID = 46

	err = persistCriterias(ctx, store, &critList)
	th.Assert(t, errors.Cause(err) == context.Canceled, "expected persistCriterias to error out due to context ending")
}

func Test_parseHITCriteria(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	ctx := context.Background()
	crit := testCHPLCrit
	expectedCrit := testCrit

	// basic test

	hitCrit, err := parseHITCriteria(ctx, &crit, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, hitCrit.Equal(&expectedCrit), "CHPL Criteria did not parse into CertifcationCriteria as expected.")
}

func Test_GetCHPLCriteria(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM certification_criteria;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()

	var tc *th.TestClient
	var ctx context.Context

	apiKey := viper.GetString("chplapikey")
	viper.Set("chplapikey", "tmp_api_key")
	defer viper.Set("chplapikey", apiKey)

	// basic test

	// mock JSON includes 182 criteria entries
	expectedCriteriaStored := 182

	tc, err = basicTestCriteriaClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx = context.Background()

	err = GetCHPLCriteria(ctx, store, &(tc.Client))
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == expectedCriteriaStored, fmt.Sprintf("Expected %d criteria stored. Actually had %d criteria stored.", expectedCriteriaStored, ct))

	// test context ended
	// also checks what happens when an http request fails

	hook := logtest.NewGlobal()
	expectedErr := "Got error:\nmaking the GET request to the CHPL server failed: Get \"https://chpl.healthit.gov/rest/data/certification-criteria?api_key=tmp_api_key\": context canceled"

	tc, err = basicTestCriteriaClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = store.DB.Exec("DELETE FROM certification_criteria;") // reset values
	th.Assert(t, err == nil, err)

	err = GetCHPLCriteria(ctx, store, &(tc.Client))

	// expect presence of a log message
	found := false
	for i := range hook.Entries {
		if strings.Contains(hook.Entries[i].Message, expectedErr) {
			found = true
			break
		}
	}
	th.Assert(t, found, "expected an error to be logged from context")

	// test with malformed json

	malformedJSON := `
			"asdf": [
			{}]}
			`
	tc = th.NewTestClientWithResponse([]byte(malformedJSON))
	defer tc.Close()

	ctx = context.Background()

	_, err = store.DB.Exec("DELETE FROM certification_criteria;") // reset values
	th.Assert(t, err == nil, err)

	err = GetCHPLCriteria(ctx, store, &(tc.Client))
	switch errors.Cause(err).(type) {
	case *json.SyntaxError:
		// ok
	default:
		t.Fatal("Expected JSON syntax error")
	}

}
