package postgresql

import (
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/spf13/viper"
)

func Test_PersistNPIOrganization(t *testing.T) {
	var err error

	store, err := NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpass"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	if err != nil {
		t.Errorf("Error creating Store type: %s", err.Error())
	}
	defer store.Close()

	var npio1 = &endpointmanager.NPIOrganization{
		ID:            1,
		NPI_ID:        "1",
		Name:          "Hospital #1 of America",
		SecondaryName: "Hospital #1 of America Second Name",
		Location: &endpointmanager.Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		Taxonomy: "208D00000X"}

	var npio2 = &endpointmanager.NPIOrganization{
		ID:            1,
		NPI_ID:        "",
		Name:          "",
		SecondaryName: "",
		Location:      &endpointmanager.Location{},
		Taxonomy:      "208D00000X"}

	// add organizations

	err = store.AddNPIOrganization(npio1)
	if err != nil {
		t.Errorf("Error adding npi organization: %s", err.Error())
	}

	err = store.AddNPIOrganization(npio2)
	if err != nil {
		t.Errorf("Error adding npi organization: %s", err.Error())
	}

	// retrieve organizations

	npio1_get, err := store.GetNPIOrganization(npio1.ID)
	if err != nil {
		t.Errorf("Error getting npi organization: %s", err.Error())
	}
	if !npio1_get.Equal(npio1) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	npio2_get, err := store.GetNPIOrganization(npio2.ID)
	if err != nil {
		t.Errorf("Error getting npi organization: %s", err.Error())
	}
	if !npio2_get.Equal(npio2) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	// retrieve organizations by NPI_ID

	npio1_get_npi, err := store.GetNPIOrganizationByNPIID(npio1.NPI_ID)
	if err != nil {
		t.Errorf("Error getting npi organization: %s", err.Error())
	}
	if !npio1_get_npi.Equal(npio1) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	npio2_get_npi, err := store.GetNPIOrganizationByNPIID(npio2.NPI_ID)
	if err != nil {
		t.Errorf("Error getting npi organization: %s", err.Error())
	}
	if !npio2_get_npi.Equal(npio2) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	// update organization

	temp_taxonomy := npio1.Taxonomy
	npio1.Taxonomy = "1234567"

	err = store.UpdateNPIOrganization(npio1)
	if err != nil {
		t.Errorf("Error updating npi organization: %s", err.Error())
	}

	// Restore taxonomy
	npio1.Taxonomy = temp_taxonomy

	npio1_get, err = store.GetNPIOrganization(npio1.ID)
	if err != nil {
		t.Errorf("Error getting npi organization: %s", err.Error())
	}
	if npio1_get.Equal(npio1) {
		t.Errorf("retrieved UPDATED organization is equal to original organization.")
	}
	if npio1_get.UpdatedAt.Equal(npio1.CreatedAt) {
		t.Errorf("UpdatedAt is not being properly set on update.")
	}

	// update organization using UpdateNPIOrganizationByNPIID

	temp_taxonomy = npio1.Taxonomy
	npio1.Taxonomy = "1234567"

	err = store.UpdateNPIOrganizationByNPIID(npio1)
	if err != nil {
		t.Errorf("Error updating npi organization: %s", err.Error())
	}

	// Restore taxonomy
	npio1.Taxonomy = temp_taxonomy

	npio1_get, err = store.GetNPIOrganization(npio1.ID)
	if err != nil {
		t.Errorf("Error getting npi organization: %s", err.Error())
	}
	if npio1_get.Equal(npio1) {
		t.Errorf("retrieved UPDATED organization is equal to original organization.")
	}
	if npio1_get.UpdatedAt.Equal(npio1.CreatedAt) {
		t.Errorf("UpdatedAt is not being properly set on update.")
	}

	// delete organizations

	err = store.DeleteNPIOrganization(npio1)
	if err != nil {
		t.Errorf("Error deleting npi organization: %s", err.Error())
	}

	_, err = store.GetNPIOrganization(npio1.ID) // ensure we deleted the entry
	if err == nil {
		t.Errorf("npio1 was not deleted: %s", err.Error())
	}

	_, err = store.GetNPIOrganization(npio2.ID) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("error retrieving npio2 after deleting npio1: %s", err.Error())
	}

	err = store.DeleteNPIOrganization(npio2)
	if err != nil {
		t.Errorf("Error deleting npi organization: %s", err.Error())
	}
}
