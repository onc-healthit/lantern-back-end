package endpointmanager

import (
	"testing"

	_ "github.com/lib/pq"
)

func Test_ProviderOrganizationEqual(t *testing.T) {
	var po1 = &ProviderOrganization{
		ID:   1,
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
		ID:   1,
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

	po2.ID = 2
	if !po1.Equal(po2) {
		t.Errorf("Expect provider organization 1 to equal provider organization 2. ids should be ignored. id should be different. %d vs %d", po2.ID, po2.ID)
	}
	po2.ID = po1.ID

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
