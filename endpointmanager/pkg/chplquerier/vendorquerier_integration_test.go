// +build integration

package chplquerier

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/pkg/errors"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/viper"
)

func Test_persistVendor(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	var ctx context.Context
	var cancel context.CancelFunc

	chplVend := testCHPLVendor1
	vend := testVendor1

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM vendors;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()

	// check that ended context when no element in store fails as expected
	ctx, cancel = context.WithCancel(context.Background())
	cancel()
	err = persistVendor(ctx, store, &chplVend)
	th.Assert(t, errors.Cause(err) == context.Canceled, "should have errored out with root cause that the context was canceled")

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 0, "should not have stored data")

	// reset context
	ctx = context.Background()

	// check that new item is stored
	err = persistVendor(ctx, store, &chplVend)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")
	storedVend, err := store.GetVendorUsingName(ctx, "Epic Systems Corporation")
	th.Assert(t, err == nil, err)
	th.Assert(t, vend.Equal(storedVend), "1stored data does not equal expected store data")

	// check that a change replaces the item
	chplVend.Status = chplStatus{ID: 2, Status: "other"}
	vend.Status = "other"
	err = persistVendor(ctx, store, &chplVend)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "initial stored data does not equal expected store data")
	storedVend, err = store.GetVendorUsingName(ctx, "Epic Systems Corporation")
	th.Assert(t, err == nil, err)
	th.Assert(t, vend.Equal(storedVend), "updated stored data does not equal expected store data")

	// check that error adding to store throws error
	chplVend = testCHPLVendor1
	chplVend.Name = strings.Repeat("a", 510) // name too long. throw db error.
	chplVend.DeveloperCode = "asdf"
	chplVend.DeveloperID = 9999
	err = persistVendor(ctx, store, &chplVend)
	th.Assert(t, err != nil, "expected error adding vendor")

	// check that error updating to store throws error
	chplVend = testCHPLVendor1
	chplVend.Status = chplStatus{ID: 2, Status: strings.Repeat("a", 510)}
	err = persistVendor(ctx, store, &chplVend)
	th.Assert(t, err != nil, "expected error updating vendor")
}

func Test_persistVendors(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	ctx := context.Background()

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM vendors;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()

	// standard persist

	chplVend1 := testCHPLVendor1
	chplVend2 := testCHPLVendor2

	vendList := chplVendorList{Developers: []chplVendor{chplVend1, chplVend2}}

	err = persistVendors(ctx, store, &vendList)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not persist two vendors as expected")

	_, err = store.GetVendorUsingName(ctx, chplVend1.Name)
	th.Assert(t, err == nil, "Did not store first vendor as expected")
	_, err = store.GetVendorUsingName(ctx, chplVend2.Name)
	th.Assert(t, err == nil, "Did not store second vendor as expected")

	// persist with errors

	_, err = store.DB.Exec("DELETE FROM vendors;") // reset values
	th.Assert(t, err == nil, err)

	hook := logtest.NewGlobal()

	chplVend2.Name = strings.Repeat("a", 510)
	expectedErr := "adding vendor to store failed: pq: value too long for type character varying(500)"
	vendList = chplVendorList{Developers: []chplVendor{chplVend1, chplVend2}}

	err = persistVendors(ctx, store, &vendList)
	// don't expect the function to return with errors
	th.Assert(t, err == nil, err)
	// only expect one item to be stored
	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not persist one vendor as expected")

	_, err = store.GetVendorUsingName(ctx, chplVend1.Name)
	th.Assert(t, err == nil, "Did not store first vendor as expected")

	// expect presence of a log message
	found := false
	for i := range hook.Entries {
		if hook.Entries[i].Message == expectedErr {
			found = true
			break
		}
	}
	th.Assert(t, found, "expected an error to be logged")

	// persist when context has ended

	_, err = store.DB.Exec("DELETE FROM vendors;") // reset values
	th.Assert(t, err == nil, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	chplVend2 = testCHPLVendor2
	vendList = chplVendorList{Developers: []chplVendor{chplVend1, chplVend2}}

	err = persistVendors(ctx, store, &vendList)
	th.Assert(t, errors.Cause(err) == context.Canceled, "expected persistVendors to error out due to context ending")
}

func Test_GetCHPLVendors(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM vendors;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()

	var tc *th.TestClient
	var ctx context.Context

	apiKey := viper.GetString("chplapikey")
	viper.Set("chplapikey", "tmp_api_key")
	defer viper.Set("chplapikey", apiKey)

	// basic test

	// mock JSON includes 38 vendor entries.
	expectedVendsStored := 38

	tc, err = basicVendorTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx = context.Background()

	err = GetCHPLVendors(ctx, store, &(tc.Client), "")
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == expectedVendsStored, fmt.Sprintf("Expected %d vendors stored. Actually had %d vendors stored.", expectedVendsStored, ct))

	// test context ended
	// also checks what happens when an http request fails

	hook := logtest.NewGlobal()
	expectedErr := "Got error:\nmaking the GET request to the CHPL server failed: Get \"https://chpl.healthit.gov/rest/developers?api_key=tmp_api_key\": context canceled\n\nfrom URL: https://chpl.healthit.gov/rest/developers?api_key=tmp_api_key"

	tc, err = basicVendorTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = store.DB.Exec("DELETE FROM vendors;") // reset values
	th.Assert(t, err == nil, err)

	err = GetCHPLVendors(ctx, store, &(tc.Client), "")
	found := false
	for i := range hook.Entries {
		if strings.Contains(hook.Entries[i].Message, expectedErr) {
			found = true
			break
		}
	}
	th.Assert(t, found, "expected an error to be logged")

	// test http status != 200

	hook = logtest.NewGlobal()
	expectedErr = "CHPL request responded with status: 404 Not Found"

	tc = th.NewTestClientWith404()
	defer tc.Close()

	ctx = context.Background()

	err = GetCHPLVendors(ctx, store, &(tc.Client), "")

	// expect presence of a log message
	found = false
	for i := range hook.Entries {
		if strings.Contains(hook.Entries[i].Message, expectedErr) {
			found = true
			break
		}
	}
	th.Assert(t, found, "expected response error specifying response code")

	// test with malformed json

	malformedJSON := `
			"asdf": [
			{}]}
			`
	tc = th.NewTestClientWithResponse([]byte(malformedJSON))
	defer tc.Close()

	ctx = context.Background()

	_, err = store.DB.Exec("DELETE FROM healthit_products;") // reset values
	th.Assert(t, err == nil, err)

	err = GetCHPLVendors(ctx, store, &(tc.Client), "")
	switch errors.Cause(err).(type) {
	case *json.SyntaxError:
		// ok
	default:
		t.Fatal("Expected JSON syntax error")
	}

}
