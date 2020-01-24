package endpointlinker

import (
	"testing"
	"strconv"
)

func Test_NormalizeOrgName(t *testing.T) {
	orgName := "AMBULANCE & chair. SERVICE!"
	expected := "AMBULANCE  CHAIR SERVICE"
	normalized := normalizeOrgName(orgName)
	if normalized != expected {
		t.Errorf("Organization name normalization failed. Expected: "+ expected + " Got: " + normalized)
	}
}

func Test_CalculateJaccardIndex(t *testing.T) {
	jacardIndex := calculateJaccardIndex("FOO BAR", "FOO BAR")
	if jacardIndex != 1 {
		ind := strconv.FormatFloat(jacardIndex, 'f', -1, 64)
		t.Errorf("Jacard index expected to be 1, was " + ind)
	}

	jacardIndex = calculateJaccardIndex("FOO BAZ BAR", "FOO BAR")
	if jacardIndex != .6666666666666666 {
		ind := strconv.FormatFloat(jacardIndex, 'f', -1, 64)
		t.Errorf("Jacard index expected to be 1, was " + ind)
	}
}

func Test_IntersectionCount(t *testing.T) {
	emptyListIntersections := intersectionCount([]string{},[]string{})
	if emptyListIntersections != 0 {
		t.Errorf("Intersection count of empty lists should be zero, got " + strconv.Itoa(emptyListIntersections))
	}

	emptyListIntersections = intersectionCount([]string{"foo"},[]string{})
	if emptyListIntersections != 0 {
		t.Errorf("Intersection count of empty lists should be zero, got " + strconv.Itoa(emptyListIntersections))
	}

	nonEmptyListIntersections := intersectionCount([]string{"foo"},[]string{"bar"})
	if nonEmptyListIntersections != 0 {
		t.Errorf("Intersection count of empty lists should be zero, got " + strconv.Itoa(nonEmptyListIntersections))
	}

	nonEmptyListIntersections = intersectionCount([]string{"foo"},[]string{"foo"})
	if nonEmptyListIntersections != 1 {
		t.Errorf("Intersection count of empty lists should be zero, got " + strconv.Itoa(nonEmptyListIntersections))
	}

	nonEmptyListIntersections = intersectionCount([]string{"foo","bar"},[]string{"bar"})
	if nonEmptyListIntersections != 1 {
		t.Errorf("Intersection count of empty lists should be zero, got " + strconv.Itoa(nonEmptyListIntersections))
	}

	nonEmptyListIntersections = intersectionCount([]string{"foo","bar"},[]string{"bar", "foo"})
	if nonEmptyListIntersections != 2 {
		t.Errorf("Intersection count of empty lists should be zero, got " + strconv.Itoa(nonEmptyListIntersections))
	}

	nonEmptyListIntersections = intersectionCount([]string{"foo","bar","foo","foo"},[]string{"bar", "foo", "foo"})
	if nonEmptyListIntersections != 3 {
		t.Errorf("Intersection count of empty lists should be zero, got " + strconv.Itoa(nonEmptyListIntersections))
	}
}