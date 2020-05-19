package endpointmanager

import (
	"time"

	"testing"

	_ "github.com/lib/pq"
)

func Test_VendorEqual(t *testing.T) {
	v1 := &Vendor{
		ID:            1,
		Name:          "Epic Systems Corporation",
		DeveloperCode: "1447",
		CHPLID:        448,
		URL:           "http://www.epic.com",
		Location: &Location{
			Address1: "1979 Milky Way",
			City:     "Verona",
			State:    "WI",
			ZipCode:  "53593"},
		Status:             "active",
		LastModifiedInCHPL: time.Date(2020, time.February, 24, 0, 0, 0, 0, time.UTC),
	}
	v2 := &Vendor{
		ID:            1,
		Name:          "Epic Systems Corporation",
		DeveloperCode: "1447",
		CHPLID:        448,
		URL:           "http://www.epic.com",
		Location: &Location{
			Address1: "1979 Milky Way",
			City:     "Verona",
			State:    "WI",
			ZipCode:  "53593"},
		Status:             "active",
		LastModifiedInCHPL: time.Date(2020, time.February, 24, 0, 0, 0, 0, time.UTC),
	}

	if !v1.Equal(v2) {
		t.Errorf("Expected v1 to equal v2. They are not equal.")
	}

	v2.ID = 2
	if !v1.Equal(v2) {
		t.Errorf("Expect vendor 1 to equal vendor 2. ids should be ignored. %d vs %d", v1.ID, v2.ID)
	}
	v2.ID = v1.ID

	v2.Name = "other"
	if v1.Equal(v2) {
		t.Errorf("Did not expect vendor 1 to equal vendor 2. Name should be different. %s vs %s", v1.Name, v2.Name)
	}
	v2.Name = v1.Name

	v2.DeveloperCode = "other"
	if v1.Equal(v2) {
		t.Errorf("Did not expect vendor 1 to equal vendor 2. DeveloperCode should be different. %s vs %s", v1.DeveloperCode, v2.DeveloperCode)
	}
	v2.DeveloperCode = v1.DeveloperCode

	v2.CHPLID = 3
	if v1.Equal(v2) {
		t.Errorf("Did not expect vendor 1 to equal vendor 2. CHPLID should be different. %d vs %d", v1.CHPLID, v2.CHPLID)
	}
	v2.CHPLID = v1.CHPLID

	v2.Location.Address1 = "other"
	if v1.Equal(v2) {
		t.Errorf("Did not expect vendor 1 to equal vendor 2. Location.Address1 should be different. %s vs %s", v1.Location.Address1, v2.Location.Address1)
	}
	v2.Location.Address1 = v1.Location.Address1

	v2.URL = "other"
	if v1.Equal(v2) {
		t.Errorf("Did not expect vendor 1 to equal vendor 2. URL should be different. %s vs %s", v1.URL, v2.URL)
	}
	v2.URL = v1.URL

	v2.Status = "other"
	if v1.Equal(v2) {
		t.Errorf("Did not expect vendor 1 to equal vendor 2. Status should be different. %s vs %s", v1.Status, v2.Status)
	}
	v2.Status = v1.Status

	v2.LastModifiedInCHPL = v2.LastModifiedInCHPL.Add(500)
	if v1.Equal(v2) {
		t.Errorf("Did not expect vendor 1 to equal vendor 2. LastModifiedInCHPL should be different.")
	}
	v2.LastModifiedInCHPL = v1.LastModifiedInCHPL

	v2 = nil
	if v1.Equal(v2) {
		t.Errorf("Did not expect v1 to equal nil v2.")
	}
	v2 = v1

	v1 = nil
	if v1.Equal(v2) {
		t.Errorf("Did not expect nil v1 to equal v2.")
	}

	v2 = nil
	if !v1.Equal(v2) {
		t.Errorf("Nil v1 should equal nil v2.")
	}
}
