// +build integration

package postgresql

import (
	"context"
	"fmt"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/versionsoperatorparser"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_PersistFHIREndpoint(t *testing.T) {
	SetupStore()
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

	// endpoints
	var endpoint1 = &endpointmanager.FHIREndpoint{
		URL:               "example.com/FHIR/DSTU2/",
		OrganizationNames: []string{"Example Inc."},
		NPIIDs:            []string{"1"},
		ListSource:        "https://github.com/cerner/ignite-endpoints"}

	var endpoint2 = &endpointmanager.FHIREndpoint{
		URL:               "other.example.com/FHIR/DSTU2/",
		OrganizationNames: []string{"Other Example Inc."}}

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

	e1, err1 := store.GetFHIREndpointUsingURLAndListSource(ctx, endpoint1.URL, endpoint1.ListSource)
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

	// update endpoint
	e1.ListSource = "Unknown"

	var vsr versionsoperatorparser.VersionsResponse
	vsr.Response = make(map[string]interface{})
	vsr.Response["default"] = "4.0"
	vsr.Response["versions"] = []string{"4.0"}
	e1.VersionsResponse = vsr

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

	// add or update endpoint
	e1.ListSource = "New List Source"
	err = store.AddOrUpdateFHIREndpoint(ctx, e1)
	if err != nil {
		t.Errorf("Error adding/updating fhir endpoint: %s", err.Error())
	}
	e1, err = store.GetFHIREndpointUsingURLAndListSource(ctx, e1.URL, e1.ListSource)
	if err != nil {
		t.Errorf("Error getting fhir endpoint: %s", err.Error())
	}
	if e1.ID == endpoint1.ID {
		t.Errorf("should have created a new entry")
	}

	e1.OrganizationNames = []string{"Org 1", "Org 2"}
	e1.NPIIDs = []string{"2", "3"}
	var modvsr versionsoperatorparser.VersionsResponse
	modvsr.Response = make(map[string]interface{})
	modvsr.Response["default"] = "4.0"
	modvsr.Response["versions"] = []string{"4.0","2.0"}
	e1.VersionsResponse = modvsr
	err = store.AddOrUpdateFHIREndpoint(ctx, e1)
	if err != nil {
		t.Errorf("Error adding/updating fhir endpoint: %s", err.Error())
	}
	e1, err = store.GetFHIREndpoint(ctx, e1.ID)
	if err != nil {
		t.Errorf("Error adding/updating fhir endpoint: %s", err.Error())
	}
	if !helpers.StringArraysEqual(e1.OrganizationNames, []string{"Org 1", "Org 2", "Example Inc."}) {
		t.Errorf("Expected organization names array to be merged with new org names")
	}
	if !helpers.StringArraysEqual(e1.NPIIDs, []string{"1", "2", "3"}) {
		t.Errorf("Expected NPI IDs array to be merged with new NPI IDs")
	}
	if !e1.VersionsResponse.Equal(modvsr) {
		fmt.Println(e1.VersionsResponse)
		fmt.Println(modvsr)
		t.Errorf("Expected VersionsResponse to be updated with new value")
	}

	// retreive all endpoints

	endpts, err := store.GetAllFHIREndpoints(ctx)
	if err != nil {
		t.Errorf("Error getting fhir endpoints: %s", err1.Error())
	}
	eLen := 3
	if len(endpts) != eLen {
		t.Errorf("number of retrieved endpoints is not equal to number of saved endpoints")
	}

	for _, ep := range endpts {
		if ep.ID == endpoint1.ID {
			eName := []string{"Example Inc."}
			if !helpers.StringArraysEqual(ep.OrganizationNames, eName) {
				t.Errorf("Expected org name to be %v. Got %v.", eName, ep.OrganizationNames)
			}
		}
		if ep.ID == endpoint2.ID {
			eName := []string{"Other Example Inc."}
			if !helpers.StringArraysEqual(ep.OrganizationNames, eName) {
				t.Errorf("Expected org name to be %v. Got %v.", eName, ep.OrganizationNames)
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
		t.Errorf("endpoint1 was not deleted")
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
