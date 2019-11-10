package postgresql

import (
	"testing"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/spf13/viper"
)

func Test_PersistHealthITProduct(t *testing.T) {
	var err error

	store, err := NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpass"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	if err != nil {
		t.Errorf("Error creating Store type: %s", err.Error())
	}
	defer store.Close()

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

	err = store.AddHealthITProduct(hitp1)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}

	err = store.AddHealthITProduct(hitp2)
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}

	// retrieve products

	h1, err := store.GetHealthITProduct(hitp1.ID)
	if err != nil {
		t.Errorf("Error getting health it product: %s", err.Error())
	}
	if !h1.Equal(hitp1) {
		t.Errorf("retrieved product is not equal to saved product.")
	}

	h2, err := store.GetHealthITProductUsingNameAndVersion(hitp2.Name, hitp2.Version)
	if err != nil {
		t.Errorf("Error getting health it product: %s", err.Error())
	}
	if !h2.Equal(hitp2) {
		t.Errorf("retrieved product is not equal to saved product.")
	}

	// update product

	h1.APISyntax = "FHIR R5"

	err = store.UpdateHealthITProduct(h1)
	if err != nil {
		t.Errorf("Error updating health it product: %s", err.Error())
	}

	h1, err = store.GetHealthITProduct(hitp1.ID)
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

	err = store.DeleteHealthITProduct(hitp1)
	if err != nil {
		t.Errorf("Error deleting health it product: %s", err.Error())
	}

	_, err = store.GetHealthITProduct(hitp1.ID) // ensure we deleted the entry
	if err == nil {
		t.Errorf("hitp1 was not deleted: %s", err.Error())
	}

	_, err = store.GetHealthITProduct(hitp2.ID) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("error retrieving hitp2 after deleting hitp1: %s", err.Error())
	}

	err = store.DeleteHealthITProduct(hitp2)
	if err != nil {
		t.Errorf("Error deleting health it product: %s", err.Error())
	}
}
