// +build integration

package postgresql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

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
}

func Test_PersistHealthITProduct(t *testing.T) {
	SetupStore()

	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

	for _, vendor := range vendors {
		store.AddVendor(ctx, vendor)
	}

	var hitp1 = &endpointmanager.HealthITProduct{
		Name:     "Health IT System 1",
		Version:  "1.0",
		VendorID: vendors[0].ID, // epic
		Location: &endpointmanager.Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		AuthorizationStandard: "OAuth 2.0",
		APISyntax:             "FHIR R4",
		APIURL:                "example.com",
		CertificationCriteria: []int{31, 32},
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2019, 10, 19, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2015",
		LastModifiedInCHPL:    time.Date(2019, 10, 19, 0, 0, 0, 0, time.UTC),
		CHPLID:                "ID"}
	var hitp2 = &endpointmanager.HealthITProduct{
		Name:                 "Health IT System 2",
		Version:              "2.0",
		VendorID:             vendors[1].ID, // cerner
		APISyntax:            "FHIR DSTU2",
		CertificationEdition: "2014"}
	// add products

	err = store.AddHealthITProduct(ctx, hitp1)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}

	err = store.AddHealthITProduct(ctx, hitp2)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}

	// retrieve products

	h1, err := store.GetHealthITProduct(ctx, hitp1.ID)
	if err != nil {
		t.Errorf("Error getting health it product: %s", err.Error())
	}
	if !h1.Equal(hitp1) {
		t.Errorf("retrieved product is not equal to saved product.")
	}

	h2, err := store.GetHealthITProductUsingNameAndVersion(ctx, hitp2.Name, hitp2.Version)
	if err != nil {
		t.Errorf("Error getting health it product: %s", err.Error())
	}
	if !h2.Equal(hitp2) {
		t.Errorf("retrieved product is not equal to saved product.")
	}

	// retrieve products using vendor

	h1s, err := store.GetHealthITProductsUsingVendor(ctx, vendors[0].ID)
	if err != nil {
		t.Errorf("Error getting health it product: %s", err.Error())
	}
	if len(h1s) != 1 {
		t.Errorf("Expected to retrieve 1 entry from DB. Retrieved %d.", len(h1s))
	}
	if !h1s[0].Equal(hitp1) {
		t.Errorf("retrieved product is not equal to saved product.")
	}

	h2s, err := store.GetHealthITProductsUsingVendor(ctx, vendors[1].ID)
	if err != nil {
		t.Errorf("Error getting health it product: %s", err.Error())
	}
	if len(h2s) != 1 {
		t.Errorf("Expected to retrieve 1 entry from DB. Retrieved %d.", len(h2s))
	}
	if !h2s[0].Equal(hitp2) {
		t.Errorf("retrieved product is not equal to saved product.")
	}

	// update product

	h1.APISyntax = "FHIR R5"

	err = store.UpdateHealthITProduct(ctx, h1)
	if err != nil {
		t.Errorf("Error updating health it product: %s", err.Error())
	}

	h1, err = store.GetHealthITProduct(ctx, hitp1.ID)
	if err != nil {
		t.Errorf("Error getting health it product: %s", err.Error())
	}
	if h1.Equal(hitp1) {
		t.Errorf("retrieved UPDATED product is equal to original product.")
	}
	if h1.UpdatedAt.Equal(h1.CreatedAt) {
		t.Errorf("UpdatedAt is not being properly set on update.")
	}

	// delete products

	err = store.DeleteHealthITProduct(ctx, hitp1)
	if err != nil {
		t.Errorf("Error deleting health it product: %s", err.Error())
	}

	_, err = store.GetHealthITProduct(ctx, hitp1.ID) // ensure we deleted the entry
	if err == nil {
		t.Errorf("hitp1 was not deleted: %s", err.Error())
	}

	_, err = store.GetHealthITProduct(ctx, hitp2.ID) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("error retrieving hitp2 after deleting hitp1: %s", err.Error())
	}

	err = store.DeleteHealthITProduct(ctx, hitp2)
	if err != nil {
		t.Errorf("Error deleting health it product: %s", err.Error())
	}
}

