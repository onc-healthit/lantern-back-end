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
	"time"

	"github.com/pkg/errors"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	logtest "github.com/sirupsen/logrus/hooks/test"

	"github.com/spf13/viper"
)

var testCHPLVendor1 chplVendor = chplVendor{
	DeveloperID:   448,
	DeveloperCode: "1447",
	Name:          "Epic Systems Corporation",
	Website:       "http://www.epic.com",
	Address: chplAddress{
		AddressID: 844,
		Line1:     "1979 Milky Way",
		Line2:     nil,
		City:      "Verona,",
		State:     "WI",
		Zipcode:   "53593",
		Country:   "USA",
	},
	LastModifiedDate: "1582575046948",
	Status: chplStatus{
		ID:     1,
		Status: "Active",
	},
}

var testCHPLVendor2 chplVendor = chplVendor{
	DeveloperID:   222,
	DeveloperCode: "1221",
	Name:          "Cerner Corporation",
	Website:       "http://www.cerner.com",
	Address: chplAddress{
		AddressID: 873,
		Line1:     "2800 Rockcreek Parkway",
		Line2:     nil,
		City:      "Kansas City,",
		State:     "MO",
		Zipcode:   "64117",
		Country:   "USA",
	},
	LastModifiedDate: "1585158101131",
	Status: chplStatus{
		ID:     1,
		Status: "Active",
	},
}

var testVendor1 endpointmanager.Vendor = endpointmanager.Vendor{
	Name:          "Epic Systems Corporation",
	DeveloperCode: "1447",
	CHPLID:        448,
	URL:           "http://www.epic.com",
	Location: &endpointmanager.Location{
		Address1: "1979 Milky Way",
		City:     "Verona,",
		State:    "WI",
		ZipCode:  "53593"},
	Status:             "Active",
	LastModifiedInCHPL: time.Date(2020, time.February, 24, 20, 10, 46, 0, time.UTC),
}

func Test_makeVendorURL(t *testing.T) {

	// basic test

	apiKey := viper.GetString("chplapikey")
	viper.Set("chplapikey", "tmp_api_key")
	defer viper.Set("chplapikey", apiKey)

	expected := "https://chpl.healthit.gov/rest/developers?api_key=tmp_api_key"

	actualURL, err := makeCHPLVendorURL()
	th.Assert(t, err == nil, err)

	actual := actualURL.String()
	th.Assert(t, expected == actual, fmt.Sprintf("Expected %s to equal %s.", actual, expected))

	// test empty api key

	viper.Set("chplapikey", "")
	actualURL, err = makeCHPLVendorURL()
	th.Assert(t, err != nil, fmt.Sprintf("Expected to return an error due to the api key not being set"))
	th.Assert(t, actualURL == nil, fmt.Sprintf("Expected chpl vendor URL to be nil due to api key not being set"))

	// test invalid domain and error handling

	chplDomainOrig := chplDomain
	chplDomain = "http://%41:8080/" // invalid domain
	defer func() { chplDomain = chplDomainOrig }()

	_, err = makeCHPLVendorURL()
	switch errors.Cause(err).(type) {
	case *url.Error:
		// ok
	default:
		t.Fatal("Expected url error")
	}
}

func Test_convertVendorJSONToObj(t *testing.T) {
	var ctx context.Context
	var err error

	// basic test

	vendorListJSON := `{
		"developers": [
			{
				"developerId": 448,
				"developerCode": "1447",
				"name": "Epic Systems Corporation",
				"website": "http://www.epic.com",
				"selfDeveloper": false,
				"address": {
					"addressId": 844,
					"line1": "1979 Milky Way",
					"line2": null,
					"city": "Verona,",
					"state": "WI",
					"zipcode": "53593",
					"country": "USA"
				},
				"contact": {
					"contactId": 1002,
					"fullName": "Sasha TerMaat",
					"friendlyName": null,
					"email": "info@epic.com",
					"phoneNumber": "608-271-9000",
					"title": null
				},
				"lastModifiedDate": "1582575046948",
				"deleted": false,
				"transparencyAttestations": [
					{
						"acbId": 3,
						"acbName": "Drummond Group",
						"attestation": {
							"transparencyAttestation": "Affirmative",
							"removed": false
						}
					},
					{
						"acbId": 4,
						"acbName": "SLI Compliance",
						"attestation": null
					},
					{
						"acbId": 6,
						"acbName": "ICSA Labs",
						"attestation": {
							"transparencyAttestation": "Affirmative",
							"removed": false
						}
					},
					{
						"acbId": 2,
						"acbName": "CCHIT",
						"attestation": null
					},
					{
						"acbId": 5,
						"acbName": "Surescripts LLC",
						"attestation": null
					},
					{
						"acbId": 1,
						"acbName": "UL LLC",
						"attestation": null
					}
				],
				"statusEvents": [
					{
						"id": 1543,
						"developerId": 448,
						"status": {
							"id": 1,
							"status": "Active"
						},
						"statusDate": 1459469975038,
						"reason": null
					}
				],
				"status": {
					"id": 1,
					"status": "Active"
				}
			},
			{
				"developerId": 222,
				"developerCode": "1221",
				"name": "Cerner Corporation",
				"website": "http://www.cerner.com",
				"selfDeveloper": false,
				"address": {
					"addressId": 873,
					"line1": "2800 Rockcreek Parkway",
					"line2": null,
					"city": "Kansas City,",
					"state": "MO",
					"zipcode": "64117",
					"country": "USA"
				},
				"contact": {
					"contactId": 1139,
					"fullName": "Greg Thole",
					"friendlyName": null,
					"email": "greg.thole@cerner.com",
					"phoneNumber": "(816)201-9882",
					"title": null
				},
				"lastModifiedDate": "1585158101131",
				"deleted": false,
				"transparencyAttestations": [
					{
						"acbId": 3,
						"acbName": "Drummond Group",
						"attestation": {
							"transparencyAttestation": "Affirmative",
							"removed": false
						}
					},
					{
						"acbId": 4,
						"acbName": "SLI Compliance",
						"attestation": null
					},
					{
						"acbId": 6,
						"acbName": "ICSA Labs",
						"attestation": {
							"transparencyAttestation": "Affirmative",
							"removed": false
						}
					},
					{
						"acbId": 2,
						"acbName": "CCHIT",
						"attestation": null
					},
					{
						"acbId": 5,
						"acbName": "Surescripts LLC",
						"attestation": null
					},
					{
						"acbId": 1,
						"acbName": "UL LLC",
						"attestation": null
					}
				],
				"statusEvents": [
					{
						"id": 1496,
						"developerId": 222,
						"status": {
							"id": 1,
							"status": "Active"
						},
						"statusDate": 1459469974196,
						"reason": null
					}
				],
				"status": {
					"id": 1,
					"status": "Active"
				}
			}]}
		`

	expectedVendorList := chplVendorList{
		Developers: []chplVendor{testCHPLVendor1, testCHPLVendor2},
	}

	ctx = context.Background()
	vendorList, err := convertVendorJSONToObj(ctx, []byte(vendorListJSON))
	th.Assert(t, err == nil, err)
	th.Assert(t, vendorList.Developers != nil, "Expected developers field to be filled out for vendors.")
	th.Assert(t, len(vendorList.Developers) == len(expectedVendorList.Developers), fmt.Sprintf("Number of vendors is %d. Should be %d.", len(vendorList.Developers), len(expectedVendorList.Developers)))

	for i, vend := range vendorList.Developers {
		th.Assert(t, vend == expectedVendorList.Developers[i], "Expected parsed vendors to equal expected vendors.")
	}

	// test with canceled context

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = convertVendorJSONToObj(ctx, []byte(vendorListJSON))
	th.Assert(t, errors.Cause(err) == context.Canceled, "Expected canceled context error")

	// test with malformed JSON

	ctx = context.Background()
	malformedJSON := `
		"asdf": [
		{}]}
		`

	_, err = convertVendorJSONToObj(ctx, []byte(malformedJSON))
	switch errors.Cause(err).(type) {
	case *json.SyntaxError:
		// ok
	default:
		t.Fatal("Expected JSON syntax error")
	}
}

