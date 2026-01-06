//go:build integration
// +build integration

package capabilityhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_ResolveVendor(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	ctx := context.Background()

	var err error

	// populate healthit products
	for _, vendor := range vendors {
		err = store.AddVendor(ctx, vendor)
	}

	// populate fhir endpoint
	var epOrg7 = &endpointmanager.FHIREndpointOrganization{
		OrganizationName: "Example Inc."}

	ep := &endpointmanager.FHIREndpoint{
		URL:              "example.com/FHIR/DSTU2",
		OrganizationList: []*endpointmanager.FHIREndpointOrganization{epOrg7}}
	store.AddFHIREndpoint(ctx, ep)

	// basic test

	// capability statement
	path := filepath.Join("../../testdata", "cerner_capability_dstu2.json")
	csJSON, err := os.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	// Case 1: Vendor resolved from capability statement (publisher / hack match)
	result, err := ResolveVendor(
		ctx,
		store,
		"", // listSource
		"", // developerName (not tested here)
		cs,
	)

	th.Assert(t, err == nil, err)
	// "Cerner Corporation" second item in vendor list
	th.Assert(t, result.VendorID == vendors[1].ID,
		fmt.Sprintf("expected vendor value to be %d. Instead got %d", vendors[1].ID, result.VendorID))
	th.Assert(t, result.Source == VendorMatchCapability,
		fmt.Sprintf("expected match source %s. Instead got %s", VendorMatchCapability, result.Source))

	// test no match

	// capability statement
	path = filepath.Join("../../testdata", "novendor_capability_dstu2.json")
	csJSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err = capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	// Case 2: Capability present but no vendor match -> VendorID remains 0
	result, err = ResolveVendor(
		ctx,
		store,
		"", // listSource
		"", // developerName
		cs,
	)

	th.Assert(t, err == nil, err)
	th.Assert(t, result.VendorID == 0,
		fmt.Sprintf("expected no vendor value. Instead got %d", result.VendorID))
	th.Assert(t, result.Source == VendorMatchCapability,
		fmt.Sprintf("expected match source %s. Instead got %s", VendorMatchCapability, result.Source))

	// test no capability statement

	// Case 3: No capability statement -> no vendor match
	result, err = ResolveVendor(
		ctx,
		store,
		"",  // listSource
		"",  // developerName
		nil, // capability statement
	)

	th.Assert(t, err == nil, err)
	th.Assert(t, result.VendorID == 0,
		fmt.Sprintf("expected no vendor value. Instead got %d", result.VendorID))
	th.Assert(t, result.Source == VendorMatchNone,
		fmt.Sprintf("expected match source %s. Instead got %s", VendorMatchNone, result.Source))

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

	// Case 4: Malformed capability statement -> error returned, no vendor resolved
	result, err = ResolveVendor(ctx, store, "", "", cs)

	th.Assert(t, err != nil, "expected an error from accessing the publisher field in the capability statment.")
	th.Assert(t, result.VendorID == 0, fmt.Sprintf("expected no vendor value. Instead got %d", result.VendorID))
	th.Assert(t, result.Source == "", "expected no source on error")

	// add endpoint with list source in CHPL products info file
	// populate fhir endpoint

	var epOrg8 = &endpointmanager.FHIREndpointOrganization{
		OrganizationName: "Example Inc."}

	ep2 := &endpointmanager.FHIREndpoint{
		URL:              "example2.com/FHIR/DSTU2",
		OrganizationList: []*endpointmanager.FHIREndpointOrganization{epOrg8},
		ListSource:       "https://nextgen.com/api/practice-search"}
	store.AddFHIREndpoint(ctx, ep2)

	// capability statement
	path = filepath.Join("../../testdata", "novendor_capability_dstu2.json")
	csJSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)

	cs, err = capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	// explicit CHPL developer name that exists in the DB
	developerName := vendors[0].Name // "Epic Systems Corporation"

	// Case 5: Explicit CHPL developer name -> vendor resolved immediately
	result, err = ResolveVendor(
		ctx,
		store,
		"",            // listSource
		developerName, // CHPL developer
		nil,           // capability statement
	)

	th.Assert(t, err == nil, err)

	th.Assert(
		t,
		result.VendorID == vendors[0].ID,
		fmt.Sprintf("expected vendor %d, got %d", vendors[0].ID, result.VendorID),
	)

	th.Assert(
		t,
		result.Source == VendorMatchCHPL,
		fmt.Sprintf("expected VendorMatchCHPL, got %s", result.Source),
	)

	th.Assert(
		t,
		result.Detail == developerName,
		fmt.Sprintf("expected detail %q, got %q", developerName, result.Detail),
	)
}

