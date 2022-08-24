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

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/spf13/viper"
)

var store *postgresql.Store

var vendors []*endpointmanager.Vendor = []*endpointmanager.Vendor{
	&endpointmanager.Vendor{
		Name:          "Epic Systems Corporation",
		DeveloperCode: "A",
		CHPLID:        1,
	},
	&endpointmanager.Vendor{
		Name:          "Cerner Corporation",
		DeveloperCode: "B",
		CHPLID:        2,
	},
	&endpointmanager.Vendor{
		Name:          "Cerner Health Services, Inc.",
		DeveloperCode: "C",
		CHPLID:        3,
	},
	&endpointmanager.Vendor{
		Name:          "Carefluence",
		DeveloperCode: "D",
		CHPLID:        4,
	},
	&endpointmanager.Vendor{
		Name:          "Medical Information Technology, Inc. (MEDITECH)",
		DeveloperCode: "E",
		CHPLID:        5,
	},
	&endpointmanager.Vendor{
		Name:          "Allscripts",
		DeveloperCode: "F",
		CHPLID:        6,
	},
	&endpointmanager.Vendor{
		Name:          "NextGen Healthcare",
		DeveloperCode: "G",
		CHPLID:        7,
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

func Test_openProductLinksFile(t *testing.T) {
	path := filepath.Join("../../testdata", "test_chpl_product_mapping_bad.json")
	chplProductNameVersion, err := openProductLinksFile(path)
	th.Assert(t, err == nil, err)
	// make sure that product name with wrong key in test file is not in the returned structure
	th.Assert(t, chplProductNameVersion["badchplidentry"] == nil, "Field keyed with bad chplid key should not exist")
	// make sure that product name with wrong key in test file is not in the returned structure
	th.Assert(t, chplProductNameVersion["Allscripts FHIR"] == nil, "Field keyed as noname should not exist")
	// make sure that product version with correct key in test file is in the returned structure
	th.Assert(t, chplProductNameVersion["FooBarProduct"]["4.0"] == "somefakeCHPLID", "Link represented correctly should exist")
	// make sure that product version with wrong key in test file is not in the returned structure
	th.Assert(t, chplProductNameVersion["FooBarProduct"]["2.0"] == "", "Field keyed as noversion should not exist")
}

func Test_MatchEndpointToProduct(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	ctx := context.Background()

	var err error

	// populate healthit products
	var hitp1 = &endpointmanager.HealthITProduct{
		Name:                 "FooBarProduct",
		Version:              "2.0",
		APISyntax:            "FHIR DSTU2",
		CHPLID:               "somefakeCHPLID",
		CertificationEdition: "2014",
		CertificationStatus: "Active"}
	var hitp2 = &endpointmanager.HealthITProduct{
		Name:                 "Allscripts FHIR",
		Version:              "2.0",
		APISyntax:            "FHIR DSTU2",
		CHPLID:               "correctNameIncorrectVersion",
		CertificationEdition: "2014"}
	var hitp3 = &endpointmanager.HealthITProduct{
		Name:                 "WrongName",
		Version:              "19.4.121.0",
		APISyntax:            "FHIR DSTU2",
		CHPLID:               "correctVersionIncorrectName",
		CertificationEdition: "2014",
		CertificationStatus: "Active"}
	var hitp4 = &endpointmanager.HealthITProduct{
		Name:                 "Allscripts FHIR",
		Version:              "19.4.121.0",
		APISyntax:            "FHIR DSTU2",
		CHPLID:               "CorrectVersionAndName",
		CertificationEdition: "2014",
		CertificationStatus: "Active"}
	var hitp5 = &endpointmanager.HealthITProduct{
		Name:                 "BlueButtonPRO",
		Version:              "2",
		APISyntax:            "FHIR DSTU2",
		CHPLID:               "15.04.04.1322.Blue.02.00.0.200807",
		CertificationEdition: "2015"}
	var hitp6 = &endpointmanager.HealthITProduct{
		Name:                 "HIEBus",
		Version:              "30.0.0",
		APISyntax:            "FHIR DSTU2",
		CHPLID:               "CHP-019353",
		CertificationEdition: "2014",
		CertificationStatus: "Active"}
	var hitp7 = &endpointmanager.HealthITProduct{
		Name:                 "HIEBus™",
		Version:              "30.0.5",
		APISyntax:            "FHIR DSTU2",
		CHPLID:               "CHP-019355",
		CertificationEdition: "2014",
		CertificationStatus: "Active"}
	var hitp8 = &endpointmanager.HealthITProduct{
		Name:                 "HIEBus",
		Version:              "20.0.0",
		APISyntax:            "FHIR DSTU2",
		CHPLID:               "CHP-019350",
		CertificationEdition: "2014",
		CertificationStatus: "Retired"}

	err = store.AddHealthITProduct(ctx, hitp1)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}
	err = store.AddHealthITProduct(ctx, hitp2)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}
	err = store.AddHealthITProduct(ctx, hitp3)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}
	err = store.AddHealthITProduct(ctx, hitp4)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}
	err = store.AddHealthITProduct(ctx, hitp5)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}
	err = store.AddHealthITProduct(ctx, hitp6)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}
	err = store.AddHealthITProduct(ctx, hitp7)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}
	err = store.AddHealthITProduct(ctx, hitp8)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}

	// populate fhir endpoint
	ep := &endpointmanager.FHIREndpoint{
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{"Example Inc."}}
	store.AddFHIREndpoint(ctx, ep)

	// capability statement
	path := filepath.Join("../../testdata", "cerner_capability_dstu2.json")
	csJSON, err := ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	// endpoint info
	epInfo := &endpointmanager.FHIREndpointInfo{
		URL:                 ep.URL,
		CapabilityStatement: cs}

	path = filepath.Join("../../testdata", "test_chpl_product_mapping.json")
	chplEndpointListPath := filepath.Join("../../testdata", "test_chpl_products_info.json")

	listSourceMap, err := OpenCHPLEndpointListInfoFile(chplEndpointListPath)
	th.Assert(t, err == nil, err)

	err = MatchEndpointToProduct(ctx, epInfo, store, path, listSourceMap)
	th.Assert(t, err == nil, err)
	// No healthIT product should have matched
	th.Assert(t, epInfo.HealthITProductID == 0, fmt.Sprintf("expected HealthITProductID value to be %d. Instead got %d", 0, epInfo.HealthITProductID))

	// capability statement
	path = filepath.Join("../../testdata", "allscripts_capability_dstu2.json")
	csJSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err = capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	// populate fhir endpoint
	ep = &endpointmanager.FHIREndpoint{
		URL:               "example2.com/FHIR/DSTU2",
		OrganizationNames: []string{"Example2 Inc."}}
	store.AddFHIREndpoint(ctx, ep)

	// populate fhir endpoint with list source found in CHPL products info file
	ep2 := &endpointmanager.FHIREndpoint{
		URL:               "example3.com/FHIR/DSTU2",
		OrganizationNames: []string{"Example2 Inc."},
		ListSource:        "https://api.bluebuttonpro.com/swagger/index.html"}
	store.AddFHIREndpoint(ctx, ep2)

	ep3 := &endpointmanager.FHIREndpoint{
		URL:               "example4.com/FHIR/DSTU2",
		OrganizationNames: []string{"Example2 Inc."},
		ListSource:        "https://nextgen.com/api/practice-search"}
	store.AddFHIREndpoint(ctx, ep3)

	// endpoint info
	epInfo = &endpointmanager.FHIREndpointInfo{
		URL:                 ep.URL,
		CapabilityStatement: cs}

	// endpoint info
	epInfo2 := &endpointmanager.FHIREndpointInfo{
		URL:                 ep2.URL,
		CapabilityStatement: nil}

	err = MatchEndpointToProduct(ctx, epInfo, store, "../../testdata/test_chpl_product_mapping.json", listSourceMap)
	th.Assert(t, err == nil, err)
	healthITProductID, err := store.GetHealthITProductIDByCHPLID(ctx, "CorrectVersionAndName")
	th.Assert(t, err == nil, err)
	actualHealthITProductIDs, err := store.GetHealthITProductIDsByMapID(ctx, epInfo.HealthITProductID)
	th.Assert(t, err == nil, err)
	// healthIT product with ID healthITProductID should have matched
	th.Assert(t, actualHealthITProductIDs[0] == healthITProductID, fmt.Sprintf("expected HealthITProductID value to be %d. Instead got %d", healthITProductID, actualHealthITProductIDs[0]))

	err = MatchEndpointToProduct(ctx, epInfo2, store, "../../testdata/test_chpl_product_mapping.json", listSourceMap)
	th.Assert(t, err == nil, err)
	healthITProductID, err = store.GetHealthITProductIDByCHPLID(ctx, "15.04.04.1322.Blue.02.00.0.200807")
	th.Assert(t, err == nil, err)
	actualHealthITProductIDs, err = store.GetHealthITProductIDsByMapID(ctx, epInfo2.HealthITProductID)
	th.Assert(t, err == nil, err)
	// healthIT product with ID healthITProductID should have matched
	th.Assert(t, actualHealthITProductIDs[0] == healthITProductID, fmt.Sprintf("expected HealthITProductID value to be %d. Instead got %d", healthITProductID, actualHealthITProductIDs))

	// capability statement
	path = filepath.Join("../../testdata", "advantagecare_physicians_stu3.json")
	csJSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err = capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	var hitp9 = &endpointmanager.HealthITProduct{
		Name:                 "Epic",
		Version:              "February 2021",
		APISyntax:            "FHIR DSTU3",
		CHPLID:               "FakeCHPLID",
		CertificationEdition: "2014"}

	err = store.AddHealthITProduct(ctx, hitp9)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}

	var hitp10 = &endpointmanager.HealthITProduct{
		Name:                 "NextGen Enterprise EHR",
		Version:              "6.2021.1 Patch 79",
		APISyntax:            "FHIR DSTU3",
		CHPLID:               "15.04.04.1918.Next.60.09.1.220303",
		CertificationEdition: "2015"}

	err = store.AddHealthITProduct(ctx, hitp10)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}

	var hitp11 = &endpointmanager.HealthITProduct{
		Name:                 "NextGen Enterprise EHR",
		Version:              "6.2021.1 Cures",
		APISyntax:            "FHIR DSTU3",
		CHPLID:               "15.04.04.1918.Next.60.10.1.220318",
		CertificationEdition: "2015"}

	err = store.AddHealthITProduct(ctx, hitp11)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}

	// endpoint info
	epInfo3 := &endpointmanager.FHIREndpointInfo{
		URL:                 ep3.URL,
		CapabilityStatement: nil}

	epInfo.CapabilityStatement = cs

	err = MatchEndpointToProduct(ctx, epInfo, store, "../../testdata/test_chpl_product_mapping.json", listSourceMap)
	th.Assert(t, err == nil, err)
	actualHealthITProductIDs, err = store.GetHealthITProductIDsByMapID(ctx, epInfo.HealthITProductID)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(actualHealthITProductIDs) == 2, fmt.Sprintf("Expected endpoint to map to 2 healthIT products, instead mapped to %d", len(actualHealthITProductIDs)))

	err = MatchEndpointToProduct(ctx, epInfo3, store, "../../testdata/test_chpl_product_mapping.json", listSourceMap)
	th.Assert(t, err == nil, err)
	actualHealthITProductIDs, err = store.GetHealthITProductIDsByMapID(ctx, epInfo3.HealthITProductID)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(actualHealthITProductIDs) == 2, fmt.Sprintf("Expected endpoint to map to 2 healthIT products, instead mapped to %d", len(actualHealthITProductIDs)))
