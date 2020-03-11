package chplmapper

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/mock"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_MatchEndpointToVendorAndProduct(t *testing.T) {
	ctx := context.Background()
	hitpStore := mock.NewBasicMockHealthITProductStore()

	// basic test

	// capability statement
	path := filepath.Join("../testdata", "cerner_capability_dstu2.json")
	csJSON, err := ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	// endpoint
	ep := &endpointmanager.FHIREndpoint{
		URL:                   "example.com/FHIR/DSTU2",
		OrganizationName:      "Example Inc.",
		FHIRVersion:           "DSTU2",
		AuthorizationStandard: "OAuth 2.0",
		Location: &endpointmanager.Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		CapabilityStatement: cs}

	err = MatchEndpointToVendorAndProduct(ctx, ep, hitpStore)
	th.Assert(t, err == nil, err)
	th.Assert(t, ep.Vendor == "Cerner Corporation", fmt.Sprintf("expected vendor value to be 'Cerner Corporation'. Instead got %s", ep.Vendor))

	// test no match

	// capability statement
	path = filepath.Join("../testdata", "novendor_capability_dstu2.json")
	csJSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err = capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	// endpoint
	ep = &endpointmanager.FHIREndpoint{
		URL:                   "example.com/FHIR/DSTU2",
		OrganizationName:      "Example Inc.",
		FHIRVersion:           "DSTU2",
		AuthorizationStandard: "OAuth 2.0",
		Location: &endpointmanager.Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		CapabilityStatement: cs}

	err = MatchEndpointToVendorAndProduct(ctx, ep, hitpStore)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(ep.Vendor) == 0, fmt.Sprintf("expected no vendor value. Instead got %s", ep.Vendor))

	// test no capability statement

	// endpoint
	ep = &endpointmanager.FHIREndpoint{
		URL:                   "example.com/FHIR/DSTU2",
		OrganizationName:      "Example Inc.",
		FHIRVersion:           "DSTU2",
		AuthorizationStandard: "OAuth 2.0",
		Location: &endpointmanager.Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
	}
	err = MatchEndpointToVendorAndProduct(ctx, ep, hitpStore)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(ep.Vendor) == 0, fmt.Sprintf("expected no vendor value. Instead got %s", ep.Vendor))

	// test error getting match

	// access publisher field and make into a non-string value to throw error
	var csInt map[string]interface{}
	csJSON, err = cs.GetJSON()
	th.Assert(t, err == nil, err)
	err = json.Unmarshal(csJSON, &csInt)
	th.Assert(t, err == nil, err)
	csInt["publisher"] = []int{1, 2, 3} // bad format for publisher
	csJSON, err = json.Marshal(csInt)
	th.Assert(t, err == nil, err)
	cs, err = capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	// endpoint
	ep = &endpointmanager.FHIREndpoint{
		URL:                   "example.com/FHIR/DSTU2",
		OrganizationName:      "Example Inc.",
		FHIRVersion:           "DSTU2",
		AuthorizationStandard: "OAuth 2.0",
		Location: &endpointmanager.Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		CapabilityStatement: cs}

	err = MatchEndpointToVendorAndProduct(ctx, ep, hitpStore)
	th.Assert(t, err != nil, "expected an error from accessing the publisher field in the capability statment.")
	th.Assert(t, len(ep.Vendor) == 0, fmt.Sprintf("expected no vendor value. Instead got %s", ep.Vendor))
}

