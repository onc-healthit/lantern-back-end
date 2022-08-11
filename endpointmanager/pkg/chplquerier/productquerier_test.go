package chplquerier

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/pkg/errors"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/viper"
)

var testCHPLProd chplCertifiedProduct = chplCertifiedProduct{
	ID:                  7849,
	ChplProductNumber:   "15.04.04.2657.Care.01.00.0.160701",
	Edition:             "2014",
	Developer:           "Carefluence",
	Product:             "Carefluence Open API",
	Version:             "1",
	CertificationDate:   1467331200000,
	CertificationStatus: "Active",
	CriteriaMet:         []int{30, 31, 32, 33, 34, 35, 36, 37, 38},
	APIDocumentation:    []string{"☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html", "☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html", "☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"},
}

var testHITP endpointmanager.HealthITProduct = endpointmanager.HealthITProduct{
	Name:                  "Carefluence Open API",
	Version:               "1",
	CertificationStatus:   "Active",
	CertificationDate:     time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
	CertificationEdition:  "2014",
	CHPLID:                "15.04.04.2657.Care.01.00.0.160701",
	CertificationCriteria: []int{30, 31, 32, 33, 34, 35, 36, 37, 38},
	APIURL:                "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
}

func Test_makeCHPLProductURL(t *testing.T) {

	// basic test

	apiKey := viper.GetString("chplapikey")
	viper.Set("chplapikey", "tmp_api_key")
	defer viper.Set("chplapikey", apiKey)

	expected := "https://chpl.healthit.gov/rest/search/beta?api_key=tmp_api_key&fields=id%2Cedition%2Cdeveloper%2Cproduct%2Cversion%2CchplProductNumber%2CcertificationStatus%2CcriteriaMet%2CapiDocumentation%2CcertificationDate%2CpracticeType&pageNumber=0&pageSize=100"
	pageSize := 100
	pageNumber := 0

	actualURL, err := makeCHPLProductURL(pageSize, pageNumber)
	th.Assert(t, err == nil, err)

	actual := actualURL.String()
	th.Assert(t, expected == actual, fmt.Sprintf("Expected %s to equal %s.", actual, expected))

	// test empty api key

	viper.Set("chplapikey", "")
	actualURL, err = makeCHPLProductURL(pageSize, pageNumber)
	th.Assert(t, err != nil, "Expected to return an error due to the api key not being set")
	th.Assert(t, actualURL == nil, "Expected chpl product URL to be nil due to api key not being set")

	// test invalid domain and error handling

	chplDomainOrig := chplDomain
	chplDomain = "http://%41:8080/" // invalid domain
	defer func() { chplDomain = chplDomainOrig }()

	_, err = makeCHPLProductURL(pageSize, pageNumber)
	switch errors.Cause(err).(type) {
	case *url.Error:
		// ok
	default:
		t.Fatal("Expected url error")
	}
}

func Test_convertProductJSONToObj(t *testing.T) {
	var ctx context.Context
	var err error

	// basic test

	prodListJSON := `{
		"results": [
		{
			"id": 7849,
			"chplProductNumber": "15.04.04.2657.Care.01.00.0.160701",
			"edition": "2014",
			"developer": "Carefluence",
			"product": "Carefluence Open API",
			"version": "1",
			"certificationDate": 1467331200000,
			"certificationStatus": "Active",
			"criteriaMet": [30, 31, 32, 33, 34, 35, 36, 37, 38],
			"apiDocumentation": ["☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html", "☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html", "☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"]
		},
		{
			"id": 7850,
			"chplProductNumber": "15.04.04.2657.Care.01.00.0.160703",
			"edition": "2014",
			"developer": "Carefluence",
			"product": "Carefluence Open API",
			"version": "0.3",
			"certificationDate": 1467320000000,
			"certificationStatus": "Active",
			"criteriaMet": [30, 31, 32, 33, 34, 35, 36, 37, 38],
			"apiDocumentation": ["☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html", "☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html", "☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"]
		}]}
		`

	expectedProd1 := testCHPLProd

	expectedProd2 := chplCertifiedProduct{
		ID:                  7850,
		ChplProductNumber:   "15.04.04.2657.Care.01.00.0.160703",
		Edition:             "2014",
		Developer:           "Carefluence",
		Product:             "Carefluence Open API",
		Version:             "0.3",
		CertificationDate:   1467320000000,
		CertificationStatus: "Active",
		CriteriaMet:         []int{30, 31, 32, 33, 34, 35, 36, 37, 38},
		APIDocumentation:    []string{"☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html", "☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html", "☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"},
	}

	expectedProdList := chplCertifiedProductList{
		Results: []chplCertifiedProduct{expectedProd1, expectedProd2},
	}

	ctx = context.Background()
	prodList, err := convertProductJSONToObj(ctx, []byte(prodListJSON))
	th.Assert(t, err == nil, err)
	th.Assert(t, prodList.Results != nil, "Expected results field to be filled out for  product list.")
	th.Assert(t, len(prodList.Results) == len(expectedProdList.Results), fmt.Sprintf("Number of products is %d. Should be %d.", len(prodList.Results), len(expectedProdList.Results)))

	for i, prod := range prodList.Results {
		th.Assert(t, reflect.DeepEqual(prod, expectedProdList.Results[i]), "Expected parsed products to equal expected products.")
	}

	// test with canceled context

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = convertProductJSONToObj(ctx, []byte(prodListJSON))
	th.Assert(t, errors.Cause(err) == context.Canceled, "Expected canceled context error")

	// test with malformed JSON

	ctx = context.Background()
	malformedJSON := `
		"asdf": [
		{}]}
		`

	_, err = convertProductJSONToObj(ctx, []byte(malformedJSON))
	switch errors.Cause(err).(type) {
	case *json.SyntaxError:
		// ok
	default:
		t.Fatal("Expected JSON syntax error")
	}
}

