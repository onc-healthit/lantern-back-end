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
	epStore, err := getMockStoreEP()
	th.Assert(t, err == nil, err)
	hitpStore, err := getMockStore()

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

	matched, err := MatchEndpointToVendorAndProduct(ctx, ep, epStore, hitpStore)
	th.Assert(t, err == nil, err)
	th.Assert(t, matched, "expected endpoint to be successfully matched")
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

	matched, err = MatchEndpointToVendorAndProduct(ctx, ep, epStore, hitpStore)
	th.Assert(t, err == nil, err)
	th.Assert(t, !matched, "expected no match")
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
	matched, err = MatchEndpointToVendorAndProduct(ctx, ep, epStore, hitpStore)
	th.Assert(t, err == nil, err)
	th.Assert(t, !matched, "expected no match")
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

	matched, err = MatchEndpointToVendorAndProduct(ctx, ep, epStore, hitpStore)
	th.Assert(t, err != nil, "expected an error from accessing the publisher field in the capability statment.")
	th.Assert(t, !matched, "expected no match")
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
	store, err := getMockStore()
	th.Assert(t, err == nil, err)

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
	store, err := getMockStore()
	th.Assert(t, err == nil, err)

	var vendorsNorm []string

	vendorsRaw, err := store.GetHealthITProductDevelopers(ctx)
	th.Assert(t, err == nil, err)

	for _, vendorRaw := range vendorsRaw {
		vendorNorm := normalizeName(vendorRaw)
		vendorsNorm = append(vendorsNorm, vendorNorm)
	}

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

func Test_hackMatch(t *testing.T) {
	var path string
	var err error
	var expected string
	var vendor string

	var dstu2JSON []byte
	var dstu2 capabilityparser.CapabilityStatement

	store, err := getMockStore()
	th.Assert(t, err == nil, err)

	var vendorsNorm []string

	ctx := context.Background()
	vendorsRaw, err := store.GetHealthITProductDevelopers(ctx)
	th.Assert(t, err == nil, err)

	for _, vendorRaw := range vendorsRaw {
		vendorNorm := normalizeName(vendorRaw)
		vendorsNorm = append(vendorsNorm, vendorNorm)
	}

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

	store, err := getMockStore()
	th.Assert(t, err == nil, err)

	var vendorsNorm []string

	ctx := context.Background()
	vendorsRaw, err := store.GetHealthITProductDevelopers(ctx)
	th.Assert(t, err == nil, err)

	for _, vendorRaw := range vendorsRaw {
		vendorNorm := normalizeName(vendorRaw)
		vendorsNorm = append(vendorsNorm, vendorNorm)
	}

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

	vendor, err = hackMatchEpic(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err != nil, "expected error to be thrown from accessing the copyright statement")
}

func getMockStore() (endpointmanager.HealthITProductStore, error) {
	hitp, err := mock.NewStore()
	if err != nil {
		return nil, err
	}

	hitp.GetHealthITProductDevelopersFn = func(ctx context.Context) ([]string, error) {
		devList := []string{
			"Epic Systems Corporation",
			"Cerner Corporation",
			"Cerner Health Services, Inc.",
			"Medical Information Technology, Inc. (MEDITECH)",
			"Allscripts",
		}

		return devList, nil
	}

	return hitp, nil
}

func getMockStoreEP() (endpointmanager.FHIREndpointStore, error) {
	ep, err := mock.NewStore()
	if err != nil {
		return nil, err
	}

	ep.UpdateFHIREndpointFn = func(ctx context.Context, ep *endpointmanager.FHIREndpoint) error {
		return nil
	}

	return ep, nil
}
