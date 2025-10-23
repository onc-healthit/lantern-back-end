package helpers

import (
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_StringArrayContains(t *testing.T) {
	var sa []string
	var s string
	var contains bool

	// check with uninitialized objects
	contains = StringArrayContains(sa, s)
	th.Assert(t, !contains, "did not expect uninitialized array to contain empty string")

	// check with string in uninitialized array
	s = "string"
	contains = StringArrayContains(sa, s)
	th.Assert(t, !contains, "did not expect uninitialized array to contain string")

	// check with initialized empty array
	sa = make([]string, 5)
	contains = StringArrayContains(sa, s)
	th.Assert(t, !contains, "did not expect empty array to contain string")

	// check with multi-element array that doesn't include string
	sa = []string{"this", "is", "an", "array"}
	contains = StringArrayContains(sa, s)
	th.Assert(t, !contains, "did not expect array not containing string to return that it contains the string")

	// check with multi-element array that does include string
	sa = []string{"this", "is", "a", "string", "array"}
	contains = StringArrayContains(sa, s)
	th.Assert(t, contains, "expected array containing string to return that it contains the string")
}

func Test_IntArrayContains(t *testing.T) {
	var ia []int
	var i int
	var contains bool

	// check with uninitialized objects
	contains = IntArrayContains(ia, i)
	th.Assert(t, !contains, "did not expect uninitialized array to contain 0")

	// check with int in uninitialized array
	i = 1
	contains = IntArrayContains(ia, i)
	th.Assert(t, !contains, "did not expect uninitialized array to contain int")

	// check with initialized empty array
	ia = make([]int, 5)
	contains = IntArrayContains(ia, i)
	th.Assert(t, !contains, "did not expect empty array to contain int")

	// check with multi-element array that doesn't include int
	ia = []int{4, 5, 6, 7}
	contains = IntArrayContains(ia, i)
	th.Assert(t, !contains, "did not expect array not containing int to return that it contains the int")

	// check with multi-element array that does include int
	ia = []int{1, 2, 3, 4, 5}
	contains = IntArrayContains(ia, i)
	th.Assert(t, contains, "expected array containing int to return that it contains the int")
}

func Test_StringArraysEqual(t *testing.T) {
	var sa1 []string
	var sa2 []string
	var areEqual bool

	// test uninitialized
	areEqual = StringArraysEqual(sa1, sa2)
	th.Assert(t, areEqual, "expected two uninitialized string arrays to come back as equal")

	// test initialized different length
	sa1 = make([]string, 3)
	sa2 = make([]string, 6)
	areEqual = StringArraysEqual(sa1, sa2)
	th.Assert(t, !areEqual, "did not expect two initialized string arrays of different length to come back as equal")

	// test initialized same length
	sa1 = make([]string, 3)
	sa2 = make([]string, 3)
	areEqual = StringArraysEqual(sa1, sa2)
	th.Assert(t, areEqual, "expected two initialized string arrays of same length to come back as equal")

	// test initialized slices different capacity
	sa1 = make([]string, 3, 6)
	sa2 = make([]string, 3, 9)
	areEqual = StringArraysEqual(sa1, sa2)
	th.Assert(t, areEqual, "expected two initialized string arrays of same length but different capacity to come back as equal")

	// test sa1 empty, sa2 has content
	sa2 = []string{"a", "thing"}
	areEqual = StringArraysEqual(sa1, sa2)
	th.Assert(t, !areEqual, "did not expect empty array to be equal to filled array")

	// test sa1 has content, sa2 empty
	sa1 = []string{"a", "thing"}
	sa2 = make([]string, 6)
	areEqual = StringArraysEqual(sa1, sa2)
	th.Assert(t, !areEqual, "did not expect empty array to be equal to filled array")

	// test different lengths
	sa2 = []string{"a", "thing", "and", "another"}
	areEqual = StringArraysEqual(sa1, sa2)
	th.Assert(t, !areEqual, "did not arrays of different lengths to be equal")

	// test same length different content
	sa2 = []string{"another", "thing"}
	areEqual = StringArraysEqual(sa1, sa2)
	th.Assert(t, !areEqual, "did not arrays with different content to be equal")

	// test same content same order
	sa2 = []string{"a", "thing"}
	areEqual = StringArraysEqual(sa1, sa2)
	th.Assert(t, areEqual, "expected arrays with same content to be equal")

	// test same content different order
	sa2 = []string{"thing", "a"}
	areEqual = StringArraysEqual(sa1, sa2)
	th.Assert(t, areEqual, "expected arrays with same content and different order to be equal")

}
