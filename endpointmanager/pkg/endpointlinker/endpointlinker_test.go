package endpointlinker

import (
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"strconv"
	"testing"
)

func Test_NormalizeOrgName(t *testing.T) {
	orgName := "AMBULANCE & and-chair. SERVICE!"
	expected := "AMBULANCE  AND CHAIR SERVICE"
	normalized := NormalizeOrgName(orgName)
	th.Assert(t, (normalized == expected), "Organization name normalization failed. Expected: "+expected+" Got: "+normalized)
}

func Test_calculateJaccardIndex(t *testing.T) {
	jaccardIndex := calculateJaccardIndex("FOO BAR", "FOO BAR")
	ind := strconv.FormatFloat(jaccardIndex, 'f', -1, 64)
	th.Assert(t, (jaccardIndex == 1), "Jaccard index expected to be 1, was "+ind)

	jaccardIndex = calculateJaccardIndex("FOO BAZ BAR", "FOO BAR")
	ind = strconv.FormatFloat(jaccardIndex, 'f', -1, 64)
	th.Assert(t, (jaccardIndex == .6666666666666666), "Jaccard index expected to be .6666666666666666, was "+ind)

	jaccardIndex = calculateJaccardIndex("FOO FOO BAR", "FOO BAR")
	ind = strconv.FormatFloat(jaccardIndex, 'f', -1, 64)
	th.Assert(t, (jaccardIndex == .6666666666666666), "Jaccard index expected to be .6666666666666666, was "+ind)
}

func Test_IntersectionCount(t *testing.T) {
	emptyListIntersections := intersectionCount([]string{}, []string{})
	th.Assert(t, (emptyListIntersections == 0), "Intersection count of empty lists should be zero, got "+strconv.Itoa(emptyListIntersections))

	emptyListIntersections = intersectionCount([]string{"foo"}, []string{})
	th.Assert(t, (emptyListIntersections == 0), "Intersection count of empty lists should be zero, got "+strconv.Itoa(emptyListIntersections))

	emptyListIntersections = intersectionCount([]string{"foo"}, []string{"bar"})
	th.Assert(t, (emptyListIntersections == 0), "Intersection count of empty lists should be zero, got "+strconv.Itoa(emptyListIntersections))

	nonEmptyListIntersections := intersectionCount([]string{"foo"}, []string{"foo"})
	th.Assert(t, (nonEmptyListIntersections == 1), "Intersection count of empty lists should be one, got "+strconv.Itoa(nonEmptyListIntersections))

	nonEmptyListIntersections = intersectionCount([]string{"foo", "bar"}, []string{"bar"})
	th.Assert(t, (nonEmptyListIntersections == 1), "Intersection count of empty lists should be one, got "+strconv.Itoa(nonEmptyListIntersections))

	nonEmptyListIntersections = intersectionCount([]string{"foo", "bar"}, []string{"bar", "foo"})
	th.Assert(t, (nonEmptyListIntersections == 2), "Intersection count of empty lists should be two, got "+strconv.Itoa(nonEmptyListIntersections))

	nonEmptyListIntersections = intersectionCount([]string{"foo", "bar", "foo", "foo"}, []string{"bar", "foo", "foo"})
	th.Assert(t, (nonEmptyListIntersections == 3), "Intersection count of empty lists should be three, got "+strconv.Itoa(nonEmptyListIntersections))
}

func Test_getIdsOfMatchingNPIOrgs(t *testing.T) {
	var npio1 = endpointmanager.NPIOrganization{
		ID:            1,
		NPI_ID:        "1",
		Name:          "Foo Bar",
		SecondaryName: "",
		NormalizedName: "FOO FOO BAR",
		NormalizedSecondaryName: "",
		Location: &endpointmanager.Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		Taxonomy: "208D00000X"}
	var npio2 = endpointmanager.NPIOrganization{
		ID:            2,
		NPI_ID:        "2",
		Name:          "nothing should match this",
		SecondaryName: "foo bar baz",
		NormalizedName: "NOTHING SHOULD MATCH THIS",
		NormalizedSecondaryName: "FOO FOO BAR BAZ",
		Location: &endpointmanager.Location{
			Address1: "somerandomstring",
			Address2: "Foo Bar",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		Taxonomy: "208D00000X"}
	var npio3 = endpointmanager.NPIOrganization{
		ID:            3,
		NPI_ID:        "3",
		Name:          "nothingshouldmatchthis",
		SecondaryName: "nothingshouldmatchthis",
		NormalizedName: "NOTHINGSHOULDMATCHTHIS",
		NormalizedSecondaryName: "NOTHINGSHOULDMATCHTHIS",
		Location: &endpointmanager.Location{
			Address1: "somerandomstring",
			Address2: "FooBar",
			City:     "A",
			State:    "NH",
			ZipCode:  "00000"},
		Taxonomy: "208D00000X"}

	var  orgs []endpointmanager.NPIOrganization;

	matches, confidences, err :=  getIdsOfMatchingNPIOrgs(orgs, "FOO BAR", false)
	th.Assert(t, (err == nil), "Error getting matches from empty list")
	th.Assert(t, (len(matches) == 0), "There should not have been any matches returned got: " + strconv.Itoa(len(matches)))
	th.Assert(t, (len(confidences) == 0), "There should not have been any confidences returned" + strconv.Itoa(len(matches)))

	orgs = append(orgs, npio1)
	matches, confidences, err =  getIdsOfMatchingNPIOrgs(orgs, "FOO FOO BAR", false)
	th.Assert(t, (err == nil), "Error getting matches from list")
	th.Assert(t, (len(matches) == 1), "There should have been 1 match returned got: " + strconv.Itoa(len(matches)))
	th.Assert(t, (len(confidences) == 1), "There should have been 1 confidence returned" + strconv.Itoa(len(confidences)))

	orgs = append(orgs, npio2)
	orgs = append(orgs, npio3)
	matches, confidences, err =  getIdsOfMatchingNPIOrgs(orgs, "FOO FOO BAR", true)
	th.Assert(t, (err == nil), "Error getting matches from list")
	th.Assert(t, (len(matches) == 2), "There should have been 2 matchs returned got: " + strconv.Itoa(len(matches)))
	th.Assert(t, (len(confidences) == 2), "There should have been 2 confidences returned" + strconv.Itoa(len(confidences)))

}