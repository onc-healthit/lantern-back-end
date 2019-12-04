package testhelper

import "testing"

// Assert checks that the boolean statement is true. If not, it fails the test with the given
// error value.
// Assert streamlines test checks.
func Assert(t *testing.T, boolStatement bool, errorValue interface{}) {
	if !boolStatement {
		t.Fatalf("%s: %v", t.Name(), errorValue)
	}
}
