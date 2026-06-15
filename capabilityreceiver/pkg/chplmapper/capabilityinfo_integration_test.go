//go:build integration
// +build integration

package chplmapper

import (
	"context"
	"encoding/json"
	"fmt"
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
		CertificationStatus:  "Active"}
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
		CertificationStatus:  "Active"}
	var hitp4 = &endpointmanager.HealthITProduct{
		Name:                 "Allscripts FHIR",
		Version:              "19.4.121.0",
		APISyntax:            "FHIR DSTU2",
		CHPLID:               "CorrectVersionAndName",
		CertificationEdition: "2014",
		CertificationStatus:  "Active"}
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
		CertificationStatus:  "Active"}
	var hitp7 = &endpointmanager.HealthITProduct{
		Name:                 "HIEBus™",
		Version:              "30.0.5",
		APISyntax:            "FHIR DSTU2",
		CHPLID:               "CHP-019355",
		CertificationEdition: "2014",
		CertificationStatus:  "Active"}
	var hitp8 = &endpointmanager.HealthITProduct{
		Name:                 "HIEBus",
		Version:              "20.0.0",
		APISyntax:            "FHIR DSTU2",
		CHPLID:               "CHP-019350",
		CertificationEdition: "2014",
		CertificationStatus:  "Retired"}

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

	var epOrg = &endpointmanager.FHIREndpointOrganization{
		OrganizationName: "Example Inc."}

	// populate fhir endpoint
	ep := &endpointmanager.FHIREndpoint{
		URL:              "example.com/FHIR/DSTU2",
		OrganizationList: []*endpointmanager.FHIREndpointOrganization{epOrg}}

	store.AddFHIREndpoint(ctx, ep)

	// capability statement
	path := filepath.Join("../../testdata", "cerner_capability_dstu2.json")
	csJSON, err := os.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	// endpoint info
	epInfo := &endpointmanager.FHIREndpointInfo{
		URL:                 ep.URL,
		CapabilityStatement: cs}

	path = filepath.Join("../../testdata", "test_chpl_product_mapping.json")

	// Reset to avoid accumulating mappings from previous MatchEndpointToProduct calls
	epInfo.HealthITProductID = 0

	// Case 1:
	// Capability-based matching only.
	err = MatchEndpointToProduct(ctx, epInfo, store, path, nil)
	th.Assert(t, err == nil, err)
	// No healthIT product should have matched
	th.Assert(t, epInfo.HealthITProductID == 0, fmt.Sprintf("expected HealthITProductID value to be %d. Instead got %d", 0, epInfo.HealthITProductID))

	// capability statement
	path = filepath.Join("../../testdata", "allscripts_capability_dstu2.json")
	csJSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err = capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	var epOrg2 = &endpointmanager.FHIREndpointOrganization{
		OrganizationName: "Example2 Inc."}

	// populate fhir endpoint
	ep = &endpointmanager.FHIREndpoint{
		URL:              "example2.com/FHIR/DSTU2",
		OrganizationList: []*endpointmanager.FHIREndpointOrganization{epOrg2}}

	store.AddFHIREndpoint(ctx, ep)

	var epOrg3 = &endpointmanager.FHIREndpointOrganization{
		OrganizationName: "Example2 Inc."}

	// populate fhir endpoint with list source found in CHPL products info file
	ep2 := &endpointmanager.FHIREndpoint{
		URL:              "example3.com/FHIR/DSTU2",
		OrganizationList: []*endpointmanager.FHIREndpointOrganization{epOrg3},
		ListSource:       "https://api.bluebuttonpro.com/swagger/index.html"}

	store.AddFHIREndpoint(ctx, ep2)

	var epOrg4 = &endpointmanager.FHIREndpointOrganization{
		OrganizationName: "Example2 Inc."}

	ep3 := &endpointmanager.FHIREndpoint{
		URL:              "example4.com/FHIR/DSTU2",
		OrganizationList: []*endpointmanager.FHIREndpointOrganization{epOrg4},
		ListSource:       "https://nextgen.com/api/practice-search"}

	store.AddFHIREndpoint(ctx, ep3)

	// endpoint info
	epInfo = &endpointmanager.FHIREndpointInfo{
		URL:                 ep.URL,
		CapabilityStatement: cs}

	// endpoint info
	epInfo2 := &endpointmanager.FHIREndpointInfo{
		URL:                 ep2.URL,
		CapabilityStatement: nil}

	epInfo.HealthITProductID = 0

	// Case 2:
	// Capability-based matching via CHPL product mapping file
	err = MatchEndpointToProduct(ctx, epInfo, store, "../../testdata/test_chpl_product_mapping.json", nil)
	th.Assert(t, err == nil, err)
	healthITProductID, err := store.GetHealthITProductIDByCHPLID(ctx, "CorrectVersionAndName")
	th.Assert(t, err == nil, err)
	actualHealthITProductIDs, err := store.GetHealthITProductIDsByMapID(ctx, epInfo.HealthITProductID)
	th.Assert(t, err == nil, err)
	// healthIT product with ID healthITProductID should have matched
	th.Assert(t, actualHealthITProductIDs[0] == healthITProductID, fmt.Sprintf("expected HealthITProductID value to be %d. Instead got %d", healthITProductID, actualHealthITProductIDs[0]))

	epInfo2.HealthITProductID = 0

	// Case 3:
	// CapabilityStatement is intentionally nil to verify that MatchEndpointToProduct
	// maps products when explicit CHPL product IDs are provided
	err = MatchEndpointToProduct(ctx, epInfo2, store, "../../testdata/test_chpl_product_mapping.json", []string{"15.04.04.1322.Blue.02.00.0.200807"})
	th.Assert(t, err == nil, err)
	healthITProductID, err = store.GetHealthITProductIDByCHPLID(ctx, "15.04.04.1322.Blue.02.00.0.200807")
	th.Assert(t, err == nil, err)
	actualHealthITProductIDs, err = store.GetHealthITProductIDsByMapID(ctx, epInfo2.HealthITProductID)
	th.Assert(t, err == nil, err)
	// healthIT product with ID healthITProductID should have matched
	th.Assert(t, actualHealthITProductIDs[0] == healthITProductID, fmt.Sprintf("expected HealthITProductID value to be %d. Instead got %d", healthITProductID, actualHealthITProductIDs))

	// capability statement
	path = filepath.Join("../../testdata", "advantagecare_physicians_stu3.json")
	csJSON, err = os.ReadFile(path)
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

	epInfo.HealthITProductID = 0

	// Case 4:
	// Expected 3 matches:
	// 1 from capability-based matching (Epic → FakeCHPLID)
	// 2 from explicitly supplied NextGen product IDs.
	// Capability-derived and explicit product IDs are additive.
	err = MatchEndpointToProduct(ctx, epInfo, store, "../../testdata/test_chpl_product_mapping.json",
		[]string{
			"15.04.04.1918.Next.60.09.1.220303",
			"15.04.04.1918.Next.60.10.1.220318",
		})
	th.Assert(t, err == nil, err)
	actualHealthITProductIDs, err = store.GetHealthITProductIDsByMapID(ctx, epInfo.HealthITProductID)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(actualHealthITProductIDs) == 3, fmt.Sprintf("Expected endpoint to map to 3 healthIT products, instead mapped to %d", len(actualHealthITProductIDs)))

	epInfo3.HealthITProductID = 0

	// Case 5:
	// Expected 2 matches:
	// No capability statement is present, so only explicitly supplied
	// product IDs are used for matching.
	err = MatchEndpointToProduct(ctx, epInfo3, store, "../../testdata/test_chpl_product_mapping.json",
		[]string{
			"15.04.04.1918.Next.60.09.1.220303",
			"15.04.04.1918.Next.60.10.1.220318",
		})
	th.Assert(t, err == nil, err)
	actualHealthITProductIDs, err = store.GetHealthITProductIDsByMapID(ctx, epInfo3.HealthITProductID)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(actualHealthITProductIDs) == 2, fmt.Sprintf("Expected endpoint to map to 2 healthIT products, instead mapped to %d", len(actualHealthITProductIDs)))

	// Test matching to product by name and version

	var epOrg5 = &endpointmanager.FHIREndpointOrganization{
		OrganizationName: "Example Inc."}

	// populate fhir endpoint
	ep = &endpointmanager.FHIREndpoint{
		URL:              "example5.com/FHIR/DSTU2",
		OrganizationList: []*endpointmanager.FHIREndpointOrganization{epOrg5}}

	store.AddFHIREndpoint(ctx, ep)

	// capability statement with product HIEBus
	path = filepath.Join("../../testdata", "careevolution_dstu2.json")
	csJSON, err = os.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err = capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)

	// endpoint info
	epInfo = &endpointmanager.FHIREndpointInfo{
		URL:                 ep.URL,
		CapabilityStatement: cs}

	path = filepath.Join("../../testdata", "test_chpl_product_mapping.json")
	epInfo.HealthITProductID = 0

	// Case 6:
	// Capability-based matching using BOTH software name and software version.
	// Since version is present, matching is strict: only the active product with
	// name "HIEBus" and version "30.0.0" should be mapped.
	err = MatchEndpointToProduct(ctx, epInfo, store, path, nil)
	th.Assert(t, err == nil, err)
	actualHealthITProductIDs, err = store.GetHealthITProductIDsByMapID(ctx, epInfo.HealthITProductID)
	th.Assert(t, err == nil, err)

	// Should only be one match with the correct name and version
	th.Assert(t, len(actualHealthITProductIDs) == 1, fmt.Sprintf("Expected endpoint to map to 1 healthIT products, instead mapped to %d", len(actualHealthITProductIDs)))

	// Test matching to product by name and no version

	// populate fhir endpoint
	var epOrg6 = &endpointmanager.FHIREndpointOrganization{
		OrganizationName: "Example Inc."}

	ep = &endpointmanager.FHIREndpoint{
		URL:              "example6.com/FHIR/DSTU2",
		OrganizationList: []*endpointmanager.FHIREndpointOrganization{epOrg6}}
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

	epInfo.HealthITProductID = 0

	// Case 7: capability statement without software.version.
	// Matching falls back to name-only logic, which should associate
	// the endpoint with all ACTIVE products sharing that name.
	err = MatchEndpointToProduct(ctx, epInfo, store, path, nil)
	th.Assert(t, err == nil, err)
	actualHealthITProductIDs, err = store.GetHealthITProductIDsByMapID(ctx, epInfo.HealthITProductID)
	th.Assert(t, err == nil, err)

	// Should match to 2 products since there are only 2 active healthit products with the name HIEBus- version does not matter
	th.Assert(t, len(actualHealthITProductIDs) == 2, fmt.Sprintf("Expected endpoint to map to 2 healthIT products, instead mapped to %d", len(actualHealthITProductIDs)))
}

func setup() error {
	var err error
	store, err = postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))

	return err
}

func teardown() {
	store.Close()
}
