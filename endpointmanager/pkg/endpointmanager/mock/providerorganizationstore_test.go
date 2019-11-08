package mock

import (
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

func Test_PersistProviderOrganization(t *testing.T) {
	var err error

	store, err := NewStore(host, port, user, password, dbname, sslmode)
	if err != nil {
		t.Errorf("Error creating Store type: %s", err.Error())
	}
	defer store.Close()

	var po1 = &endpointmanager.ProviderOrganization{
		Name: "Hospital #1 of America",
		URL:  "hospital.example.com",
		Location: &endpointmanager.Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		OrganizationType: "hospital",
		HospitalType:     "Acute Care",
		Ownership:        "Volunary non-profit",
		Beds:             250}
	var po2 = &endpointmanager.ProviderOrganization{
		Name:             "Group Practice #1 of America",
		URL:              "grouppractice.example.com",
		OrganizationType: "group practice",
		HospitalType:     "",
		Ownership:        "",
		Beds:             -1}

	// add organizations

	err = store.AddProviderOrganization(po1)
	if err != nil {
		t.Errorf("Error adding provider organization: %s", err.Error())
	}

	err = store.AddProviderOrganization(po2)
	if err != nil {
		t.Errorf("Error adding provider organization: %s", err.Error())
	}

	// retrieve organizations

	p1, err := store.GetProviderOrganization(po1.ID)
	if err != nil {
		t.Errorf("Error getting provider organization: %s", err.Error())
	}
	if !p1.Equal(po1) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	p2, err := store.GetProviderOrganization(po2.ID)
	if err != nil {
		t.Errorf("Error getting provider organization: %s", err.Error())
	}
	if !p2.Equal(po2) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	// update organization

	p1.HospitalType = "Critical Access"

	err = store.UpdateProviderOrganization(p1)
	if err != nil {
		t.Errorf("Error updating provider organization: %s", err.Error())
	}

	p1, err = store.GetProviderOrganization(po1.ID)
	if err != nil {
		t.Errorf("Error getting provider organization: %s", err.Error())
	}
	if p1.Equal(po1) {
		t.Errorf("retrieved UPDATED organization is equal to original organization.")
	}
	if p1.UpdatedAt.Equal(p1.CreatedAt) {
		t.Errorf("UpdatedAt is not being properly set on update.")
	}

	// delete organizations

	err = store.DeleteProviderOrganization(po1)
	if err != nil {
		t.Errorf("Error deleting provider organization: %s", err.Error())
	}

	_, err = store.GetProviderOrganization(po1.ID) // ensure we deleted the entry
	if err == nil {
		t.Errorf("po1 was not deleted: %s", err.Error())
	}

	_, err = store.GetProviderOrganization(po2.ID) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("error retrieving po2 after deleting po1: %s", err.Error())
	}

	err = store.DeleteProviderOrganization(po2)
	if err != nil {
		t.Errorf("Error deleting provider organization: %s", err.Error())
	}
}
