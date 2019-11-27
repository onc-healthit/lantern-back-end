package postgresql

import (
	"context"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/spf13/viper"
)

func Test_PersistFHIREndpoint(t *testing.T) {
	var err error
	ctx := context.Background()

	store, err := NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpass"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	if err != nil {
		t.Errorf("Error creating Store type: %s", err.Error())
	}
	defer store.Close()

	var endpoint1 = &endpointmanager.FHIREndpoint{
		URL:                   "example.com/FHIR/DSTU2",
		FHIRVersion:           "DSTU2",
		AuthorizationStandard: "OAuth 2.0",
		Location: &endpointmanager.Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		CapabilityStatement: &endpointmanager.CapabilityStatement{}}
	var endpoint2 = &endpointmanager.FHIREndpoint{
		URL:                   "other.example.com/FHIR/DSTU2",
		FHIRVersion:           "DSTU2",
		AuthorizationStandard: "R4 2.0"}

	// add endpoints

	err = store.AddFHIREndpoint(ctx, endpoint1)
	if err != nil {
		t.Errorf("Error adding fhir endpoint: %s", err.Error())
	}

	err = store.AddFHIREndpoint(ctx, endpoint2)
	if err != nil {
		t.Errorf("Error adding fhir endpoint: %s", err.Error())
	}

	// retrieve endpoints

	e1, err := store.GetFHIREndpoint(ctx, endpoint1.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpoint: %s", err.Error())
	}
	if !e1.Equal(endpoint1) {
		t.Errorf("retrieved endpoint is not equal to saved endpoint.")
	}

	e2, err := store.GetFHIREndpointUsingURL(ctx, endpoint2.URL)
	if err != nil {
		t.Errorf("Error getting fhir endpoint: %s", err.Error())
	}
	if !e2.Equal(endpoint2) {
		t.Errorf("retrieved endpoint is not equal to saved endpoint.")
	}

	// update endpoint

	e1.FHIRVersion = "Unknown"

	err = store.UpdateFHIREndpoint(ctx, e1)
	if err != nil {
		t.Errorf("Error updating fhir endpoint: %s", err.Error())
	}

	e1, err = store.GetFHIREndpoint(ctx, endpoint1.ID)
	if err != nil {
		t.Errorf("Error getting fhir endpoint: %s", err.Error())
	}
	if e1.Equal(endpoint1) {
		t.Errorf("retrieved UPDATED endpoint is equal to original endpoint.")
	}
	if e1.UpdatedAt.Equal(e1.CreatedAt) {
		t.Errorf("UpdatedAt is not being properly set on update.")
	}

	// delete endpoints

	err = store.DeleteFHIREndpoint(ctx, endpoint1)
	if err != nil {
		t.Errorf("Error deleting fhir endpoint: %s", err.Error())
	}

	_, err = store.GetFHIREndpoint(ctx, endpoint1.ID) // ensure we deleted the entry
	if err == nil {
		t.Errorf("endpoint1 was not deleted: %s", err.Error())
	}

	_, err = store.GetFHIREndpoint(ctx, endpoint2.ID) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("error retrieving endpoint2 after deleting endpoint1: %s", err.Error())
	}

	err = store.DeleteFHIREndpoint(ctx, endpoint2)
	if err != nil {
		t.Errorf("Error deleting fhir endpoint: %s", err.Error())
	}
}
