// +build integration

package postgresql

import (
	"context"
	"fmt"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_DeleteAllNPIOrganizations(t *testing.T) {
	SetupStore()
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

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
		Taxonomy:                "208D00000X",
		NormalizedName:          "HOSPITAL  OF AMERICA",
		NormalizedSecondaryName: "HOSPITAL  OF AMERICA SECOND NAME"}

	var npio2 = &endpointmanager.NPIOrganization{
		ID:            2,
		NPI_ID:        "2",
		Name:          "Hospital #2 of America",
		SecondaryName: "",
		Location: &endpointmanager.Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		Taxonomy:                "208D00000X",
		NormalizedName:          "HOSPITAL  OF AMERICA",
		NormalizedSecondaryName: ""}

	// add organizations

	err = store.AddNPIOrganization(ctx, npio1)
	if err != nil {
		t.Errorf("Error adding npi organization: %s", err.Error())
	}

	err = store.AddNPIOrganization(ctx, npio2)
	if err != nil {
		t.Errorf("Error adding npi organization: %s", err.Error())
	}

	// retrieve organizations by NPI_ID

	npio1_get_npi, err := store.GetNPIOrganizationByNPIID(ctx, npio1.NPI_ID)
	if err != nil {
		t.Errorf("Error getting npi organization: %s", err.Error())
	}
	if !npio1_get_npi.Equal(npio1) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	npio2_get_npi, err := store.GetNPIOrganizationByNPIID(ctx, npio2.NPI_ID)
	if err != nil {
		t.Errorf("Error getting npi organization: %s", err.Error())
	}
	if !npio2_get_npi.Equal(npio2) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	err = store.DeleteAllNPIOrganizations(ctx)
	if err != nil {
		t.Errorf("Error deleteing all npi organization: %s", err.Error())
	}

	// retrieve organizations by NPI_ID, they should not exist now

	npio1_get_nil, err := store.GetNPIOrganizationByNPIID(ctx, npio1.NPI_ID)
	if err == nil {
		t.Errorf("Expected error getting non-existant organization.")
	}
	if npio1_get_nil != nil {
		t.Errorf("Retrieved organization that should not exist")
	}

	npio2_get_nil, err := store.GetNPIOrganizationByNPIID(ctx, npio2.NPI_ID)
	if err == nil {
		t.Errorf("Expected error getting non-existant organization.")
	}
	if npio2_get_nil != nil {
		t.Errorf("Retrieved organization that should not exist")
	}
}