func Test_LinkProductToCriteria(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

	for _, vendor := range vendors {
		store.AddVendor(ctx, vendor)
	}

	// products
	var hitp1 = &endpointmanager.HealthITProduct{
		Name:     "Health IT System 1",
		Version:  "1.0",
		VendorID: vendors[0].ID, // epic
		Location: &endpointmanager.Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		AuthorizationStandard: "OAuth 2.0",
		APISyntax:             "FHIR R4",
		APIURL:                "example.com",
		CertificationCriteria: []int{44},
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2019, 10, 19, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2015",
		LastModifiedInCHPL:    time.Date(2019, 10, 19, 0, 0, 0, 0, time.UTC),
		CHPLID:                "ID"}
	var hitp2 = &endpointmanager.HealthITProduct{
		Name:                  "Health IT System 2",
		Version:               "2.0",
		VendorID:              vendors[1].ID, // cerner
		APISyntax:             "FHIR DSTU2",
		CertificationCriteria: []int{64},
		CertificationEdition:  "2014"}

	// criteria
	var crit1 = &endpointmanager.CertificationCriteria{
		CertificationID:        44,
		CertificationNumber:    "170.315 (f)(2)",
		Title:                  "Transmission to Public Health Agencies - Syndromic Surveillance",
		CertificationEditionID: 3,
		CertificationEdition:   "2015",
		Description:            "Syndromic Surveillance",
		Removed:                false,
	}
	var crit2 = &endpointmanager.CertificationCriteria{
		CertificationID:        64,
		CertificationNumber:    "170.314 (a)(4)",
		Title:                  "Vital signs, body mass index, and growth Charts",
		CertificationEditionID: 2,
		CertificationEdition:   "2014",
		Description:            "Vital signs",
		Removed:                false,
	}

	err = store.AddHealthITProduct(ctx, hitp1)
	th.Assert(t, err == nil, fmt.Errorf("Error adding health it product: %s", err))

	err = store.AddCriteria(ctx, crit1)
	th.Assert(t, err == nil, fmt.Errorf("Error adding criteria: %s", err))

	err = store.LinkProductToCriteria(ctx, crit1.CertificationID, hitp1.ID, crit1.CertificationNumber)
	th.Assert(t, err == nil, fmt.Errorf("Error linking product to criteria: %s", err))

	err = store.AddHealthITProduct(ctx, hitp2)
	th.Assert(t, err == nil, fmt.Errorf("Error adding health it product: %s", err))

	err = store.AddCriteria(ctx, crit2)
	th.Assert(t, err == nil, fmt.Errorf("Error adding criteria: %s", err))

	err = store.LinkProductToCriteria(ctx, crit2.CertificationID, hitp2.ID, crit2.CertificationNumber)
	th.Assert(t, err == nil, fmt.Errorf("Error linking product to criteria: %s", err))

	var count int
	row := store.DB.QueryRow("SELECT COUNT(*) FROM product_criteria")
	err = row.Scan(&count)
	th.Assert(t, err == nil, fmt.Errorf("Error getting rows from product_criteria: %s", err))
	th.Assert(t, count == 2, "Expected two rows in DB")

	retProdID, retCritID, retCritNum, err := store.GetProductCriteriaLink(ctx, crit1.CertificationID, hitp1.ID)
	th.Assert(t, err == nil, err)
	th.Assert(t, retProdID == hitp1.ID, fmt.Sprintf("expected stored ID '%d' to be the same as the ID that was stored '%d'.", retProdID, hitp1.ID))
	th.Assert(t, retCritID == crit1.CertificationID, fmt.Sprintf("expected stored ID '%d' to be the same as the ID that was stored '%d'.", retCritID, crit1.CertificationID))
	th.Assert(t, retCritNum == "170.315 (f)(2)", fmt.Sprintf("expected stored confidence '%s' to be the same as the confidence that was stored '170.315 (f)(2)'.", retCritNum))

	retProdID, retCritID, retCritNum, err = store.GetProductCriteriaLink(ctx, crit2.CertificationID, hitp2.ID)
	th.Assert(t, err == nil, err)
	th.Assert(t, retProdID == hitp2.ID, fmt.Sprintf("expected stored ID '%d' to be the same as the ID that was stored '%d'.", retProdID, hitp2.ID))
	th.Assert(t, retCritID == crit2.CertificationID, fmt.Sprintf("expected stored ID '%d' to be the same as the ID that was stored '%d'.", retCritID, crit2.CertificationID))
	th.Assert(t, retCritNum == "170.314 (a)(4)", fmt.Sprintf("expected stored confidence '%s' to be the same as the confidence that was stored '170.314 (a)(4)'.", retCritNum))
}
