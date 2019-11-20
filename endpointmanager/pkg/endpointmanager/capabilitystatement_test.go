package endpointmanager

import (
	"testing"

	_ "github.com/lib/pq"
)

func Test_CapabilityStatementEqual(t *testing.T) {
	var cs1 = &CapabilityStatement{}

	var cs2 = &CapabilityStatement{}

	if !cs1.Equal(cs2) {
		t.Errorf("Expected capability statment 1 to equal capability statment 2. They are not equal.")
	}

	cs2 = nil
	if cs1.Equal(cs2) {
		t.Errorf("Did not expect capability statment 1 to equal nil capability statment 2.")
	}
	cs2 = cs1

	cs1 = nil
	if cs1.Equal(cs2) {
		t.Errorf("Did not expect nil capability statment 1 to equal capability statment 2.")
	}

	cs2 = nil
	if !cs1.Equal(cs2) {
		t.Errorf("Nil capability statment 1 should equal nil capability statment 2.")
	}
}