func Test_getVendorMatch(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var path string
	var err error
	var expected int

	var dstu2JSON []byte
	var dstu2 capabilityparser.CapabilityStatement
	var vendor int

	ctx := context.Background()

	var dstu2Int map[string]interface{}

	// populate vendors table
	for _, vendorItem := range vendors {
		err = store.AddVendor(ctx, vendorItem)
	}

	// Case 1: Cerner -> resolved via publisherMatch
	expected = vendors[1].ID // "Cerner Corporation"

	path = filepath.Join("../../testdata", "cerner_capability_dstu2.json")
	dstu2JSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %d. Got %d.", expected, vendor))

	// Case 2: Epic -> publisher missing, resolved via hackMatch
	expected = vendors[0].ID // "Epic Systems Corporation" // this uses the "hackMatch" capability

	path = filepath.Join("../../testdata", "epic_capability_dstu2.json")
	dstu2JSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %d. Got %d.", expected, vendor))

	// Case 3: Epic with malformed copyright -> hackMatch error
	err = json.Unmarshal(dstu2JSON, &dstu2Int)
	th.Assert(t, err == nil, err)
	dstu2Int["copyright"] = []int{1, 2, 3} // bad format for copyright
	dstu2JSON, err = json.Marshal(dstu2Int)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err != nil, "expected error due to accessing the copyright")
	th.Assert(t, vendor == 0, fmt.Sprintf("expected no vendor value. Instead got %d", vendor))

	// Case 4: Allscripts -> resolved via publisherMatch
	expected = vendors[5].ID // "Allscripts"

	path = filepath.Join("../../testdata", "allscripts_capability_dstu2.json")
	dstu2JSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %d. Got %d.", expected, vendor))

	// Case 5: Meditech -> resolved via publisherMatch
	expected = vendors[4].ID // "Medical Information Technology, Inc. (MEDITECH)"

	path = filepath.Join("../../testdata", "meditech_capability_dstu2.json")
	dstu2JSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %d. Got %d.", expected, vendor))

	// test error getting match
	// Case 6: Malformed publisher -> publisherMatch error

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
	th.Assert(t, vendor == 0, fmt.Sprintf("expected no vendor value. Instead got %d", vendor))
}

func Test_publisherMatch(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var path string
	var err error
	var expected string

	var dstu2JSON []byte
	var dstu2 capabilityparser.CapabilityStatement
	var vendor string

	var dstu2Int map[string]interface{}

	ctx := context.Background()

	// populate vendors
	for _, vendorItem := range vendors {
		err = store.AddVendor(ctx, vendorItem)
	}

	vendorsRaw, err := store.GetVendorNames(ctx)
	th.Assert(t, err == nil, err)
	vendorsNorm := normalizeList(vendorsRaw)

	// Case 1: Cerner —> resolved via publisher field
	expected = "Cerner Corporation"

	path = filepath.Join("../../testdata", "cerner_capability_dstu2.json")
	dstu2JSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = publisherMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// Case 2: Epic -> publisher missing
	expected = "" // the capability statement is missing the publisher

	path = filepath.Join("../../testdata", "epic_capability_dstu2.json")
	dstu2JSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = publisherMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// Case 3: Allscripts —> publisher-based match
	expected = "Allscripts"

	path = filepath.Join("../../testdata", "allscripts_capability_dstu2.json")
	dstu2JSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = publisherMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// Case 4: Meditech -> normalization and fluff stripping
	expected = "Medical Information Technology, Inc. (MEDITECH)"
	path = filepath.Join("../../testdata", "meditech_capability_dstu2.json")
	dstu2JSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = publisherMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// test error getting match
	// Case 5: Malformed publisher -> error

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

func Test_hackMatch(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var path string
	var err error
	var expected string
	var vendor string

	var dstu2JSON []byte
	var dstu2 capabilityparser.CapabilityStatement

	ctx := context.Background()

	// populate vendors
	for _, vendorItem := range vendors {
		err = store.AddVendor(ctx, vendorItem)
	}

	vendorsRaw, err := store.GetVendorNames(ctx)
	th.Assert(t, err == nil, err)
	vendorsNorm := normalizeList(vendorsRaw)

	// basic test

	// Case: hackMatch fallback for Epic
	expected = "Epic Systems Corporation"

	path = filepath.Join("../../testdata", "epic_capability_dstu2.json")
	dstu2JSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = hackMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("Expected %s. Received %s.", expected, vendor))
}

func Test_hackMatchEpic(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)
	var path string
	var err error
	var expected string
	var vendor string

	var dstu2JSON []byte
	var dstu2 capabilityparser.CapabilityStatement

	ctx := context.Background()

	// populate vendors
	for _, vendorItem := range vendors {
		err = store.AddVendor(ctx, vendorItem)
	}

	vendorsRaw, err := store.GetVendorNames(ctx)
	th.Assert(t, err == nil, err)
	vendorsNorm := normalizeList(vendorsRaw)

	// Case 1: Epic detected via copyright string
	expected = "Epic Systems Corporation"

	path = filepath.Join("../../testdata", "epic_capability_dstu2.json")
	dstu2JSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = hackMatchEpic(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("Expected %s. Received %s.", expected, vendor))

	// cerner
	// Case 2: Capability statement has no copyright field
	expected = ""

	path = filepath.Join("../../testdata", "cerner_capability_dstu2.json")
	dstu2JSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = hackMatchEpic(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("Expected %s. Received %s.", expected, vendor))

	// meditech
	// Case 3: Capability statement has copyright, but it does not contain the string "epic"
	expected = ""

	path = filepath.Join("../../testdata", "meditech_capability_dstu2.json")
	dstu2JSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = hackMatchEpic(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("Expected %s. Received %s.", expected, vendor))

	// Case 4: Malformed copyright field (non-string)

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