func Test_getVendorMatch(t *testing.T) {
	var path string
	var err error
	var expected string

	var dstu2JSON []byte
	var dstu2 capabilityparser.CapabilityStatement
	var vendor string

	ctx := context.Background()
	store := mock.NewBasicMockHealthITProductStore()

	var dstu2Int map[string]interface{}

	// cerner
	expected = "Cerner Corporation"

	path = filepath.Join("../testdata", "cerner_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// epic
	expected = "Epic Systems Corporation" // this uses the "hackMatch" capability

	path = filepath.Join("../testdata", "epic_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// test error getting hackmatch
	err = json.Unmarshal(dstu2JSON, &dstu2Int)
	th.Assert(t, err == nil, err)
	dstu2Int["copyright"] = []int{1, 2, 3} // bad format for copyright
	dstu2JSON, err = json.Marshal(dstu2Int)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err != nil, "expected error due to accessing the copyright")
	th.Assert(t, len(vendor) == 0, fmt.Sprintf("expected no vendor value. Instead got %s", vendor))

	// allscripts
	expected = "Allscripts" // the capability statement is missing the publisher

	path = filepath.Join("../testdata", "allscripts_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// meditech
	expected = "Medical Information Technology, Inc. (MEDITECH)" // the capability statement is missing the publisher

	path = filepath.Join("../testdata", "meditech_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// test error getting match

	// access publisher field and make into a non-string value to throw error
	dstu2JSON, err = dstu2.GetJSON()
	th.Assert(t, err == nil, err)
	err = json.Unmarshal(dstu2JSON, &dstu2Int)
	th.Assert(t, err == nil, err)
	dstu2Int["publisher"] = []int{1, 2, 3} // bad format for publisher
	dstu2JSON, err = json.Marshal(dstu2Int)
	th.Assert(t, err == nil, err)
	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err != nil, "expected an error from accessing the publisher field in the capability statment.")
	th.Assert(t, len(vendor) == 0, fmt.Sprintf("expected no vendor value. Instead got %s", vendor))
}

func Test_publisherMatch(t *testing.T) {
	var path string
	var err error
	var expected string

	var dstu2JSON []byte
	var dstu2 capabilityparser.CapabilityStatement
	var vendor string

	var dstu2Int map[string]interface{}

	ctx := context.Background()
	store := mock.NewBasicMockHealthITProductStore()

	vendorsRaw, err := store.GetHealthITProductDevelopers(ctx)
	th.Assert(t, err == nil, err)
	vendorsNorm := normalizeList(vendorsRaw)

	// cerner
	expected = "Cerner Corporation"

	path = filepath.Join("../testdata", "cerner_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = publisherMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// epic
	expected = "" // the capability statement is missing the publisher

	path = filepath.Join("../testdata", "epic_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = publisherMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// allscripts
	expected = "Allscripts" // the capability statement is missing the publisher

	path = filepath.Join("../testdata", "allscripts_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = publisherMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// meditech
	expected = "Medical Information Technology, Inc. (MEDITECH)" // the capability statement is missing the publisher

	path = filepath.Join("../testdata", "meditech_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = publisherMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// test error getting match

	// access publisher field and make into a non-string value to throw error
	dstu2JSON, err = dstu2.GetJSON()
	th.Assert(t, err == nil, err)
	err = json.Unmarshal(dstu2JSON, &dstu2Int)
	th.Assert(t, err == nil, err)
	dstu2Int["publisher"] = []int{1, 2, 3} // bad format for publisher
	dstu2JSON, err = json.Marshal(dstu2Int)
	th.Assert(t, err == nil, err)
	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = publisherMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err != nil, "expected an error from accessing the publisher field in the capability statment.")
	th.Assert(t, len(vendor) == 0, fmt.Sprintf("expected no vendor value. Instead got %s", vendor))
}

func Test_matchName(t *testing.T) {
	var expected string
	var actual string
	var dev string

	devList := []string{
		"Epic Systems Corporation",
		"Cerner Group", // changed for sake of test
		"Cerner Health Services, Inc.",
		"Medical Information Technology, Inc. (MEDITECH)",
		"Allscripts",
	}
	devListNorm := normalizeList(devList)

	// allscripts
	expected = "Allscripts"
	dev = normalizeName("Allscripts")
	actual = matchName(dev, devListNorm, devList)
	th.Assert(t, expected == actual, fmt.Sprintf("Expected %s. Got %s.", expected, actual))

	// meditech
	expected = "Medical Information Technology, Inc. (MEDITECH)"
	dev = normalizeName("Medical Information Technology, Inc")
	actual = matchName(dev, devListNorm, devList)
	th.Assert(t, expected == actual, fmt.Sprintf("Expected %s. Got %s.", expected, actual))

	// cerner
	expected = "Cerner Group\tCerner Health Services, Inc."
	dev = normalizeName("Cerner")
	actual = matchName(dev, devListNorm, devList)
	th.Assert(t, expected == actual, fmt.Sprintf("Expected %s. Got %s.", expected, actual))
}

func Test_hackMatch(t *testing.T) {
	var path string
	var err error
	var expected string
	var vendor string

	var dstu2JSON []byte
	var dstu2 capabilityparser.CapabilityStatement

	store := mock.NewBasicMockHealthITProductStore()

	ctx := context.Background()
	vendorsRaw, err := store.GetHealthITProductDevelopers(ctx)
	th.Assert(t, err == nil, err)
	vendorsNorm := normalizeList(vendorsRaw)

	// basic test

	// epic
	expected = "Epic Systems Corporation"

	path = filepath.Join("../testdata", "epic_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = hackMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("Expected %s. Received %s.", expected, vendor))
}

func Test_hackMatchEpic(t *testing.T) {
	var path string
	var err error
	var expected string
	var vendor string

	var dstu2JSON []byte
	var dstu2 capabilityparser.CapabilityStatement

	store := mock.NewBasicMockHealthITProductStore()

	ctx := context.Background()
	vendorsRaw, err := store.GetHealthITProductDevelopers(ctx)
	th.Assert(t, err == nil, err)
	vendorsNorm := normalizeList(vendorsRaw)

	// epic
	expected = "Epic Systems Corporation"

	path = filepath.Join("../testdata", "epic_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = hackMatchEpic(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("Expected %s. Received %s.", expected, vendor))

	// cerner
	// has no copyright
	expected = ""

	path = filepath.Join("../testdata", "cerner_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = hackMatchEpic(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("Expected %s. Received %s.", expected, vendor))

	// meditech
	// has non-matching copyright
	expected = ""

	path = filepath.Join("../testdata", "meditech_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = hackMatchEpic(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("Expected %s. Received %s.", expected, vendor))

	// has copyright that will error

	// access copyright field and make into a non-string value to throw error
	var dstu2Int map[string]interface{}
	err = json.Unmarshal(dstu2JSON, &dstu2Int)
	th.Assert(t, err == nil, err)
	dstu2Int["copyright"] = []int{1, 2, 3} // bad format for copyright
	dstu2JSON, err = json.Marshal(dstu2Int)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	_, err = hackMatchEpic(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err != nil, "expected error to be thrown from accessing the copyright statement")
}
