// +build integration

package chplquerier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/pkg/errors"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/viper"
)

var store *postgresql.Store

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

func Test_Integration_GetCHPLProducts(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	var actualProdsStored int

	ctx := context.Background()
	client := &http.Client{
		Timeout: time.Second * 35,
	}

	// as of 12/5/19, at least 7676 entries are expected to be added to the database
	minNumExpProdsStored := 7676

	err = GetCHPLProducts(ctx, store, client)
	th.Assert(t, err == nil, err)
	rows := store.DB.QueryRow("SELECT COUNT(*) FROM healthit_products;")
	err = rows.Scan(&actualProdsStored)
	th.Assert(t, err == nil, err)
	th.Assert(t, actualProdsStored >= minNumExpProdsStored, fmt.Sprintf("Expected at least %d products stored. Actually had %d products stored.", minNumExpProdsStored, actualProdsStored))

	// expect to see this entry in the DB:
	// {
	// 	"id": 7849,
	// 	"chplProductNumber": "15.04.04.2657.Care.01.00.0.160701",
	// 	"edition": "2015",
	// 	"developer": "Carefluence",
	// 	"product": "Carefluence Open API",
	// 	"version": "1",
	// 	"certificationDate": 1467331200000,
	// 	"certificationStatus": "Active",
	// 	"criteriaMet": "170.315 (d)(1)☺170.315 (d)(10)☺170.315 (d)(9)☺170.315 (g)(4)☺170.315 (g)(5)☺170.315 (g)(6)☺170.315 (g)(7)☺170.315 (g)(8)☺170.315 (g)(9)",
	// 	"apiDocumentation": "170.315 (g)(7)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(8)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☺170.315 (g)(9)☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
	// }
	hitp, err := store.GetHealthITProductUsingNameAndVersion(ctx, "Carefluence Open API", "1")
	th.Assert(t, err == nil, err)
	th.Assert(t, hitp.CHPLID == "15.04.04.2657.Care.01.00.0.160701", "CHPL ID is not what was expected")
	th.Assert(t, hitp.CertificationEdition == "2015", "Certification edition is not what was expected")
	th.Assert(t, hitp.Developer == "Carefluence", "Developer is not what was expected")
	th.Assert(t, hitp.CertificationDate.Equal(time.Unix(1467331200, 0).UTC()), "Certification date is not what was expected")
	th.Assert(t, hitp.CertificationStatus == "Active", "Certification status is not what was expected")
	// TODO: Can continue to assert this format after changes described in https://oncprojectracking.healthit.gov/support/browse/LANTERN-156 are addressed
	//th.Assert(t, reflect.DeepEqual(hitp.CertificationCriteria, []string{"170.315 (d)(1)", "170.315 (d)(10)", "170.315 (d)(9)", "170.315 (g)(4)", "170.315 (g)(5)", "170.315 (g)(6)", "170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"}), "Certification criteria is not what was expected")
}

func Test_persistProduct(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	var ctx context.Context
	var cancel context.CancelFunc

	prod := testCHPLProd
	hitp := testHITP

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM healthit_products;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()

	// check that ended context when no element in store fails as expected
	ctx, cancel = context.WithCancel(context.Background())
	cancel()
	err = persistProduct(ctx, store, &prod)
	th.Assert(t, errors.Cause(err) == context.Canceled, "should have errored out with root cause that the context was canceled")

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 0, "should not have stored data")

	// reset context
	ctx = context.Background()

	// check that new item is stored
	err = persistProduct(ctx, store, &prod)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")
	storedHitp, err := store.GetHealthITProductUsingNameAndVersion(ctx, "Carefluence Open API", "1")
	th.Assert(t, err == nil, err)
	th.Assert(t, hitp.Equal(storedHitp), "stored data does not equal expected store data")

	// check that newer updated item replaces item
	prod.Edition = "2015"
	hitp.CertificationEdition = "2015"
	err = persistProduct(ctx, store, &prod)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")
	storedHitp, err = store.GetHealthITProductUsingNameAndVersion(ctx, "Carefluence Open API", "1")
	th.Assert(t, err == nil, err)
	th.Assert(t, hitp.Equal(storedHitp), "stored data does not equal expected store data")

	// check that older updated item does not replace item
	prod.Edition = "2014"
	hitp.CertificationEdition = "2015" // keeping 2015
	err = persistProduct(ctx, store, &prod)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not store data as expected")
	storedHitp, err = store.GetHealthITProductUsingNameAndVersion(ctx, "Carefluence Open API", "1")
	th.Assert(t, err == nil, err)
	th.Assert(t, hitp.Equal(storedHitp), "stored data does not equal expected store data")

	// check that malformed product throws error
	prod.APIDocumentation = "170.315 (g)(7),http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
	err = persistProduct(ctx, store, &prod)
	th.Assert(t, err != nil, "expected error parsing product")

	// check that ambiguous update throws error
	prod = testCHPLProd
	prod.Edition = "2015" // same date as what is in store
	prod.CertificationStatus = "Retired"
	err = persistProduct(ctx, store, &prod)
	th.Assert(t, err != nil, "expected error updating product")

	// check that error adding to store throws error
	prod = testCHPLProd
	prod.Product = "A new product"
	prod.ChplProductNumber = strings.Repeat("a", 510) // name too long. throw db error.
	err = persistProduct(ctx, store, &prod)
	th.Assert(t, err != nil, "expected error adding product")

	// check that error updating to store throws error
	prod = testCHPLProd
	prod.Product = "A new product"
	prod.Edition = "2016"
	prod.CertificationStatus = strings.Repeat("a", 510) // name too long. throw db error.
	err = persistProduct(ctx, store, &prod)
	th.Assert(t, err != nil, "expected error updating product")
}

