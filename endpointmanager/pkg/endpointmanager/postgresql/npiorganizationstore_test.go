// +build integration

package postgresql

import (
	"context"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_DeleteAllNPIOrganizations(t *testing.T) {
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
		URL:              "example.com/FHIR/DSTU2/",
		TLSVersion:       "TLS 1.1",
		MIMETypes:        []string{"application/json+fhir"},
		HTTPResponse:     200,
		Errors:           "Example Error",
		OrganizationName: "Example Inc.",
		Vendor:           "Cerner"}
	var endpoint2 = &endpointmanager.FHIREndpoint{
		URL:              "other.example.com/FHIR/DSTU2/",
		TLSVersion:       "TLS 1.2",
		MIMETypes:        []string{"application/fhir+json"},
		HTTPResponse:     404,
		Errors:           "Example Error 2",
		OrganizationName: "Other Example Inc."}

	err = store.LinkNPIOrganizationToFHIREndpoint(ctx, npio1.ID, endpoint1.ID, .85)
	if err == nil {
		t.Fatal("Expected an error linking NPI org and endpoint that are not yet in the DB.")
	}

	err = store.AddNPIOrganization(ctx, npio1)
	if err != nil {
		t.Errorf("Error adding npi organization: %s", err.Error())
	}
	err = store.AddFHIREndpoint(ctx, endpoint1)
	if err != nil {
		t.Errorf("Error adding endpoint: %s", err.Error())
	}
	err = store.LinkNPIOrganizationToFHIREndpoint(ctx, npio1.ID, endpoint1.ID, .85)
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
	err = store.LinkNPIOrganizationToFHIREndpoint(ctx, npio2.ID, endpoint2.ID, .75)
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

	rows, err := store.DB.Query("SELECT organization_id, endpoint_id, confidence FROM endpoint_organization")
	defer rows.Close()

	for rows.Next() {
		var endpointID int
		var npioID int
		var confidence float64

		err = rows.Scan(&npioID, &endpointID, &confidence)

		if endpointID == endpoint1.ID {
			if npioID != npio1.ID {
				t.Fatalf("Expected ID %d to be ID %d", npioID, npio1.ID)
			}
			if confidence != .85 {
				t.Fatalf("Expected confidence %f to be %f", confidence, .85)
			}
		} else if endpointID == endpoint2.ID {
			if npioID != npio2.ID {
				t.Fatalf("Expected ID %d to be ID %d", npioID, npio2.ID)
			}
			if confidence != .75 {
				t.Fatalf("Expected confidence %f to be %f", confidence, .75)
			}
		} else {
			t.Fatal("Getting unexpected entries in DB.")
		}
	}

	// test deletion

	// delete endpoint and ensure corresponding linked entry deleted
	store.DeleteFHIREndpoint(ctx, endpoint1)

	row = store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization")
	err = row.Scan(&count)
	if err != nil {
		t.Fatalf("Got scanning row: %+v.", err)
	}
	if count != 1 {
		t.Fatalf("Expected one row in DB")
	}

	rows, err = store.DB.Query("SELECT organization_id, endpoint_id, confidence FROM endpoint_organization")
	defer rows.Close()

	for rows.Next() {
		var endpointID int
		var npioID int
		var confidence float64

		err = rows.Scan(&npioID, &endpointID, &confidence)

		if endpointID != endpoint2.ID {
			t.Fatal("Getting unexpected entries in DB.")
		}
	}

	// delete npi org and ensure linked entry is deleted
	store.DeleteNPIOrganization(ctx, npio2)

	row = store.DB.QueryRow("SELECT COUNT(*) FROM endpoint_organization")
	err = row.Scan(&count)
	if err != nil {
		t.Fatalf("Got scanning row: %+v.", err)
	}
	if count != 0 {
		t.Fatalf("Expected no rows in DB")
	}
}
