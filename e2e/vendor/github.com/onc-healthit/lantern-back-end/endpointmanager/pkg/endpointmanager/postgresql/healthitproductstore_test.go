// +build integration

package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_PersistHealthITProduct(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

	var hitp1 = &endpointmanager.HealthITProduct{
		Name:      "Health IT System 1",
		Version:   "1.0",
		Developer: "Epic",
		Location: &endpointmanager.Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		AuthorizationStandard: "OAuth 2.0",
		APISyntax:             "FHIR R4",
		APIURL:                "example.com",
		CertificationCriteria: []string{"criteria1", "criteria2"},
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2019, 10, 19, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2015",
		LastModifiedInCHPL:    time.Date(2019, 10, 19, 0, 0, 0, 0, time.UTC),
		CHPLID:                "ID"}
	var hitp2 = &endpointmanager.HealthITProduct{
		Name:                 "Health IT System 2",
		Version:              "2.0",
		Developer:            "Cerner",
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

	h1s, err := store.GetHealthITProductsUsingVendor(ctx, "Epic")
	if err != nil {
		t.Errorf("Error getting health it product: %s", err.Error())
	}
	if len(h1s) != 1 {
		t.Errorf("Expected to retrieve 1 entry from DB. Retrieved %d.", len(h1s))
	}
	if !h1s[0].Equal(hitp1) {
		t.Errorf("retrieved product is not equal to saved product.")
	}

	h2s, err := store.GetHealthITProductsUsingVendor(ctx, "Cerner")
	if err != nil {
		t.Errorf("Error getting health it product: %s", err.Error())
	}
	if len(h2s) != 1 {
		t.Errorf("Expected to retrieve 1 entry from DB. Retrieved %d.", len(h2s))
	}
	if !h2s[0].Equal(hitp2) {
		t.Errorf("retrieved product is not equal to saved product.")
	}

	// get developer list
	devs, err := store.GetHealthITProductDevelopers(ctx)
	if err != nil {
		t.Errorf("Error getting developer list: %s", err.Error())
	}
	if len(devs) != 2 {
		t.Error("Expected developer list to have two entries")
	}
	if !contains(devs, "Epic") {
		t.Error("Expected developer list to contain 'Epic'")
	}
	if !contains(devs, "Cerner") {
		t.Error("Expected developer list to contain 'Epic'")
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

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
