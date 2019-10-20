package main

import (
	"testing"

	_ "github.com/lib/pq"
)

func Test_PersistProviderOrganization(t *testing.T) {
	var err error

	var po1 = &ProviderOrganization{
		Name:             "Hospital #1 of America",
		URL:              "hospital.example.com",
		OrganizationType: "hospital",
		HospitalType:     "Acute Care",
		Ownership:        "Volunary non-profit",
		Beds:             250}
	var po2 = &ProviderOrganization{
		Name:             "Group Practice #1 of America",
		URL:              "grouppractice.example.com",
		OrganizationType: "group practice",
		HospitalType:     "",
		Ownership:        "",
		Beds:             -1}

	db = connectToDB(t)
	defer db.Close()

	// add organizations

	err = po1.Add()
	if err != nil {
		t.Errorf("Error adding provider organization: %s", err.Error())
	}

	err = po2.Add()
	if err != nil {
		t.Errorf("Error adding provider organization: %s", err.Error())
	}

	// retrieve organizations

	p1, err := GetProviderOrganization(po1.GetID())
	if err != nil {
		t.Errorf("Error getting provider organization: %s", err.Error())
	}
	if !p1.Equal(po1) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	p2, err := GetProviderOrganization(po2.GetID())
	if err != nil {
		t.Errorf("Error getting provider organization: %s", err.Error())
	}
	if !p2.Equal(po2) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	// update organization

	p1.HospitalType = "Critical Access"

	err = p1.Update()
	if err != nil {
		t.Errorf("Error updating provider organization: %s", err.Error())
	}

	p1, err = GetProviderOrganization(po1.GetID())
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

	err = po1.Delete()
	if err != nil {
		t.Errorf("Error deleting provider organization: %s", err.Error())
	}

	p2, err = GetProviderOrganization(po2.GetID()) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("po2 no longer exists in DB after deleting po1: %s", err.Error())
	}

	err = po2.Delete()
	if err != nil {
		t.Errorf("Error deleting provider organization: %s", err.Error())
	}
}

func Test_ProviderOrganizationEqual(t *testing.T) {
	var po1 = &ProviderOrganization{
		id:   1,
		Name: "Hospital #1 of America",
		URL:  "hospital.example.com",
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		OrganizationType: "hospital",
		HospitalType:     "Acute Care",
		Ownership:        "Volunary non-profit",
		Beds:             250}

	var po2 = &ProviderOrganization{
		id:   1,
		Name: "Hospital #1 of America",
		URL:  "hospital.example.com",
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		OrganizationType: "hospital",
		HospitalType:     "Acute Care",
		Ownership:        "Volunary non-profit",
		Beds:             250}

	if !po1.Equal(po2) {
		t.Errorf("Expected provider organization 1 to equal provider organization 2. They are not equal.")
	}

	po2.id = 2
	if po1.Equal(po2) {
		t.Errorf("Did not expect provider organization 1 to equal provider organization 2. id should be different. %d vs %d", po2.id, po2.id)
	}
	po2.id = po1.id

	po2.Name = "other"
	if po1.Equal(po2) {
		t.Errorf("Did not expect provider organization 1 to equal provider organization 2. Name should be different. %s vs %s", po2.Name, po2.Name)
	}
	po2.Name = po1.Name

	po2.URL = "other"
	if po1.Equal(po2) {
		t.Errorf("Did not expect provider organization 1 to equal provider organization 2. URL should be different. %s vs %s", po2.URL, po2.URL)
	}
	po2.URL = po1.URL

	po2.Location.Address1 = "other"
	if po1.Equal(po2) {
		t.Errorf("Did not expect provider organization 1 to equal provider organization 2. Location.Address1 should be different. %s vs %s", po2.Location.Address1, po2.Location.Address1)
	}
	po2.Location.Address1 = po1.Location.Address1

	po2.OrganizationType = "other"
	if po1.Equal(po2) {
		t.Errorf("Did not expect provider organization 1 to equal provider organization 2. OrganizationType should be different. %s vs %s", po2.OrganizationType, po2.OrganizationType)
	}
	po2.OrganizationType = po1.OrganizationType

	po2.HospitalType = "other"
	if po1.Equal(po2) {
		t.Errorf("Did not expect provider organization 1 to equal provider organization 2. HospitalType should be different. %s vs %s", po2.HospitalType, po2.HospitalType)
	}
	po2.HospitalType = po1.HospitalType

	po2.Ownership = "other"
	if po1.Equal(po2) {
		t.Errorf("Did not expect provider organization 1 to equal provider organization 2. Ownership should be different. %s vs %s", po2.Ownership, po2.Ownership)
	}
	po2.Ownership = po1.Ownership

	po2.Beds = 0
	if po1.Equal(po2) {
		t.Errorf("Did not expect provider organization 1 to equal provider organization 2. Beds should be different. %d vs %d", po2.Beds, po2.Beds)
	}
	po2.Beds = po1.Beds

	po2 = nil
	if po1.Equal(po2) {
		t.Errorf("Did not expect provider organization 1 to equal nil provider organization 2.")
	}
	po2 = po1

	po1 = nil
	if po1.Equal(po2) {
		t.Errorf("Did not expect nil provider organization 1 to equal provider organization 2.")
	}

	po2 = nil
	if !po1.Equal(po2) {
		t.Errorf("Nil provider organization 1 should equal nil provider organization 2.")
	}
}
