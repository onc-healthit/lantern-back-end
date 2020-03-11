package endpointlinker

import (
	"testing"
	"strconv"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"

)

func Test_NormalizeOrgName(t *testing.T) {
	orgName := "AMBULANCE & and-chair. SERVICE!"
	expected := "AMBULANCE  AND CHAIR SERVICE"
	normalized := NormalizeOrgName(orgName)
	th.Assert(t, (normalized == expected), "Organization name normalization failed. Expected: "+ expected + " Got: " + normalized)
}

func Test_calculateJaccardIndex(t *testing.T) {
	jacccardIndex := calculateJaccardIndex("FOO BAR", "FOO BAR")
	ind := strconv.FormatFloat(jacccardIndex, 'f', -1, 64)
	th.Assert(t, (jacccardIndex == 1), "Jacard index expected to be 1, was " + ind)

	jacccardIndex = calculateJaccardIndex("FOO BAZ BAR", "FOO BAR")
	ind = strconv.FormatFloat(jacccardIndex, 'f', -1, 64)
	th.Assert(t, (jacccardIndex == .6666666666666666), "Jacard index expected to be .6666666666666666, was " + ind)

	jacccardIndex = calculateJaccardIndex("FOO FOO BAR", "FOO BAR")
	ind = strconv.FormatFloat(jacccardIndex, 'f', -1, 64)
	th.Assert(t, (jacccardIndex == .6666666666666666), "Jacard index expected to be .6666666666666666, was " + ind)
}

func Test_IntersectionCount(t *testing.T) {
	emptyListIntersections := intersectionCount([]string{},[]string{})
	th.Assert(t, (emptyListIntersections == 0), "Intersection count of empty lists should be zero, got " + strconv.Itoa(emptyListIntersections))

	emptyListIntersections = intersectionCount([]string{"foo"},[]string{})
	th.Assert(t, (emptyListIntersections == 0), "Intersection count of empty lists should be zero, got " + strconv.Itoa(emptyListIntersections))

	emptyListIntersections = intersectionCount([]string{"foo"},[]string{"bar"})
	th.Assert(t, (emptyListIntersections == 0), "Intersection count of empty lists should be zero, got " + strconv.Itoa(emptyListIntersections))

	nonEmptyListIntersections := intersectionCount([]string{"foo"},[]string{"foo"})
	th.Assert(t, (nonEmptyListIntersections == 1), "Intersection count of empty lists should be one, got " + strconv.Itoa(nonEmptyListIntersections))

	nonEmptyListIntersections = intersectionCount([]string{"foo","bar"},[]string{"bar"})
	th.Assert(t, (nonEmptyListIntersections == 1), "Intersection count of empty lists should be one, got " + strconv.Itoa(nonEmptyListIntersections))

	nonEmptyListIntersections = intersectionCount([]string{"foo","bar"},[]string{"bar", "foo"})
	th.Assert(t, (nonEmptyListIntersections == 2), "Intersection count of empty lists should be two, got " + strconv.Itoa(nonEmptyListIntersections))

	nonEmptyListIntersections = intersectionCount([]string{"foo","bar","foo","foo"},[]string{"bar", "foo", "foo"})
	th.Assert(t, (nonEmptyListIntersections == 3), "Intersection count of empty lists should be three, got " + strconv.Itoa(nonEmptyListIntersections))

}