func Test_PersistNPIOrganization(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

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
		Taxonomy:                "208D00000X",
		NormalizedName:          "HOSPITAL  OF AMERICA",
		NormalizedSecondaryName: "HOSPITAL  OF AMERICA SECOND NAME"}

	var npio2 = &endpointmanager.NPIOrganization{
		ID:                      2,
		NPI_ID:                  "2",
		Name:                    "A Primary Name",
		SecondaryName:           "A Secondary Name",
		Location:                &endpointmanager.Location{},
		Taxonomy:                "208D00000X",
		NormalizedName:          "A PRIMARY NAME",
		NormalizedSecondaryName: "A SECONDARY NAME"}

	// add organizations

	err = store.AddNPIOrganization(ctx, npio1)
	if err != nil {
		t.Errorf("Error adding npi organization: %s", err.Error())
	}

	err = store.AddNPIOrganization(ctx, npio2)
	if err != nil {
		t.Errorf("Error adding npi organization: %s", err.Error())
	}

	// retrieve organizations

	npio1_get, err := store.GetNPIOrganization(ctx, npio1.ID)
	if err != nil {
		t.Errorf("Error getting npi organization: %s", err.Error())
	}
	if !npio1_get.Equal(npio1) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	npio2_get, err := store.GetNPIOrganization(ctx, npio2.ID)
	if err != nil {
		t.Errorf("Error getting npi organization: %s", err.Error())
	}
	if !npio2_get.Equal(npio2) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	// retrieve organizations by NPI_ID

	npio1_get_npi, err := store.GetNPIOrganizationByNPIID(ctx, npio1.NPI_ID)
	if err != nil {
		t.Errorf("Error getting npi organization: %s", err.Error())
	}
	if !npio1_get_npi.Equal(npio1) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	npio2_get_npi, err := store.GetNPIOrganizationByNPIID(ctx, npio2.NPI_ID)
	if err != nil {
		t.Errorf("Error getting npi organization: %s", err.Error())
	}
	if !npio2_get_npi.Equal(npio2) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	// retrieve NPI organization normalized names

	npio_get_names, err := store.GetAllNPIOrganizationNormalizedNames(ctx)
	if err != nil {
		t.Errorf("Error getting npi organization normalized names: %s", err.Error())
	}
	eLength := 2
	if len(npio_get_names) != eLength {
		t.Errorf("Expected npi org list to have length %d. Got %d.", eLength, len(npio_get_names))
	}

	for _, org := range npio_get_names {
		if org.ID == npio1.ID {
			ePrim := "HOSPITAL  OF AMERICA"
			eSec := "HOSPITAL  OF AMERICA SECOND NAME"
			if org.NormalizedName != ePrim {
				t.Errorf("Expected normalized primary name to be %s. Got %s.", ePrim, org.NormalizedName)
			}
			if org.NormalizedSecondaryName != eSec {
				t.Errorf("Expected normalized secondary name to be %s. Got %s.", eSec, org.NormalizedSecondaryName)
			}
		}
		if org.ID == npio2.ID {
			ePrim := "A PRIMARY NAME"
			eSec := "A SECONDARY NAME"
			if org.NormalizedName != ePrim {
				t.Errorf("Expected normalized primary name to be %s. Got %s.", ePrim, org.NormalizedName)
			}
			if org.NormalizedSecondaryName != eSec {
				t.Errorf("Expected normalized secondary name to be %s. Got %s.", eSec, org.NormalizedSecondaryName)
			}
		}
	}

	// update organization

	temp_taxonomy := npio1.Taxonomy
	npio1.Taxonomy = "1234567"

	err = store.UpdateNPIOrganization(ctx, npio1)
	if err != nil {
		t.Errorf("Error updating npi organization: %s", err.Error())
	}

	// Restore taxonomy
	npio1.Taxonomy = temp_taxonomy

	npio1_get, err = store.GetNPIOrganization(ctx, npio1.ID)
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

	err = store.UpdateNPIOrganizationByNPIID(ctx, npio1)
	if err != nil {
		t.Errorf("Error updating npi organization: %s", err.Error())
	}

	// Restore taxonomy
	npio1.Taxonomy = temp_taxonomy

	npio1_get, err = store.GetNPIOrganization(ctx, npio1.ID)
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

	err = store.DeleteNPIOrganization(ctx, npio1)
	if err != nil {
		t.Errorf("Error deleting npi organization: %s", err.Error())
	}

	_, err = store.GetNPIOrganization(ctx, npio1.ID) // ensure we deleted the entry
	if err == nil {
		t.Errorf("npio1 was not deleted: %s", err.Error())
	}

	_, err = store.GetNPIOrganization(ctx, npio2.ID) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("error retrieving npio2 after deleting npio1: %s", err.Error())
	}

	err = store.DeleteNPIOrganization(ctx, npio2)
	if err != nil {
		t.Errorf("Error deleting npi organization: %s", err.Error())
	}
}

