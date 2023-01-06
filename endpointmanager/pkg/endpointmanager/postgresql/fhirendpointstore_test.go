// +build integration

package postgresql

import (
	"context"
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

	var endpointOrganization1 = &endpointmanager.FHIREndpointOrganization{
		OrganizationName: "Example Inc.",
		OrganizationNPIID: "1"}

	var endpointOrganization2 = &endpointmanager.FHIREndpointOrganization{
		OrganizationName: "Other Example Inc."}

	// endpoints
	var endpoint1 = &endpointmanager.FHIREndpoint{
		URL:               "https://example.com/FHIR/DSTU2/",
		OrganizationList: []*endpointmanager.FHIREndpointOrganization{endpointOrganization1},
		ListSource:        "https://github.com/cerner/ignite-endpoints"}

	var endpoint2 = &endpointmanager.FHIREndpoint{
		URL:               "https://other.example.com/FHIR/DSTU2/",
		OrganizationList: []*endpointmanager.FHIREndpointOrganization{endpointOrganization2}}

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

	var org1 = &endpointmanager.FHIREndpointOrganization{
		OrganizationName: "Org 1",
		OrganizationNPIID: "2"}

	var org2 = &endpointmanager.FHIREndpointOrganization{
		OrganizationName: "Org 2",
		OrganizationNPIID: "3"}

	e1.OrganizationList = []*endpointmanager.FHIREndpointOrganization{org1, org2}
	vsr.Response["versions"] = []string{"4.0", "2.0"}
	e1.VersionsResponse = vsr
	err = store.AddOrUpdateFHIREndpoint(ctx, e1)
	if err != nil {
		t.Errorf("Error adding/updating fhir endpoint: %s", err.Error())
	}
	e1, err = store.GetFHIREndpoint(ctx, e1.ID)
	if err != nil {
		t.Errorf("Error adding/updating fhir endpoint: %s", err.Error())
	}
	organizationNameList := e1.GetOrganizationNames()
	NPIIDsList := e1.GetNPIIDs()

	if !helpers.StringArraysEqual(organizationNameList, []string{"Org 1", "Org 2", "Example Inc."}) {
		t.Errorf("Expected organization names array to be merged with new org names. Actual: %v", organizationNameList)
	}
	if !helpers.StringArraysEqual(NPIIDsList, []string{"1", "2", "3"}) {
		t.Errorf("Expected NPI IDs array to be merged with new NPI IDs")
	}
	if !e1.VersionsResponse.Equal(vsr) {
		t.Errorf("Expected VersionsResponse %v to be updated with new value so that it equals %v", e1.VersionsResponse, vsr)
	}

	// update endpoint NPI Org

	var NPIOrg = &endpointmanager.FHIREndpointOrganization{
		OrganizationName: "Fake Organization",
		OrganizationNPIID: "123"}

	var fhirEndpointNPIOrg = &endpointmanager.FHIREndpoint{
		URL:               "https://example.com/FHIR/DSTU2/",
		OrganizationList: []*endpointmanager.FHIREndpointOrganization{NPIOrg},
	}

	// Add new organization name and NPI ID to fhir_endpoints table
	err = store.UpdateFHIREndpointsNPIOrg(ctx, fhirEndpointNPIOrg, true)

	endpointArr, err := store.GetFHIREndpointUsingURL(ctx, fhirEndpointNPIOrg.URL)

	endptsLen := 2
	if len(endpointArr) != endptsLen {
		t.Errorf("number of retrieved endpoints %v is not equal to number of saved endpoints %v", len(endpointArr), endptsLen)
	}

	// Update should have applied to all endpoints in fhir_endpoints table with URL example.com/FHIR/DSTU2/
	for _, elem := range endpointArr {
		if elem.ID == e1.ID {
			expectedOrganizationNames := []string{"Fake Organization", "Org 1", "Org 2", "Example Inc."}
			expectedNPIIDs := []string{"1", "2", "3", "123"}

			elemOrganizationsList := elem.GetOrganizationNames()
			elemNPIIDList := elem.GetNPIIDs()
			if !helpers.StringArraysEqual(elemOrganizationsList, expectedOrganizationNames) {
				t.Errorf("Expected organization names array to be merged with new org names. Expected: %v, Actual: %v", expectedOrganizationNames, elemOrganizationsList)
			}
			if !helpers.StringArraysEqual(elemNPIIDList, expectedNPIIDs) {
				t.Errorf("Expected NPI IDs array to be merged with new NPI IDs. Expected: %v, Actual: %v", expectedNPIIDs, elemNPIIDList)
			}

		} else if elem.ID == endpoint1.ID {
			expectedOrganizationNames := []string{"Fake Organization", "Example Inc."}
			expectedNPIIDs := []string{"1", "123"}

			elemOrganizationsList := elem.GetOrganizationNames()
			elemNPIIDsList := elem.GetNPIIDs()

			if !helpers.StringArraysEqual(elemOrganizationsList, expectedOrganizationNames) {
				t.Errorf("Expected organization names array to be merged with new org names. Expected: %v, Actual: %v", expectedOrganizationNames, elemOrganizationsList)
			}
			if !helpers.StringArraysEqual(elemNPIIDsList, expectedNPIIDs) {
				t.Errorf("Expected NPI IDs array to be merged with new NPI IDs. Expected: %v, Actual: %v", expectedNPIIDs, elemNPIIDsList)
			}

		} else {
			t.Errorf("Retrieved unexpected endpoint")
		}
	}

	// Remove new organization names and NPI IDs

	err = store.UpdateFHIREndpointsNPIOrg(ctx, fhirEndpointNPIOrg, false)

	endpointArr, err = store.GetFHIREndpointUsingURL(ctx, fhirEndpointNPIOrg.URL)

	if len(endpointArr) != endptsLen {
		t.Errorf("number of retrieved endpoints %v is not equal to number of saved endpoints %v", len(endpointArr), endptsLen)
	}

	for _, elem := range endpointArr {
		if elem.ID == e1.ID {
			expectedOrganizationNames := []string{"Org 1", "Org 2", "Example Inc."}
			expectedNPIIDs := []string{"1", "2", "3"}

			elemOrganizationsList := elem.GetOrganizationNames()
			elemNPIIDList := elem.GetNPIIDs()

			if !helpers.StringArraysEqual(elemOrganizationsList, expectedOrganizationNames) {
				t.Errorf("Expected organization names array to be merged with new org names. Expected: %v, Actual: %v", expectedOrganizationNames, elemOrganizationsList)
			}
			if !helpers.StringArraysEqual(elemNPIIDList, expectedNPIIDs) {
				t.Errorf("Expected NPI IDs array to be merged with new NPI IDs. Expected: %v, Actual: %v", expectedNPIIDs, elemNPIIDList)
			}

		} else if elem.ID == endpoint1.ID {
			expectedOrganizationNames := []string{"Example Inc."}
			expectedNPIIDs := []string{"1"}

			elemOrganizationsList := elem.GetOrganizationNames()
			elemNPIIDList := elem.GetNPIIDs()

			if !helpers.StringArraysEqual(elemOrganizationsList, expectedOrganizationNames) {
				t.Errorf("Expected organization names array to be merged with new org names. Expected: %v, Actual: %v", expectedOrganizationNames, elemOrganizationsList)
			}
			if !helpers.StringArraysEqual(elemNPIIDList, expectedNPIIDs) {
				t.Errorf("Expected NPI IDs array to be merged with new NPI IDs. Expected: %v, Actual: %v", expectedNPIIDs, elemNPIIDList)
			}

		} else {
			t.Errorf("Retrieved unexpected endpoint")
		}
	}

	// retreive all endpoints

	endpts, err := store.GetAllFHIREndpoints(ctx)
	if err != nil {
		t.Errorf("Error getting fhir endpoints: %s", err1.Error())
	}
	eLen := 3
	if len(endpts) != eLen {
		t.Errorf("number of retrieved endpoints %v  is not equal to number of saved endpoints %v", len(endpts), eLen)
	}

	for _, ep := range endpts {
		
		epOrganizationsList := ep.GetOrganizationNames()

		if ep.ID == endpoint1.ID {
			eName := []string{"Example Inc."}
			if !helpers.StringArraysEqual(epOrganizationsList, eName) {
				t.Errorf("Expected org name to be %v. Got %v.", eName, epOrganizationsList)
			}
		}
		if ep.ID == endpoint2.ID {
			eName := []string{"Other Example Inc."}
			if !helpers.StringArraysEqual(epOrganizationsList, eName) {
				t.Errorf("Expected org name to be %v. Got %v.", eName, epOrganizationsList)
			}
		}
	}

	// retrieve all distinct endpoints
	distinctEndpts, err := store.GetAllDistinctFHIREndpoints(ctx)
	if err != nil {
		t.Errorf("Error getting fhir endpoints: %s", err1.Error())
	}
	eLen = 2
	if len(distinctEndpts) != eLen {
		t.Errorf("number of retrieved distinct endpoints is not equal to number of saved distinct endpoints")
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
