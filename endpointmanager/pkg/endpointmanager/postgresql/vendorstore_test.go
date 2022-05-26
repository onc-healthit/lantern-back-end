// +build integration

package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_PersistVendor(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

	// vendors
	epic := &endpointmanager.Vendor{
		Name:          "Epic Systems Corporation",
		DeveloperCode: "1447",
		CHPLID:        448,
		URL:           "http://www.epic.com",
		Location: &endpointmanager.Location{
			Address1: "1979 Milky Way",
			City:     "Verona",
			State:    "WI",
			ZipCode:  "53593"},
		Status:             "active",
		LastModifiedInCHPL: time.Date(2020, time.February, 24, 0, 0, 0, 0, time.UTC),
	}
	cerner := &endpointmanager.Vendor{
		Name:          "Cerner Corporation",
		DeveloperCode: "1221",
		CHPLID:        222,
		URL:           "http://www.cerner.com",
		Location: &endpointmanager.Location{
			Address1: "2800 Rockcreek Parkway",
			City:     "Kansas City",
			State:    "MO",
			ZipCode:  "64117"},
		Status:             "active",
		LastModifiedInCHPL: time.Date(2020, time.March, 25, 0, 0, 0, 0, time.UTC),
	}

	// add vendors

	err = store.AddVendor(ctx, epic)
	if err != nil {
		t.Errorf("Error adding fhir endpoint: %s", err.Error())
	}

	err = store.AddVendor(ctx, cerner)
	if err != nil {
		t.Errorf("Error adding fhir endpoint: %+v", err)
	}

	// retrieve endpoints

	v1, err := store.GetVendor(ctx, epic.ID)
	if err != nil {
		t.Errorf("Error getting vendor: %s", err.Error())
	}
	if !v1.Equal(epic) {
		t.Errorf("retrieved vendor is not equal to saved vendor.")
	}

	v2, err := store.GetVendorUsingName(ctx, cerner.Name)
	if err != nil {
		t.Errorf("Error getting vendor: %s", err.Error())
	}
	if !v2.Equal(cerner) {
		t.Errorf("retrieved vendor is not equal to saved vendor.")
	}
	v3, err := store.GetVendorUsingCHPLID(ctx, cerner.CHPLID)
	if err != nil {
		t.Errorf("Error getting vendor: %s", err.Error())
	}
	if !v3.Equal(cerner) {
		t.Errorf("retrieved vendor is not equal to saved vendor.")
	}

	// update vendor

	v1.URL = "www.example.com"

	err = store.UpdateVendor(ctx, v1)
	if err != nil {
		t.Errorf("Error updating vendor: %s", err.Error())
	}

	v1updated, err := store.GetVendor(ctx, epic.ID)
	if err != nil {
		t.Errorf("Error getting vendor: %s", err.Error())
	}
	if v1updated.Equal(epic) {
		t.Errorf("retrieved UPDATED vendor is equal to original vendor.")
	}
	if v1updated.UpdatedAt.Equal(v1updated.CreatedAt) {
		t.Errorf("UpdatedAt is not being properly set on update.")
	}

	// get vendor names

	vendorNames, err := store.GetVendorNames(ctx)
	if err != nil {
		t.Errorf("Error getting vendor names: %s", err.Error())
	}
	eLength := 2
	if len(vendorNames) != eLength {
		t.Errorf("Expected vendor name list to have length %d. Got %d.", eLength, len(vendorNames))
	}

	for _, v := range vendorNames {
		if v != epic.Name && v != cerner.Name {
			t.Errorf("List contains a name other than %s or %s: %s.", epic.Name, cerner.Name, v)
		}
	}

	// delete vendors

	err = store.DeleteVendor(ctx, epic)
	if err != nil {
		t.Errorf("Error deleting vendor: %s", err.Error())
	}

	_, err = store.GetVendor(ctx, epic.ID) // ensure we deleted the entry
	if err == nil {
		t.Errorf("vendor was not deleted")
	}

	_, err = store.GetVendor(ctx, cerner.ID) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("error retrieving cerner vendor after deleting epic vendor: %s", err.Error())
	}

	err = store.DeleteVendor(ctx, cerner)
	if err != nil {
		t.Errorf("Error deleting vendor: %s", err.Error())
	}
}