func Test_persistProducts(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	ctx := context.Background()

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM healthit_products;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()

	// standard persist

	prod1 := testCHPLProd
	prod2 := testCHPLProd
	prod2.Product = "another prod"

	prodList := chplCertifiedProductList{Results: []chplCertifiedProduct{prod1, prod2}}

	err = persistProducts(ctx, store, &prodList)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not persist two products as expected")

	_, err = store.GetHealthITProductUsingNameAndVersion(ctx, prod1.Product, prod1.Version)
	th.Assert(t, err == nil, "Did not store first product as expected")
	_, err = store.GetHealthITProductUsingNameAndVersion(ctx, prod2.Product, prod2.Version)
	th.Assert(t, err == nil, "Did not store second product as expected")

	// persist with errors

	_, err = store.DB.Exec("DELETE FROM healthit_products;") // reset values
	th.Assert(t, err == nil, err)

	hook := logtest.NewGlobal()

	prod2.APIDocumentation = "170.315 (g)(7),http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
	expectedErr := "retreiving the API URL from the health IT product API documentation list failed: unexpected format for api doc string"
	prodList = chplCertifiedProductList{Results: []chplCertifiedProduct{prod1, prod2}}

	err = persistProducts(ctx, store, &prodList)
	// don't expect the function to return with errors
	th.Assert(t, err == nil, err)
	// only expect one item to be stored
	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 1, "did not persist one product as expected")

	_, err = store.GetHealthITProductUsingNameAndVersion(ctx, prod1.Product, prod1.Version)
	th.Assert(t, err == nil, "Did not store first product as expected")

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

	_, err = store.DB.Exec("DELETE FROM healthit_products;") // reset values
	th.Assert(t, err == nil, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	prod2 = testCHPLProd
	prod2.Product = "another prod"

	err = persistProducts(ctx, store, &prodList)
	th.Assert(t, errors.Cause(err) == context.Canceled, "expected persistProducts to error out due to context ending")
}

func Test_GetCHPLProducts(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	var ct int
	ctStmt, err := store.DB.Prepare("SELECT COUNT(*) FROM healthit_products;")
	th.Assert(t, err == nil, err)
	defer ctStmt.Close()

	var tc *th.TestClient
	var ctx context.Context

	// basic test

	// mock JSON includes 201 product entries, but w duplicates, the number stored is 168.
	expectedProdsStored := 168

	tc, err = basicTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx = context.Background()

	err = GetCHPLProducts(ctx, store, &(tc.Client))
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == expectedProdsStored, fmt.Sprintf("Expected %d products stored. Actually had %d products stored.", expectedProdsStored, ct))

	// test context ended
	// also checks what happens when an http request fails

	tc, err = basicTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = store.DB.Exec("DELETE FROM healthit_products;") // reset values
	th.Assert(t, err == nil, err)

	err = GetCHPLProducts(ctx, store, &(tc.Client))
	switch reqErr := errors.Cause(err).(type) {
	case *url.Error:
		th.Assert(t, reqErr.Err == context.Canceled, "Expected error stating that context was canceled")
	default:
		t.Fatal("Expected context canceled error")
	}

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

	err = GetCHPLProducts(ctx, store, &(tc.Client))
	switch errors.Cause(err).(type) {
	case *json.SyntaxError:
		// ok
	default:
		t.Fatal("Expected JSON syntax error")
	}

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
