// +build integration

package chplmapper

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/spf13/viper"
)

var store *postgresql.Store

var hitps []*endpointmanager.HealthITProduct = []*endpointmanager.HealthITProduct{
	&endpointmanager.HealthITProduct{
		Name:                 "Carefluence Open API",
		Version:              "1",
		Developer:            "Carefluence",
		CertificationStatus:  "Active",
		CertificationDate:    time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
		CertificationEdition: "2014",
		CHPLID:               "15.04.04.2657.Care.01.00.0.160701",
		APIURL:               "http://carefluence.com/Carefluence-OpenAPI-Documentation.html",
	},
	&endpointmanager.HealthITProduct{
		Name:                 "EpicCare Ambulatory Base",
		Version:              "February 2020",
		Developer:            "Epic Systems Corporation",
		CertificationStatus:  "Active",
		CertificationDate:    time.Date(2020, 2, 20, 0, 0, 0, 0, time.UTC),
		CertificationEdition: "2015",
		CHPLID:               "15.04.04.1447.Epic.AM.13.1.200220",
		APIURL:               "https://open.epic.com/Interface/FHIR",
	},
	&endpointmanager.HealthITProduct{
		Name:                 "PowerChart (Clinical)",
		Version:              "2018.01",
		Developer:            "Cerner Corporation",
		CertificationStatus:  "Active",
		CertificationDate:    time.Date(2018, 7, 27, 0, 0, 0, 0, time.UTC),
		CertificationEdition: "2015",
		CHPLID:               "15.04.04.1221.Powe.18.03.1.180727",
		APIURL:               "http://fhir.cerner.com/authorization/",
	},
	&endpointmanager.HealthITProduct{
		Name:                 "Health Services Analytics",
		Version:              "8.00 SP1-SP5",
		Developer:            "Cerner Health Services, Inc.",
		CertificationStatus:  "Withdrawn by Developer",
		CertificationDate:    time.Date(2017, 12, 5, 0, 0, 0, 0, time.UTC),
		CertificationEdition: "2014",
		CHPLID:               "14.07.07.1222.HEA5.03.01.1.171205",
	},
	&endpointmanager.HealthITProduct{
		Name:                 "MEDITECH 6.0 Electronic Health Record Core HCIS",
		Version:              "v6.08",
		Developer:            "Medical Information Technology, Inc. (MEDITECH)",
		CertificationStatus:  "Active",
		CertificationDate:    time.Date(2017, 12, 20, 0, 0, 0, 0, time.UTC),
		CertificationEdition: "2015",
		CHPLID:               "15.04.04.2931.MEDI.EH.00.1.171220",
		APIURL:               "https://home.meditech.com/en/d/restapiresources/pages/apidoc.htm",
	},
	&endpointmanager.HealthITProduct{
		Name:                 "Sunrise Acute Care for Hospital-based Providers",
		Version:              "16.3 CU3",
		Developer:            "Allscripts",
		CertificationStatus:  "Active",
		CertificationDate:    time.Date(2017, 12, 1, 0, 0, 0, 0, time.UTC),
		CertificationEdition: "2015",
		CHPLID:               "15.04.04.2891.Sunr.16.03.1.171201",
		APIURL:               "https://developer.allscripts.com/Content/fhir/",
	},
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

func Test_MatchEndpointToVendorAndProduct(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	ctx := context.Background()

	var err error

	// populate healthit products
	for _, hitp := range hitps {
		err = store.AddHealthITProduct(ctx, hitp)
	}

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

	err = MatchEndpointToVendorAndProduct(ctx, ep, store)
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

	err = MatchEndpointToVendorAndProduct(ctx, ep, store)
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
	err = MatchEndpointToVendorAndProduct(ctx, ep, store)
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

	err = MatchEndpointToVendorAndProduct(ctx, ep, store)
	th.Assert(t, err != nil, "expected an error from accessing the publisher field in the capability statment.")
	th.Assert(t, len(ep.Vendor) == 0, fmt.Sprintf("expected no vendor value. Instead got %s", ep.Vendor))
}

func Test_getVendorMatch(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var path string
	var err error
	var expected string

	var dstu2JSON []byte
	var dstu2 capabilityparser.CapabilityStatement
	var vendor string

	ctx := context.Background()

	var dstu2Int map[string]interface{}

	// populate healthit products
	for _, hitp := range hitps {
		err = store.AddHealthITProduct(ctx, hitp)
	}

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

	// populate healthit products
	for _, hitp := range hitps {
		err = store.AddHealthITProduct(ctx, hitp)
	}

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

	// populate healthit products
	for _, hitp := range hitps {
		err = store.AddHealthITProduct(ctx, hitp)
	}

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
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)
	var path string
	var err error
	var expected string
	var vendor string

	var dstu2JSON []byte
	var dstu2 capabilityparser.CapabilityStatement

	ctx := context.Background()

	// populate healthit products
	for _, hitp := range hitps {
		err = store.AddHealthITProduct(ctx, hitp)
	}

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

func setup() error {
	var err error
	store, err = postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))

	return err
}

func teardown() {
	store.Close()
}
