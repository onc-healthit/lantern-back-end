package endpointmanager

import (
	"testing"

	_ "github.com/lib/pq"
)

func Test_LocationEqual(t *testing.T) {
	var l1 = &Location{
		Address1: "123 Gov Way",
		Address2: "Suite 123",
		Address3: "Mailstop 1",
		City:     "A City",
		State:    "AK",
		ZipCode:  "00000"}

	var l2 = &Location{
		Address1: "123 Gov Way",
		Address2: "Suite 123",
		Address3: "Mailstop 1",
		City:     "A City",
		State:    "AK",
		ZipCode:  "00000"}

	if !l1.Equal(l2) {
		t.Errorf("Expected location 1 to equal location 2. They are not equal.")
	}

	l2.Address1 = "other"
	if l1.Equal(l2) {
		t.Errorf("Did not expect location 1 to equal location 2. Address1 should be different. %s vs %s", l2.Address1, l2.Address1)
	}
	l2.Address1 = l1.Address1

	l2.Address2 = "other"
	if l1.Equal(l2) {
		t.Errorf("Did not expect location 1 to equal location 2. Address2 should be different. %s vs %s", l2.Address2, l2.Address2)
	}
	l2.Address2 = l1.Address2

	l2.Address3 = "other"
	if l1.Equal(l2) {
		t.Errorf("Did not expect location 1 to equal location 2. Address3 should be different. %s vs %s", l2.Address3, l2.Address3)
	}
	l2.Address3 = l1.Address3

	l2.City = "other"
	if l1.Equal(l2) {
		t.Errorf("Did not expect location 1 to equal location 2. City should be different. %s vs %s", l2.City, l2.City)
	}
	l2.City = l1.City

	l2.State = "other"
	if l1.Equal(l2) {
		t.Errorf("Did not expect location 1 to equal location 2. State should be different. %s vs %s", l2.State, l2.State)
	}
	l2.State = l1.State

	l2.ZipCode = "other"
	if l1.Equal(l2) {
		t.Errorf("Did not expect location 1 to equal location 2. ZipCode should be different. %s vs %s", l2.ZipCode, l2.ZipCode)
	}
	l2.ZipCode = l1.ZipCode

	l2 = nil
	if l1.Equal(l2) {
		t.Errorf("Did not expect location 1 to equal nil location 2.")
	}
	l2 = l1

	l1 = nil
	if l1.Equal(l2) {
		t.Errorf("Did not expect nil location 1 to equal location 2.")
	}

	l2 = nil
	if !l1.Equal(l2) {
		t.Errorf("Nil location 1 should equal nil location 2.")
	}
}