func Test_getAPIURL(t *testing.T) {

	// basic test

	apiDocArray := []string{"☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html", "☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html", "☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html"}
	expectedURL := "http://carefluence.com/Carefluence-OpenAPI-Documentation.html"

	actualURL, err := getAPIURL(apiDocArray)
	th.Assert(t, err == nil, err)
	th.Assert(t, expectedURL == actualURL, fmt.Sprintf("Expected '%s'. Got '%s'.", expectedURL, actualURL))

	// provide bad string - unexpected delimeter

	apiDocArray = []string{"☹http://carefluence.com/Carefluence-OpenAPI-Documentation.html☹"}

	_, err = getAPIURL(apiDocArray)
	th.Assert(t, err != nil, "Expected error due to malformed api doc string")

	// provide empty array

	apiDocArray = []string{}
	expectedURL = ""

	actualURL, err = getAPIURL(apiDocArray)
	th.Assert(t, err == nil, err)
	th.Assert(t, expectedURL == actualURL, "Expected an empty string")
}

func Test_prodNeedsUpdate(t *testing.T) {

	type expectedResult struct {
		name        string
		hitProd     endpointmanager.HealthITProduct
		needsUpdate bool
		err         error
	}

	expectedResults := []expectedResult{}

	base := testHITP

	same := testHITP
	expectedResults = append(expectedResults, expectedResult{name: "same", hitProd: same, needsUpdate: false, err: nil})

	badEdition := testHITP
	badEdition.CertificationEdition = "asdf"
	expectedResults = append(expectedResults, expectedResult{name: "badEdition", hitProd: badEdition, needsUpdate: false, err: errors.New(`strconv.Atoi: parsing "asdf": invalid syntax`)})

	editionAfter := testHITP
	editionAfter.CertificationEdition = "2015"
	expectedResults = append(expectedResults, expectedResult{name: "editionAfter", hitProd: editionAfter, needsUpdate: true, err: nil})

	dateAfter := testHITP
	dateAfter.CertificationDate = time.Date(2016, 9, 1, 0, 0, 0, 0, time.UTC)
	expectedResults = append(expectedResults, expectedResult{name: "dateAfter", hitProd: dateAfter, needsUpdate: true, err: nil})

	editionBefore := testHITP
	editionBefore.CertificationEdition = "2011"
	expectedResults = append(expectedResults, expectedResult{name: "editionBefore", hitProd: editionBefore, needsUpdate: false, err: nil})

	dateBefore := testHITP
	dateBefore.CertificationDate = time.Date(2016, 5, 1, 0, 0, 0, 0, time.UTC)
	expectedResults = append(expectedResults, expectedResult{name: "dateBefore", hitProd: dateBefore, needsUpdate: false, err: nil})

	critListShorter := testHITP
	critListShorter.CertificationCriteria = []int{30, 31, 32, 33, 34, 35, 36, 37}
	expectedResults = append(expectedResults, expectedResult{name: "critListShorter", hitProd: critListShorter, needsUpdate: false, err: fmt.Errorf("HealthITProducts certification edition and date are equal; unknown precendence for updates; not performing update: %s:%s to %s:%s", testHITP.Name, testHITP.CHPLID, testHITP.Name, testHITP.CHPLID)})

	chplID := testHITP
	chplID.CHPLID = "15.04.04.2657.Care.01.00.0.160733"
	expectedResults = append(expectedResults, expectedResult{name: "chplID", hitProd: chplID, needsUpdate: false, err: nil})

	certStatus := testHITP
	certStatus.CertificationStatus = "Retired"
	expectedResults = append(expectedResults, expectedResult{name: "certStatus", hitProd: certStatus, needsUpdate: false, err: fmt.Errorf("HealthITProducts certification edition and date are equal; unknown precendence for updates; not performing update: %s:%s to %s:%s", testHITP.Name, testHITP.CHPLID, testHITP.Name, testHITP.CHPLID)})

	for _, expRes := range expectedResults {
		needsUpdate, err := prodNeedsUpdate(&base, &(expRes.hitProd))
		th.Assert(t, needsUpdate == expRes.needsUpdate, fmt.Sprintf("For 'prodNeedsUpdate' using %s, expected %t and got %t.", expRes.name, expRes.needsUpdate, needsUpdate))
		if err != nil && expRes.err == nil {
			t.Fatalf("For 'prodNeedsUpdate' using %s, did not expect error but got error\n%v", expRes.name, err)
		}
		if err == nil && expRes.err != nil {
			t.Fatalf("For 'prodNeedsUpdate' using %s, did not receive error but expected error\n%v", expRes.name, expRes.err)
		}
		if err != nil && expRes.err != nil {
			origErr := errors.Cause(err)
			if origErr.Error() != expRes.err.Error() {
				t.Fatalf("For 'prodNeedsUpdate' using %s, expected error\n%v\nAnd got error\n%v", expRes.name, expRes.err, origErr)
			}
		}
	}

	baseBadEdition := testHITP
	baseBadEdition.CertificationEdition = "asdf"
	name := "baseBadEdition"
	expectedNeedsUpdate := false
	expectedErrorStr := `strconv.Atoi: parsing "asdf": invalid syntax`

	needsUpdate, err := prodNeedsUpdate(&baseBadEdition, &base)
	th.Assert(t, needsUpdate == expectedNeedsUpdate, fmt.Sprintf("For 'prodNeedsUpdate' using %s, expected %t and got %t.", name, expectedNeedsUpdate, needsUpdate))
	th.Assert(t, err != nil, "Expected an error")
	origErr := errors.Cause(err)
	th.Assert(t, origErr.Error() == expectedErrorStr, fmt.Sprintf("For 'prodNeedsUpdate' using %s, expected error\n%v\nAnd got error\n%v", name, expectedErrorStr, origErr))
}

