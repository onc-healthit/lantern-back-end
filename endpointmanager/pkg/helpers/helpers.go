package helpers

import "sort"

// StringArrayContains checks if the string array contains the provided string.
func StringArrayContains(l []string, s string) bool {
	for _, s2 := range l {
		if s == s2 {
			return true
		}
	}
	return false
}

// IntArrayContains checks if the integer array contains the provided integer.
func IntArrayContains(l []int, i int) bool {
	for _, i2 := range l {
		if i == i2 {
			return true
		}
	}
	return false
}

// StringArraysEqual checks if l1 and l2 have the same contents regardless of order.
func StringArraysEqual(l1 []string, l2 []string) bool {
	if len(l1) != len(l2) {
		return false
	}

	// don't care about order
	a := make([]string, len(l1))
	b := make([]string, len(l2))
	copy(a, l1)
	copy(b, l2)
	sort.Strings(a)
	sort.Strings(b)
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