func Test_getVendorJSON(t *testing.T) {
	var err error
	var tc *th.TestClient
	var ctx context.Context

	// basic test

	apiKey := viper.GetString("chplapikey")
	viper.Set("chplapikey", "tmp_api_key")
	defer viper.Set("chplapikey", apiKey)

	// mock JSON includes 38 vendor entries
	expectedVendorsReceived := 38

	tc, err = basicVendorTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx = context.Background()

	vendorJSON, err := getVendorJSON(ctx, &(tc.Client), "")
	th.Assert(t, err == nil, err)

	// convert received JSON so we can count the number of entries received
	vendors, err := convertVendorJSONToObj(ctx, vendorJSON)
	th.Assert(t, err == nil, err)
	actualVendorsReceived := len(vendors.Developers)
	th.Assert(t, actualVendorsReceived == expectedVendorsReceived, fmt.Sprintf("Expected to receive %d vendors Actually received %d vendors.", expectedVendorsReceived, actualVendorsReceived))

	// test context ended.

	tc, err = basicVendorTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	hook := logtest.NewGlobal()
	expectedErr := "Got error:\nmaking the GET request to the CHPL server failed: Get https://chpl.healthit.gov/rest/developers?api_key=tmp_api_key: context canceled"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = getVendorJSON(ctx, &(tc.Client), "")
	th.Assert(t, err == nil, err)

	// expect presence of a log message
	found := false
	for i := range hook.Entries {
		if strings.Contains(hook.Entries[i].Message, expectedErr) {
			found = true
			break
		}
	}
	th.Assert(t, found, "expected a context canceled error to be logged")

	// test http status != 200

	tc = th.NewTestClientWith404()
	defer tc.Close()

	hook = logtest.NewGlobal()
	expectedErr = "CHPL request responded with status: 404 Not Found"

	ctx = context.Background()

	_, err = getVendorJSON(ctx, &(tc.Client), "")
	th.Assert(t, err == nil, err)

	// expect presence of a log message
	found = false
	for i := range hook.Entries {
		if strings.Contains(hook.Entries[i].Message, expectedErr) {
			found = true
			break
		}
	}
	th.Assert(t, found, "expected response error specifying response code")

	// test error on URL creation

	chplDomainOrig := chplDomain
	chplDomain = "http://%41:8080/" // invalid domain
	defer func() { chplDomain = chplDomainOrig }()

	tc, err = basicVendorTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	ctx = context.Background()

	_, err = getVendorJSON(ctx, &(tc.Client), "")
	switch errors.Cause(err).(type) {
	case *url.Error:
		// ok
	default:
		t.Fatal("Expected url error")
	}
}

func Test_parseVendor(t *testing.T) {
	chplVend := testCHPLVendor1
	expectedVend := testVendor1

	// basic test

	vend, err := parseVendor(&chplVend)
	th.Assert(t, err == nil, err)
	th.Assert(t, vend.Equal(&expectedVend), "CHPL Vendor did not parse into Vendor as expected.")
}

func basicVendorTestClient() (*th.TestClient, error) {

	path := filepath.Join("testdata", "chpl_vendors.json")
	okResponse, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tc := th.NewTestClientWithResponse(okResponse)

	return tc, nil
}