func Test_getProductJSON(t *testing.T) {
	var err error
	var tc *th.TestClient
	var ctx context.Context

	apiKey := viper.GetString("chplapikey")
	viper.Set("chplapikey", "tmp_api_key")
	defer viper.Set("chplapikey", apiKey)

	// basic test

	// mock JSON includes 100 product entries
	expectedProdsReceived := 100

	tc, err = basicTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx = context.Background()
	pageSize := 100
	pageNumber := 0

	prodJSON, err := getProductJSON(ctx, &(tc.Client), "", pageSize, pageNumber)
	th.Assert(t, err == nil, err)

	// convert received JSON so we can count the number of entries received
	prods, err := convertProductJSONToObj(ctx, prodJSON)
	th.Assert(t, err == nil, prodJSON)
	actualProdsReceived := len(prods.Results)
	th.Assert(t, actualProdsReceived == expectedProdsReceived, fmt.Sprintf("Expected to receive %d products Actually received %d products.", expectedProdsReceived, actualProdsReceived))

	// test context ended.

	hook := logtest.NewGlobal()
	expectedErr := "Got error:\nmaking the GET request to the CHPL server failed:"

	tc, err = basicTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _ = getProductJSON(ctx, &(tc.Client), "", pageSize, pageNumber)
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

	expectedErr = "Got error:\nCHPL request responded with status: 404 Not Found\n\nfrom URL: https://chpl.healthit.gov/rest/search/beta?api_key=tmp_api_key&fields=id%2Cedition%2Cdeveloper%2Cproduct%2Cversion%2CchplProductNumber%2CcertificationStatus%2CcriteriaMet%2CapiDocumentation%2CcertificationDate%2CpracticeType"

	tc = th.NewTestClientWith404()
	defer tc.Close()

	ctx = context.Background()

	_, _ = getProductJSON(ctx, &(tc.Client), "", pageSize, pageNumber)
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

	tc, err = basicTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx = context.Background()

	_, err = getProductJSON(ctx, &(tc.Client), "", pageSize, pageNumber)
	switch errors.Cause(err).(type) {
	case *url.Error:
		// ok
	default:
		t.Fatal("Expected url error")
	}
}

func basicTestClient() (*th.TestClient, error) {

	path := filepath.Join("testdata", "chpl_certified_products.json")
	okResponse, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tc := th.NewTestClientWithResponse(okResponse)

	return tc, nil
}