<<<<<<< HEAD
=======

>>>>>>> 92ea5194 (Chpl endpoint list mapping (#293))

	// Test matching to product by name and version

	// populate fhir endpoint
	ep = &endpointmanager.FHIREndpoint{
		URL:               "example5.com/FHIR/DSTU2",
		OrganizationNames: []string{"Example Inc."}}
	store.AddFHIREndpoint(ctx, ep)

	// capability statement with product HIEBus
	path = filepath.Join("../../testdata", "careevolution_dstu2.json")
	csJSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err = capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	// endpoint info
	epInfo = &endpointmanager.FHIREndpointInfo{
		URL:                 ep.URL,
		CapabilityStatement: cs}

	path = filepath.Join("../../testdata", "test_chpl_product_mapping.json")
	err = MatchEndpointToProduct(ctx, epInfo, store, path, listSourceMap)
	th.Assert(t, err == nil, err)
	actualHealthITProductIDs, err = store.GetHealthITProductIDsByMapID(ctx, epInfo.HealthITProductID)
	th.Assert(t, err == nil, err)

	// Should only be one match with the correct name and version 
	th.Assert(t, len(actualHealthITProductIDs) == 1, fmt.Sprintf("Expected endpoint to map to 1 healthIT products, instead mapped to %d", len(actualHealthITProductIDs)))

	// Test matching to product by name and no version

	// populate fhir endpoint
	ep = &endpointmanager.FHIREndpoint{
		URL:               "example6.com/FHIR/DSTU2",
		OrganizationNames: []string{"Example Inc."}}
	store.AddFHIREndpoint(ctx, ep)

	// remove the version field from the software element of the care evolution capability statement
	var csInt map[string]interface{}
	var softwareInt map[string]interface{}

	err = json.Unmarshal(csJSON, &csInt)
	th.Assert(t, err == nil, err)

	softwareInt = csInt["software"].(map[string]interface{})

	delete(softwareInt, "version")

	csInt["software"] = softwareInt

	csJSON, err = json.Marshal(csInt)
	th.Assert(t, err == nil, err)

	cs, err = capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	// endpoint info
	epInfo = &endpointmanager.FHIREndpointInfo{
		URL:                 ep.URL,
		CapabilityStatement: cs}

	path = filepath.Join("../../testdata", "test_chpl_product_mapping.json")
	err = MatchEndpointToProduct(ctx, epInfo, store, path, listSourceMap)
	th.Assert(t, err == nil, err)
	actualHealthITProductIDs, err = store.GetHealthITProductIDsByMapID(ctx, epInfo.HealthITProductID)
	th.Assert(t, err == nil, err)

	// Should match to 2 products since there are only 2 active healthit products with the name HIEBus- version does not matter
	th.Assert(t, len(actualHealthITProductIDs) == 2, fmt.Sprintf("Expected endpoint to map to 2 healthIT products, instead mapped to %d", len(actualHealthITProductIDs)))
}

func Test_MatchEndpointToVendor(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	ctx := context.Background()

	var err error

	// populate healthit products
	for _, vendor := range vendors {
		err = store.AddVendor(ctx, vendor)
	}
	// populate fhir endpoint
	ep := &endpointmanager.FHIREndpoint{
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{"Example Inc."}}
	store.AddFHIREndpoint(ctx, ep)

	// basic test

	// capability statement
	path := filepath.Join("../../testdata", "cerner_capability_dstu2.json")
	csJSON, err := ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	chplEndpointListPath := filepath.Join("../../testdata", "test_chpl_products_info.json")
	listSourceMap, err := OpenCHPLEndpointListInfoFile(chplEndpointListPath)
	th.Assert(t, err == nil, err)

	// endpoint info
	epInfo := &endpointmanager.FHIREndpointInfo{
		URL:                 ep.URL,
		CapabilityStatement: cs}

	err = MatchEndpointToVendor(ctx, epInfo, store, listSourceMap)
	th.Assert(t, err == nil, err)
	// "Cerner Corporation" second item in vendor list
	th.Assert(t, epInfo.VendorID == vendors[1].ID, fmt.Sprintf("expected vendor value to be %d. Instead got %d", vendors[1].ID, epInfo.VendorID))

	// test no match

	// capability statement
	path = filepath.Join("../../testdata", "novendor_capability_dstu2.json")
	csJSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err = capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	// endpoint
	epInfo = &endpointmanager.FHIREndpointInfo{
		URL:                 ep.URL,
		CapabilityStatement: cs}

	err = MatchEndpointToVendor(ctx, epInfo, store, listSourceMap)
	th.Assert(t, err == nil, err)
	th.Assert(t, epInfo.VendorID == 0, fmt.Sprintf("expected no vendor value. Instead got %d", epInfo.VendorID))

	// test no capability statement

	// endpoint
	epInfo = &endpointmanager.FHIREndpointInfo{
		URL: ep.URL}
	err = MatchEndpointToVendor(ctx, epInfo, store, listSourceMap)
	th.Assert(t, err == nil, err)
	th.Assert(t, epInfo.VendorID == 0, fmt.Sprintf("expected no vendor value. Instead got %d", epInfo.VendorID))

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
	epInfo = &endpointmanager.FHIREndpointInfo{
		URL:                 ep.URL,
		CapabilityStatement: cs}

	err = MatchEndpointToVendor(ctx, epInfo, store, listSourceMap)
	th.Assert(t, err != nil, "expected an error from accessing the publisher field in the capability statment.")
	th.Assert(t, epInfo.VendorID == 0, fmt.Sprintf("expected no vendor value. Instead got %d", epInfo.VendorID))

	// add endpoint with list source in CHPL products info file
	// populate fhir endpoint
	ep2 := &endpointmanager.FHIREndpoint{
		URL:               "example2.com/FHIR/DSTU2",
		OrganizationNames: []string{"Example Inc."},
		ListSource:        "https://nextgen.com/api/practice-search"}
	store.AddFHIREndpoint(ctx, ep2)

	// endpoint info
	epInfo = &endpointmanager.FHIREndpointInfo{
		URL:                 ep2.URL,
		CapabilityStatement: cs}

	err = MatchEndpointToVendor(ctx, epInfo, store, listSourceMap)
	th.Assert(t, err == nil, err)
	th.Assert(t, epInfo.VendorID == vendors[6].ID, fmt.Sprintf("expected vendor value to be %d. Instead got %d", vendors[6].ID, epInfo.VendorID))
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

	// populate healthit products
	for _, vendorItem := range vendors {
		err = store.AddVendor(ctx, vendorItem)
	}

	// cerner
	expected = vendors[1].ID // "Cerner Corporation"

	path = filepath.Join("../../testdata", "cerner_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %d. Got %d.", expected, vendor))

	// epic
	expected = vendors[0].ID // "Epic Systems Corporation" // this uses the "hackMatch" capability

	path = filepath.Join("../../testdata", "epic_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %d. Got %d.", expected, vendor))

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
	th.Assert(t, vendor == 0, fmt.Sprintf("expected no vendor value. Instead got %d", vendor))

	// allscripts
	expected = vendors[5].ID // "Allscripts"

	path = filepath.Join("../../testdata", "allscripts_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %d. Got %d.", expected, vendor))

	// meditech
	expected = vendors[4].ID // "Medical Information Technology, Inc. (MEDITECH)"

	path = filepath.Join("../../testdata", "meditech_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = getVendorMatch(ctx, dstu2, store)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %d. Got %d.", expected, vendor))

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

	// populate healthit products
	for _, vendorItem := range vendors {
		err = store.AddVendor(ctx, vendorItem)
	}

	vendorsRaw, err := store.GetVendorNames(ctx)
	th.Assert(t, err == nil, err)
	vendorsNorm := normalizeList(vendorsRaw)

	// cerner
	expected = "Cerner Corporation"

	path = filepath.Join("../../testdata", "cerner_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = publisherMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// epic
	expected = "" // the capability statement is missing the publisher

	path = filepath.Join("../../testdata", "epic_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = publisherMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// allscripts
	expected = "Allscripts" // the capability statement is missing the publisher

	path = filepath.Join("../../testdata", "allscripts_capability_dstu2.json")
	dstu2JSON, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)

	dstu2, err = capabilityparser.NewCapabilityStatement(dstu2JSON)
	th.Assert(t, err == nil, err)

	vendor, err = publisherMatch(dstu2, vendorsNorm, vendorsRaw)
	th.Assert(t, err == nil, err)
	th.Assert(t, vendor == expected, fmt.Sprintf("expected vendor to be %s. Got %s.", expected, vendor))

	// meditech
	expected = "Medical Information Technology, Inc. (MEDITECH)" // the capability statement is missing the publisher

	path = filepath.Join("../../testdata", "meditech_capability_dstu2.json")
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
	for _, vendorItem := range vendors {
		err = store.AddVendor(ctx, vendorItem)
	}

	vendorsRaw, err := store.GetVendorNames(ctx)
	th.Assert(t, err == nil, err)
	vendorsNorm := normalizeList(vendorsRaw)

	// basic test

	// epic
	expected = "Epic Systems Corporation"

	path = filepath.Join("../../testdata", "epic_capability_dstu2.json")
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
	for _, vendorItem := range vendors {
		err = store.AddVendor(ctx, vendorItem)
	}

	vendorsRaw, err := store.GetVendorNames(ctx)
	th.Assert(t, err == nil, err)
	vendorsNorm := normalizeList(vendorsRaw)

	// epic
	expected = "Epic Systems Corporation"

	path = filepath.Join("../../testdata", "epic_capability_dstu2.json")
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

	path = filepath.Join("../../testdata", "cerner_capability_dstu2.json")
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

	path = filepath.Join("../../testdata", "meditech_capability_dstu2.json")
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