// +build integration

package postgresql

import (
	"context"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_PersistFHIREndpoint(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

	// endpoints
	var endpoint1 = &endpointmanager.FHIREndpoint{
		URL:              "example.com/FHIR/DSTU2/",
		OrganizationName: "Example Inc.",
		ListSource:       "https://github.com/cerner/ignite-endpoints"}
	var endpoint2 = &endpointmanager.FHIREndpoint{
		URL:              "other.example.com/FHIR/DSTU2/",
		OrganizationName: "Other Example Inc."}

	// add endpoints

	err = store.AddFHIREndpoint(ctx, endpoint1)
	if err != nil {
		t.Errorf("Error adding fhir endpoint: %s", err.Error())
	}

	err = store.AddFHIREndpoint(ctx, endpoint2)
	if err != nil {
		t.Errorf("Error adding fhir endpoint: %+v", err)
	}

	// retrieve endpoints

	e1, err1 := store.GetFHIREndpointUsingURL(ctx, endpoint1.URL)
	if err1 != nil {
		t.Errorf("Error getting fhir endpoint: %s", err1.Error())
	}
	if !e1.Equal(endpoint1) {
		t.Errorf("retrieved endpoint is not equal to saved endpoint.")
	}

	eID1, err2 := store.GetFHIREndpoint(ctx, e1.ID)
	if err2 != nil {
		t.Errorf("Error getting fhir endpoint: %s", err2.Error())
	}
	if !eID1.Equal(endpoint1) {
		t.Errorf("retrieved endpoint is not equal to saved endpoint.")
	}

	// retreive all endpoints
	endpts, err := store.GetAllFHIREndpoints(ctx)
	if err != nil {
		t.Errorf("Error getting fhir endpoints: %s", err1.Error())
	}
	eLen := 2
	if len(endpts) != eLen {
		t.Errorf("number of retrieved endpoints is not equal to number of saved endpoints")
	}

	// update endpoint

	e1.ListSource = "Unknown"

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

	// get all org names

	endpointNames, err := store.GetAllFHIREndpointOrgNames(ctx)
	if err != nil {
		t.Errorf("Error getting endpoint organization normalized names: %s", err.Error())
	}
	eLength := 2
	if len(endpointNames) != eLength {
		t.Errorf("Expected endpoint org list to have length %d. Got %d.", eLength, len(endpointNames))
	}

	for _, ep := range endpointNames {
		if ep.ID == endpoint1.ID {
			eName := "Example Inc."
			if ep.OrganizationName != eName {
				t.Errorf("Expected org name to be %s. Got %s.", eName, ep.OrganizationName)
			}
		}
		if ep.ID == endpoint2.ID {
			eName := "Other Example Inc."
			if ep.OrganizationName != eName {
				t.Errorf("Expected org name to be %s. Got %s.", eName, ep.OrganizationName)
			}
		}
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
