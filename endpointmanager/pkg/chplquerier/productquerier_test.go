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

var criteriaMetArr []criteriaMet = []criteriaMet{
	{
		Id:     30,
		Number: "170.315 (d)(10)",
		Title:  "Auditing Actions on Health Information",
	},
	{
		Id:     31,
		Number: "170.315 (g)(5)",
		Title:  "Accessibility-Centered Design",
	},
	{
		Id:     32,
		Number: "170.315 (d)(9)",
		Title:  "Trusted Connection",
	},
	{
		Id:     33,
		Number: "170.315 (d)(1)",
		Title:  "Authentication, Access Control, Authorization",
	},
	{
		Id:     34,
		Number: "170.315 (g)(4)",
		Title:  "Quality Management System",
	},
	{
		Id:     35,
		Number: "170.315 (g)(8)",
		Title:  "Application Access - Data Category Request",
	},
	{
		Id:     36,
		Number: "170.315 (g)(6)",
		Title:  "Consolidated CDA Creation",
	},
	{
		Id:     37,
		Number: "170.315 (g)(7)",
		Title:  "Application Access - Patient Selection",
	},
	{
		Id:     38,
		Number: "170.315 (g)(9)",
		Title:  "Application Access - All Data Request",
	},
}

var apiDocArr []apiDocumentation = []apiDocumentation{
	{
		Criterion: criteriaMet{
			Id:     58,
			Number: "170.315 (g)(9)",
			Title:  "Application Access - All Data Request",
		},
		Value: "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	},
	{
		Criterion: criteriaMet{
			Id:     57,
			Number: "170.315 (g)(8)",
			Title:  "Application Access - Data Category Request",
		},
		Value: "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	},
	{
		Criterion: criteriaMet{
			Id:     56,
			Number: "170.315 (g)(7)",
			Title:  "Application Access - Patient Selection",
		},
		Value: "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	},
}

var testCHPLProd chplCertifiedProduct = chplCertifiedProduct{
	ID:                  7849,
	ChplProductNumber:   "15.04.04.2657.Care.01.00.0.160701",
	Edition:             details{Id: 1, Name: "2014"},
	Developer:           details{Id: 1, Name: "Carefluence"},
	Product:             details{Id: 1, Name: "Carefluence Open API"},
	Version:             details{Id: 1, Name: "1"},
	CertificationDate:   "2016-07-01",
	CertificationStatus: details{Id: 1, Name: "Active"},
	CriteriaMet:         criteriaMetArr,
	APIDocumentation:    apiDocArr,
	PracticeType:        details{Id: 1, Name: "Inpatient"},
	ACB:				 "SLI Compliance",

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
	PracticeType:          "Inpatient",
	ACB:                   "SLI Compliance",
}

func Test_makeCHPLProductURL(t *testing.T) {

	// basic test

	apiKey := viper.GetString("chplapikey")
	viper.Set("chplapikey", "tmp_api_key")
	defer viper.Set("chplapikey", apiKey)

	expected := "https://chpl.healthit.gov/rest/search/v2?api_key=tmp_api_key&pageNumber=0&pageSize=100"
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
			"edition": {"name": "2014", "id": 1},
			"developer": {"name": "Carefluence", "id": 1},
			"product": {"name": "Carefluence Open API", "id": 1},
			"version": {"name": "1", "id": 1},
			"certificationDate": "2016-07-01",
			"certificationStatus": {"name": "Active", "id": 1},
			"practiceType": {"name": "Inpatient", "id":1},
			"acb": "SLI Compliance",
			"criteriaMet": [
                {
                    "id": 30,
                    "number": "170.315 (d)(10)",
                    "title": "Auditing Actions on Health Information"
                },
                {
                    "id": 31,
                    "number": "170.315 (g)(5)",
                    "title": "Accessibility-Centered Design"
                },
                {
                    "id": 32,
                    "number": "170.315 (d)(9)",
                    "title": "Trusted Connection"
                },
                {
                    "id": 33,
                    "number": "170.315 (d)(1)",
                    "title": "Authentication, Access Control, Authorization"
                },
                {
                    "id": 34,
                    "number": "170.315 (g)(4)",
                    "title": "Quality Management System"
                },
                {
                    "id": 35,
                    "number": "170.315 (g)(8)",
                    "title": "Application Access - Data Category Request"
                },
                {
                    "id": 36,
                    "number": "170.315 (g)(6)",
                    "title": "Consolidated CDA Creation"
                },
                {
                    "id": 37,
                    "number": "170.315 (g)(7)",
                    "title": "Application Access - Patient Selection"
                },
                {
                    "id": 38,
                    "number": "170.315 (g)(9)",
                    "title": "Application Access - All Data Request"
                }
            ],
			"apiDocumentation": [
                {
                    "criterion": {
                        "id": 58,
                        "number": "170.315 (g)(9)",
                        "title": "Application Access - All Data Request"
                    },
                    "value": "http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
                },
                {
                    "criterion": {
                        "id": 57,
                        "number": "170.315 (g)(8)",
                        "title": "Application Access - Data Category Request"
                    },
                    "value": "http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
                },
                {
                    "criterion": {
                        "id": 56,
                        "number": "170.315 (g)(7)",
                        "title": "Application Access - Patient Selection"
                    },
                    "value": "http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
                }
            ]
		},
		{
			"id": 7850,
			"chplProductNumber": "15.04.04.2657.Care.01.00.0.160703",
			"edition": {"name": "2014", "id": 1},
			"developer": {"name": "Carefluence", "id": 1},
			"product": {"name": "Carefluence Open API", "id": 1},
			"version": {"name": "0.3", "id": 1},
			"certificationDate": "2016-10-01",
			"certificationStatus": {"name": "Active", "id": 1},
			"practiceType": {"name": "Inpatient", "id":1},
			"criteriaMet": [
                {
                    "id": 30,
                    "number": "170.315 (d)(10)",
                    "title": "Auditing Actions on Health Information"
                },
                {
                    "id": 31,
                    "number": "170.315 (g)(5)",
                    "title": "Accessibility-Centered Design"
                },
                {
                    "id": 32,
                    "number": "170.315 (d)(9)",
                    "title": "Trusted Connection"
                },
                {
                    "id": 33,
                    "number": "170.315 (d)(1)",
                    "title": "Authentication, Access Control, Authorization"
                },
                {
                    "id": 34,
                    "number": "170.315 (g)(4)",
                    "title": "Quality Management System"
                },
                {
                    "id": 35,
                    "number": "170.315 (g)(8)",
                    "title": "Application Access - Data Category Request"
                },
                {
                    "id": 36,
                    "number": "170.315 (g)(6)",
                    "title": "Consolidated CDA Creation"
                },
                {
                    "id": 37,
                    "number": "170.315 (g)(7)",
                    "title": "Application Access - Patient Selection"
                },
                {
                    "id": 38,
                    "number": "170.315 (g)(9)",
                    "title": "Application Access - All Data Request"
                }
            ],
			"apiDocumentation": [
                {
                    "criterion": {
                        "id": 58,
                        "number": "170.315 (g)(9)",
                        "title": "Application Access - All Data Request"
                    },
                    "value": "http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
                },
                {
                    "criterion": {
                        "id": 57,
                        "number": "170.315 (g)(8)",
                        "title": "Application Access - Data Category Request"
                    },
                    "value": "http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
                },
                {
                    "criterion": {
                        "id": 56,
                        "number": "170.315 (g)(7)",
                        "title": "Application Access - Patient Selection"
                    },
                    "value": "http://carefluence.com/Carefluence-OpenAPI-Documentation.html"
                }
            ]
		}]}
		`

	expectedProd1 := testCHPLProd

	expectedProd2 := chplCertifiedProduct{
		ID:                  7850,
		ChplProductNumber:   "15.04.04.2657.Care.01.00.0.160703",
		Edition:             details{Id: 1, Name: "2014"},
		Developer:           details{Id: 1, Name: "Carefluence"},
		Product:             details{Id: 1, Name: "Carefluence Open API"},
		Version:             details{Id: 1, Name: "0.3"},
		CertificationDate:   "2016-10-01",
		CertificationStatus: details{Id: 1, Name: "Active"},
		CriteriaMet:         criteriaMetArr,
		APIDocumentation:    apiDocArr,
		PracticeType:        details{Id: 1, Name: "Inpatient"},
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

	apiDocArray := []apiDocumentation{
		{
			Criterion: criteriaMet{
				Id:     1,
				Number: "170.315 (g)(7)",
				Title:  "Application Access - Patient Selection",
			},
			Value: "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
		},
		{
			Criterion: criteriaMet{
				Id:     1,
				Number: "170.315 (g)(7)",
				Title:  "Application Access - Patient Selection",
			},
			Value: "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
		},
		{
			Criterion: criteriaMet{
				Id:     1,
				Number: "170.315 (g)(7)",
				Title:  "Application Access - Patient Selection",
			},
			Value: "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
		},
	}
	expectedURL := "http://carefluence.com/Carefluence-OpenAPI-Documentation.html"

	actualURL, err := getAPIURL(apiDocArray)
	th.Assert(t, err == nil, err)
	th.Assert(t, expectedURL == actualURL, fmt.Sprintf("Expected '%s'. Got '%s'.", expectedURL, actualURL))

	// provide bad string - invalid URL

	apiDocArray = []apiDocumentation{
		{
			Criterion: criteriaMet{
				Id:     1,
				Number: "170.315 (g)(7)",
				Title:  "Application Access - Patient Selection",
			},
			Value: ".com/Carefluence-OpenAPI-Documentation.html",
		},
	}

	_, err = getAPIURL(apiDocArray)
	th.Assert(t, err != nil, "Expected error since the URL in the health IT product API documentation string is not valid")

	// provide empty array

	apiDocArray = []apiDocumentation{}
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
	expectedResults = append(expectedResults, expectedResult{name: "critListShorter", hitProd: critListShorter, needsUpdate: false, err: nil})

	critListLonger := testHITP
	critListLonger.CertificationCriteria = []int{30, 31, 32, 33, 34, 35, 36, 37, 38, 40}
	expectedResults = append(expectedResults, expectedResult{name: "critListLonger", hitProd: critListLonger, needsUpdate: true, err: nil})

	critListDif := testHITP
	critListDif.CertificationCriteria = []int{30, 31, 32, 33, 34, 35, 36, 37, 40}
	expectedResults = append(expectedResults, expectedResult{name: "critListDifference", hitProd: critListDif, needsUpdate: false, err: fmt.Errorf("HealthITProducts certification criteria have the same length but are not equal; not performing update: %s:%s to %s:%s", testHITP.Name, testHITP.CHPLID, testHITP.Name, testHITP.CHPLID)})

	chplID := testHITP
	chplID.CHPLID = "15.04.04.2657.Care.01.00.0.160733"
	expectedResults = append(expectedResults, expectedResult{name: "chplID", hitProd: chplID, needsUpdate: false, err: nil})

	certStatus := testHITP
	certStatus.CertificationStatus = "Retired"
	expectedResults = append(expectedResults, expectedResult{name: "certStatus", hitProd: certStatus, needsUpdate: true, err: nil})

	practiceType := testHITP
	practiceType.PracticeType = "Ambulatory"
	expectedResults = append(expectedResults, expectedResult{name: "practiceType", hitProd: practiceType, needsUpdate: false, err: nil})

	vendorIDChange := testHITP
	vendorIDChange.VendorID = -1
	expectedResults = append(expectedResults, expectedResult{name: "vendorID", hitProd: vendorIDChange, needsUpdate: true, err: nil})

	apiURLChange := testHITP
	apiURLChange.APIURL = "http:/newapiURL.html"
	expectedResults = append(expectedResults, expectedResult{name: "apiURL", hitProd: apiURLChange, needsUpdate: true, err: nil})


	acbChange := testHITP
	acbChange.ACB = "Drummond Group"
	expectedResults = append(expectedResults, expectedResult{name: "acbChange", hitProd: acbChange, needsUpdate: true, err: nil})

	acbChangeNoValue := testHITP
	acbChangeNoValue.ACB = ""
	expectedResults = append(expectedResults, expectedResult{name: "acbNoValue", hitProd: acbChangeNoValue, needsUpdate: false, err: nil})

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

	expectedErr = "Got error:\nCHPL request responded with status: 404 Not Found\n\nfrom URL: https://chpl.healthit.gov/rest/search/v2?api_key=tmp_api_key&pageNumber=0&pageSize=100"

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