func Test_LinkNPIOrganizationToFHIREndpoint(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

	// orgs
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
		Taxonomy:                "208D00000X",
		NormalizedName:          "HOSPITAL  OF AMERICA",
		NormalizedSecondaryName: "HOSPITAL  OF AMERICA SECOND NAME"}
	var npio2 = &endpointmanager.NPIOrganization{
		ID:                      2,
		NPI_ID:                  "2",
		Name:                    "A Primary Name",
		SecondaryName:           "A Secondary Name",
		Location:                &endpointmanager.Location{},
		Taxonomy:                "208D00000X",
		NormalizedName:          "A PRIMARY NAME",
		NormalizedSecondaryName: "A SECONDARY NAME"}

	// endpoints
	var endpoint1 = &endpointmanager.FHIREndpoint{
		URL:               "example.com/FHIR/DSTU2/",
		OrganizationNames: []string{"Example Inc."},
		NPIIDs:            []string{"1"},
		ListSource:        "https://github.com/cerner/ignite-endpoints"}
	var endpoint2 = &endpointmanager.FHIREndpoint{
		URL:               "other.example.com/FHIR/DSTU2/",
		OrganizationNames: []string{"Other Example Inc."}}

	err = store.AddNPIOrganization(ctx, npio1)
	if err != nil {
		t.Errorf("Error adding npi organization: %s", err.Error())
	}
	err = store.AddFHIREndpoint(ctx, endpoint1)
	if err != nil {
		t.Errorf("Error adding endpoint: %s", err.Error())
	}
	err = store.LinkNPIOrganizationToFHIREndpoint(ctx, npio1.NPI_ID, endpoint1.URL, .85)
	if err != nil {
		t.Fatalf("Got error linking NPI org and endpoint: %+v.", err)
	}

	err = store.AddNPIOrganization(ctx, npio2)
	if err != nil {
		t.Errorf("Error adding npi organization: %s", err.Error())
	}
	err = store.AddFHIREndpoint(ctx, endpoint2)
	if err != nil {
		t.Errorf("Error adding endpoint: %s", err.Error())
	}
	err = store.LinkNPIOrganizationToFHIREndpoint(ctx, npio2.NPI_ID, endpoint2.URL, .75)
	if err != nil {
		t.Fatalf("Got error linking NPI org and endpoint: %+v.", err)
	}

	var count int
	row := store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization")
	err = row.Scan(&count)
	if err != nil {
		t.Fatalf("Got scanning row: %+v.", err)
	}
	if count != 2 {
		t.Fatalf("Expected two rows in DB")
	}

	sNpiID, sEpURL, sConfidence, err := store.GetNPIOrganizationFHIREndpointLink(ctx, npio1.NPI_ID, endpoint1.URL)
	th.Assert(t, err == nil, err)
	th.Assert(t, sNpiID == npio1.NPI_ID, fmt.Sprintf("expected stored ID '%s' to be the same as the ID that was stored '%s'.", sNpiID, npio1.NPI_ID))
	th.Assert(t, sEpURL == endpoint1.URL, fmt.Sprintf("expected stored url '%s' to be the same as the url that was stored '%s'.", sEpURL, endpoint1.URL))
	th.Assert(t, sConfidence == .85, fmt.Sprintf("expected stored confidence '%f' to be the same as the confidence that was stored '%f'.", sConfidence, .85))

	sNpiID, sEpURL, sConfidence, err = store.GetNPIOrganizationFHIREndpointLink(ctx, npio2.NPI_ID, endpoint2.URL)
	th.Assert(t, err == nil, err)
	th.Assert(t, sNpiID == npio2.NPI_ID, fmt.Sprintf("expected stored ID '%s' to be the same as the ID that was stored '%s'.", sNpiID, npio2.NPI_ID))
	th.Assert(t, sEpURL == endpoint2.URL, fmt.Sprintf("expected stored url '%s' to be the same as the url that was stored '%s'.", sEpURL, endpoint2.URL))
	th.Assert(t, sConfidence == .75, fmt.Sprintf("expected stored confidence '%f' to be the same as the confidence that was stored '%f'.", sConfidence, .75))

	err = store.UpdateNPIOrganizationFHIREndpointLink(ctx, npio1.NPI_ID, endpoint1.URL, .5)
	th.Assert(t, err == nil, err)
	sNpiID, sEpURL, sConfidence, err = store.GetNPIOrganizationFHIREndpointLink(ctx, npio1.NPI_ID, endpoint1.URL)
	th.Assert(t, err == nil, err)
	th.Assert(t, sNpiID == npio1.NPI_ID, fmt.Sprintf("expected stored ID '%s' to be the same as the ID that was stored '%s'.", sNpiID, npio1.NPI_ID))
	th.Assert(t, sEpURL == endpoint1.URL, fmt.Sprintf("expected stored url '%s' to be the same as the url that was stored '%s'.", sEpURL, endpoint1.URL))
	th.Assert(t, sConfidence == .5, fmt.Sprintf("expected stored confidence '%f' to be the same as the confidence that was stored '%f'.", sConfidence, .5))
}
