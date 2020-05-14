package endpointmanager

import (
	"testing"

	_ "github.com/lib/pq"
)

func Test_NPIOrganizationEqual(t *testing.T) {
	var npio1 = &NPIOrganization{
		ID:     1,
		NPI_ID: "1",
		Names:  []string{"Hospital #1 of America", "Hospital #1 of America Second Name"},
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		Taxonomy: "208D00000X"}

	var npio2 = &NPIOrganization{
		ID:     1,
		NPI_ID: "1",
		Names:  []string{"Hospital #1 of America", "Hospital #1 of America Second Name"},
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		Taxonomy: "208D00000X"}

	if !npio1.Equal(npio2) {
		t.Errorf("Expected npi organization 1 to equal npi organization 2. They are not equal.")
	}

	npio2.ID = 2
	if !npio1.Equal(npio2) {
		t.Errorf("Expect npi organization 1 to equal npi organization 2. ids should be ignored. id should be different. %d vs %d", npio2.ID, npio2.ID)
	}
	npio2.ID = npio1.ID

	npio2.Names = []string{"other"}
	if npio1.Equal(npio2) {
		t.Errorf("Did not expect npi organization 1 to equal npi organization 2. Name should be different. %v vs %v", npio2.Names, npio2.Names)
	}
	npio2.Names = npio1.Names

	npio2.Location.Address1 = "other"
	if npio1.Equal(npio2) {
		t.Errorf("Did not expect npi organization 1 to equal npi organization 2. Location.Address1 should be different. %s vs %s", npio2.Location.Address1, npio2.Location.Address1)
	}
	npio2.Location.Address1 = npio1.Location.Address1

	npio2.Taxonomy = "other"
	if npio1.Equal(npio2) {
		t.Errorf("Did not expect npi organization 1 to equal npi organization 2. Taxonomy should be different. %s vs %s", npio2.Taxonomy, npio2.Taxonomy)
	}
	npio2.Taxonomy = npio1.Taxonomy

	npio2 = nil
	if npio1.Equal(npio2) {
		t.Errorf("Did not expect npi organization 1 to equal nil npi organization 2.")
	}
	npio2 = npio1

	npio1 = nil
	if npio1.Equal(npio2) {
		t.Errorf("Did not expect nil npi organization 1 to equal npi organization 2.")
	}

	npio2 = nil
	if !npio1.Equal(npio2) {
		t.Errorf("Nil npi organization 1 should equal nil npi organization 2.")
	}
}
