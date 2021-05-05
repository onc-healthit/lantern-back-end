// +build integration

package postgresql

import (
	"context"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_DeleteAllNPIContacts(t *testing.T) {
	SetupStore()
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

	var contact1 = &endpointmanager.NPIContact{
		ID:                           1,
		NPI_ID:                       "1",
		EndpointType:                 "FHIR",
		EndpointTypeDescription:      "TestEndpointTypeDescription",
		Endpoint:                     "http://foobar1.com",
		ValidURL:                     true,
		Affiliation:                  "TestAffiliation",
		EndpointDescription:          "TestEndpointDescription",
		AffiliationLegalBusinessName: "TestAffiliationLegalBusinessName",
		UseCode:                      "TestCode",
		UseDescription:               "TestUseDescription",
		OtherUseDescription:          "TestOtherUseDescription",
		ContentType:                  "TestContentType",
		ContentDescription:           "TestContentDescription",
		OtherContentDescription:      "TestOtherContentDescription",
		Location: &endpointmanager.Location{
			Address1: "TestAddressLine1",
			Address2: "TestAddressLine2",
			City:     "TestCity",
			State:    "TestState",
			ZipCode:  "12345"},
	}

	var contact2 = &endpointmanager.NPIContact{
		ID:                           2,
		NPI_ID:                       "2",
		EndpointType:                 "FHIR",
		EndpointTypeDescription:      "TestEndpointTypeDescription",
		Endpoint:                     "http://foobar2.com",
		ValidURL:                     true,
		Affiliation:                  "TestAffiliation",
		EndpointDescription:          "TestEndpointDescription",
		AffiliationLegalBusinessName: "TestAffiliationLegalBusinessName",
		UseCode:                      "TestCode",
		UseDescription:               "TestUseDescription",
		OtherUseDescription:          "TestOtherUseDescription",
		ContentType:                  "TestContentType",
		ContentDescription:           "TestContentDescription",
		OtherContentDescription:      "TestOtherContentDescription",
		Location: &endpointmanager.Location{
			Address1: "TestAddressLine1",
			Address2: "TestAddressLine2",
			City:     "TestCity",
			State:    "TestState",
			ZipCode:  "12345"},
	}

	// add contacts

	err = store.AddNPIContact(ctx, contact1)
	if err != nil {
		t.Errorf("Error adding npi contact: %s", err.Error())
	}

	err = store.AddNPIContact(ctx, contact2)
	if err != nil {
		t.Errorf("Error adding npi contact: %s", err.Error())
	}

	// retrieve contacts by NPI_ID

	contact1_get, err := store.GetNPIContactByNPIID(ctx, contact1.NPI_ID)
	if err != nil {
		t.Errorf("Error getting npi contact: %s", err.Error())
	}
	if contact1_get.Endpoint != contact1.Endpoint {
		t.Errorf("retrieved contact is not equal to saved contact.")
	}

	contact2_get, err := store.GetNPIContactByNPIID(ctx, contact2.NPI_ID)
	if err != nil {
		t.Errorf("Error getting npi contact: %s", err.Error())
	}
	if contact2_get.Endpoint != contact2.Endpoint {
		t.Errorf("retrieved contact is not equal to saved contact.")
	}

	err = store.DeleteAllNPIContacts(ctx)
	if err != nil {
		t.Errorf("Error deleteing all npi contact: %s", err.Error())
	}

	// retrieve contacts by NPI_ID, they should not exist now

	contact1_get_nil, err := store.GetNPIContactByNPIID(ctx, contact1.NPI_ID)
	if err == nil {
		t.Errorf("Expected error getting non-existant contact.")
	}
	if contact1_get_nil != nil {
		t.Errorf("Retrieved contact that should not exist")
	}

	contact2_get_nil, err := store.GetNPIContactByNPIID(ctx, contact2.NPI_ID)
	if err == nil {
		t.Errorf("Expected error getting non-existant contact.")
	}
	if contact2_get_nil != nil {
		t.Errorf("Retrieved contact that should not exist")
	}
}

func Test_PersistNPIContact(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

	var contact1 = &endpointmanager.NPIContact{
		ID:                           1,
		NPI_ID:                       "1",
		EndpointType:                 "FHIR",
		EndpointTypeDescription:      "TestEndpointTypeDescription",
		Endpoint:                     "http://foobar1.com",
		ValidURL:                     true,
		Affiliation:                  "TestAffiliation",
		EndpointDescription:          "TestEndpointDescription",
		AffiliationLegalBusinessName: "TestAffiliationLegalBusinessName",
		UseCode:                      "TestCode",
		UseDescription:               "TestUseDescription",
		OtherUseDescription:          "TestOtherUseDescription",
		ContentType:                  "TestContentType",
		ContentDescription:           "TestContentDescription",
		OtherContentDescription:      "TestOtherContentDescription",
		Location: &endpointmanager.Location{
			Address1: "TestAddressLine1",
			Address2: "TestAddressLine2",
			City:     "TestCity",
			State:    "TestState",
			ZipCode:  "12345"},
	}

	var contact2 = &endpointmanager.NPIContact{
		ID:                           2,
		NPI_ID:                       "2",
		EndpointType:                 "FHIR",
		EndpointTypeDescription:      "TestEndpointTypeDescription",
		Endpoint:                     "http://foobar2.com",
		ValidURL:                     true,
		Affiliation:                  "TestAffiliation",
		EndpointDescription:          "TestEndpointDescription",
		AffiliationLegalBusinessName: "TestAffiliationLegalBusinessName",
		UseCode:                      "TestCode",
		UseDescription:               "TestUseDescription",
		OtherUseDescription:          "TestOtherUseDescription",
		ContentType:                  "TestContentType",
		ContentDescription:           "TestContentDescription",
		OtherContentDescription:      "TestOtherContentDescription",
		Location: &endpointmanager.Location{
			Address1: "TestAddressLine1",
			Address2: "TestAddressLine2",
			City:     "TestCity",
			State:    "TestState",
			ZipCode:  "12345"},
	}

	// add contacts

	err = store.AddNPIContact(ctx, contact1)
	if err != nil {
		t.Errorf("Error adding npi contact: %s", err.Error())
	}

	err = store.AddNPIContact(ctx, contact2)
	if err != nil {
		t.Errorf("Error adding npi contact: %s", err.Error())
	}

	// retrieve contacts by NPI_ID

	contact1_get, err := store.GetNPIContactByNPIID(ctx, contact1.NPI_ID)
	if err != nil {
		t.Errorf("Error getting npi contact: %s", err.Error())
	}
	if contact1_get.Endpoint != contact1.Endpoint {
		t.Errorf("retrieved contact is not equal to saved contact.")
	}

	contact2_get, err := store.GetNPIContactByNPIID(ctx, contact2.NPI_ID)
	if err != nil {
		t.Errorf("Error getting npi contact: %s", err.Error())
	}
	if contact2_get.Endpoint != contact2.Endpoint {
		t.Errorf("retrieved contact is not equal to saved contact.")
	}

	// update contact using UpdateNPIContactByNPIID

	temp_affiliation := contact1.Affiliation
	contact1.Affiliation = "ChangedAffiliation"

	err = store.UpdateNPIContactByNPIID(ctx, contact1)
	if err != nil {
		t.Errorf("Error updating npi contact: %s", err.Error())
	}

	// Restore affiliation
	contact1.Affiliation = temp_affiliation

	contact1_get, err = store.GetNPIContactByNPIID(ctx, contact1.NPI_ID)
	if err != nil {
		t.Errorf("Error getting npi contact: %s", err.Error())
	}
	if contact1_get.Affiliation == contact1.Affiliation {
		t.Errorf("retrieved UPDATED contact is equal to original contact.")
	}
	if contact1_get.UpdatedAt.Equal(contact1.CreatedAt) {
		t.Errorf("UpdatedAt is not being properly set on update.")
	}

	// delete contacts

	err = store.DeleteNPIContact(ctx, contact1)
	if err != nil {
		t.Errorf("Error deleting npi contact: %s", err.Error())
	}

	_, err = store.GetNPIContactByNPIID(ctx, contact1.NPI_ID) // ensure we deleted the entry
	if err == nil {
		t.Errorf("contact1 was not deleted: %s", err.Error())
	}

	_, err = store.GetNPIContactByNPIID(ctx, contact2.NPI_ID) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("error retrieving contact2 after deleting contact1: %s", err.Error())
	}

	err = store.DeleteNPIContact(ctx, contact2)
	if err != nil {
		t.Errorf("Error deleting npi contact: %s", err.Error())
	}
}
