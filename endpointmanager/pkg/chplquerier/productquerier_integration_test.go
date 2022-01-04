// +build integration

package chplquerier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/viper"
)

var store *postgresql.Store

var testVendorCHPLProd *endpointmanager.Vendor = &endpointmanager.Vendor{
	Name:          "Carefluence",
	DeveloperCode: "D",
	CHPLID:        4,
}

var testCriteria = &endpointmanager.CertificationCriteria{
	CertificationID:        30,
	CertificationNumber:    "170.315 (d)(2)",
	Title:                  "Transmission to Public Health Agencies - Syndromic Surveillance",
	CertificationEditionID: 3,
	CertificationEdition:   "2015",
	Description:            "Syndromic Surveillance",
	Removed:                false,
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
	store.AddVendor(ctx, testVendorCHPLProd) // add vendor product so we can link to it
	hitp.VendorID = testVendorCHPLProd.ID

	// add all criteria
	apiKey := viper.GetString("chplapikey")
	viper.Set("chplapikey", "tmp_api_key")
	defer viper.Set("chplapikey", apiKey)

	criteriaClient, err := basicTestCriteriaClient()
	th.Assert(t, err == nil, err)
	defer criteriaClient.Close()

	ctx = context.Background()

	err = GetCHPLCriteria(ctx, store, &(criteriaClient.Client), "")
	th.Assert(t, err == nil, err)

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
	prod.APIDocumentation = []string{",http://carefluence.com/Carefluence-OpenAPI-Documentation.html"}
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

	// test criteria linking

	// add criteria so we can link to it
	tmpCrit := testCriteria
	err = store.AddCriteria(ctx, tmpCrit)
	th.Assert(t, err == nil, "did not expect error adding criteria")

	prod = testCHPLProd
	prod.Product = "A new product for criteria testing"
	err = persistProduct(ctx, store, &prod)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not add a new element to healthit_product store")

	storedHitp, err = store.GetHealthITProductUsingNameAndVersion(ctx, "A new product for criteria testing", "1")
	th.Assert(t, err == nil, "error getting stored hitp from database")
	retProd, retCritID, retCritNum, err := store.GetProductCriteriaLink(ctx, tmpCrit.CertificationID, storedHitp.ID)
	th.Assert(t, err == nil, fmt.Errorf("link did not occur, %s", err))
	th.Assert(t, retProd == storedHitp.ID, "returned product ID is not expected value")
	th.Assert(t, retCritID == tmpCrit.CertificationID, "returned criteria ID is not expected value")
	th.Assert(t, retCritNum == tmpCrit.CertificationNumber, "returned criteria number is not expected value")

	// test critera linking update
	prod.CriteriaMet = []int{31, 32, 33, 34, 35, 36, 37, 38}
	prod.Edition = "2020"
	err = persistProduct(ctx, store, &prod)
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == 2, "did not add a new element to healthit_product store")

	storedHitp, err = store.GetHealthITProductUsingNameAndVersion(ctx, "A new product for criteria testing", "1")
	th.Assert(t, err == nil, "error getting stored hitp from database")
	_, _, _, err = store.GetProductCriteriaLink(ctx, tmpCrit.CertificationID, storedHitp.ID)
	th.Assert(t, err != nil, fmt.Errorf("Should have returned nothing since the criteria no longer exists in the product"))
}

func Test_persistProducts(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error

	ctx := context.Background()

	store.AddVendor(ctx, testVendorCHPLProd)

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

	prod2.APIDocumentation = []string{"☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☹"}
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
	th.Assert(t, found, "expected an api error to be logged")

	// persist when context has ended

	_, err = store.DB.Exec("DELETE FROM healthit_products;") // reset values
	th.Assert(t, err == nil, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	prod2 = testCHPLProd
	prod2.Product = "another prod"

	err = persistProducts(ctx, store, &prodList)
}

func Test_parseHITProd(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	ctx := context.Background()
	prod := testCHPLProd
	expectedHITProd := testHITP

	store.AddVendor(ctx, testVendorCHPLProd)
	expectedHITProd.VendorID = testVendorCHPLProd.ID

	// basic test

	hitProd, err := parseHITProd(ctx, &prod, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, hitProd.Equal(&expectedHITProd), "CHPL Product did not parse into HealthITProduct as expected.")

	// test bad url in api doc string

	prod.APIDocumentation = []string{"☹.com/Carefluence-OpenAPI-Documentation.html, ☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html, ☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"}
	_, err = parseHITProd(ctx, &prod, store)
	switch errors.Cause(err).(type) {
	case *url.Error:
		// expect url.Error because bad URL provided and we check that using the url package.
	default:
		t.Fatal(err)
	}
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

	apiKey := viper.GetString("chplapikey")
	viper.Set("chplapikey", "tmp_api_key")
	defer viper.Set("chplapikey", apiKey)

	// basic test

	// prep with vendors
	tc, err = basicVendorTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx = context.Background()

	err = GetCHPLVendors(ctx, store, &(tc.Client), "")
	th.Assert(t, err == nil, err)

	// mock JSON includes 201 product entries, but w duplicates, the number stored is 168.
	expectedProdsStored := 168

	tc, err = basicTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx = context.Background()

	err = GetCHPLProducts(ctx, store, &(tc.Client), "")
	th.Assert(t, err == nil, err)

	err = ctStmt.QueryRow().Scan(&ct)
	th.Assert(t, err == nil, err)
	th.Assert(t, ct == expectedProdsStored, fmt.Sprintf("Expected %d products stored. Actually had %d products stored.", expectedProdsStored, ct))

	// test context ended
	// also checks what happens when an http request fails

	hook := logtest.NewGlobal()
	expectedErr := "Got error:\nmaking the GET request to the CHPL server failed: Get \"https://chpl.healthit.gov/rest/search/beta?api_key=tmp_api_key&fields=id%2Cedition%2Cdeveloper%2Cproduct%2Cversion%2CchplProductNumber%2CcertificationStatus%2CcriteriaMet%2CapiDocumentation%2CcertificationDate%2CpracticeType\": context canceled"
	tc, err = basicTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = store.DB.Exec("DELETE FROM healthit_products;") // reset values
	th.Assert(t, err == nil, err)

	err = GetCHPLProducts(ctx, store, &(tc.Client), "")

	// expect presence of a log message
	found := false
	for i := range hook.Entries {
		log.Info(hook.Entries[i].Message)
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

	err = GetCHPLProducts(ctx, store, &(tc.Client), "")

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

	err = GetCHPLProducts(ctx, store, &(tc.Client), "")
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
