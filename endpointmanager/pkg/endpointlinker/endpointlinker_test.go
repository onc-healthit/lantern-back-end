package endpointlinker

import (
	"fmt"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
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
	var exactPrimaryNameOrg = endpointmanager.NPIOrganization{
		ID:                      1,
		NPI_ID:                  "1",
		Name:                    "Foo Bar",
		SecondaryName:           "",
		NormalizedName:          "FOO FOO BAR",
		NormalizedSecondaryName: "FOO FOO BAR BAZ",
		Location: &endpointmanager.Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		Taxonomy: "208D00000X"}
	var nonExactSecondaryNameOrg = endpointmanager.NPIOrganization{
		ID:                      2,
		NPI_ID:                  "2",
		Name:                    "Foo Bar",
		SecondaryName:           "foo bar baz",
		NormalizedName:          "NOTHING SHOULD MATCH THIS",
		NormalizedSecondaryName: "FOO FOO BAR BAZ",
		Location: &endpointmanager.Location{
			Address1: "somerandomstring",
			Address2: "Foo Bar",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		Taxonomy: "208D00000X"}
	var exactSecondaryNameOrg = endpointmanager.NPIOrganization{
		ID:                      4,
		NPI_ID:                  "4",
		Name:                    "Foo Bar",
		SecondaryName:           "foo bar baz",
		NormalizedName:          "FOO FOO BAR BAZ",
		NormalizedSecondaryName: "FOO FOO BAR",
		Location: &endpointmanager.Location{
			Address1: "somerandomstring",
			Address2: "Foo Bar",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		Taxonomy: "208D00000X"}
	var exactSecondaryNameOrgNoPrimaryName = endpointmanager.NPIOrganization{
		ID:                      5,
		NPI_ID:                  "5",
		Name:                    "Foo Bar",
		SecondaryName:           "foo bar baz",
		NormalizedName:          "",
		NormalizedSecondaryName: "FOO FOO BAR",
		Location: &endpointmanager.Location{
			Address1: "somerandomstring",
			Address2: "Foo Bar",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		Taxonomy: "208D00000X"}
	var nonMatchingOrg = endpointmanager.NPIOrganization{
		ID:                      6,
		NPI_ID:                  "6",
		Name:                    "nothingshouldmatchthis",
		SecondaryName:           "nothingshouldmatchthis",
		NormalizedName:          "NOTHINGSHOULDMATCHTHIS",
		NormalizedSecondaryName: "NOTHINGSHOULDMATCHTHIS",
		Location: &endpointmanager.Location{
			Address1: "somerandomstring",
			Address2: "FooBar",
			City:     "A",
			State:    "NH",
			ZipCode:  "00000"},
		Taxonomy: "208D00000X"}

	var orgs []endpointmanager.NPIOrganization

	matches, confidences, err := getIdsOfMatchingNPIOrgs(orgs, "FOO BAR", false)
	th.Assert(t, (err == nil), "Error getting matches from empty list")
	th.Assert(t, (len(matches) == 0), "There should not have been any matches returned got: "+strconv.Itoa(len(matches)))
	th.Assert(t, (len(confidences) == 0), "There should not have been any confidences returned"+strconv.Itoa(len(matches)))

	orgs = append(orgs, nonMatchingOrg)
	matches, confidences, err = getIdsOfMatchingNPIOrgs(orgs, "FOO BAR", false)
	th.Assert(t, (err == nil), "Error getting matches from list")
	th.Assert(t, (len(matches) == 0), "There should not have been any matches returned got: "+strconv.Itoa(len(matches)))
	th.Assert(t, (len(confidences) == 0), "There should not have been any confidences returned"+strconv.Itoa(len(matches)))

	orgs = append(orgs, exactPrimaryNameOrg)
	orgs = append(orgs, nonExactSecondaryNameOrg)
	orgs = append(orgs, exactSecondaryNameOrg)
	orgs = append(orgs, exactSecondaryNameOrgNoPrimaryName)

	matches, confidences, err = getIdsOfMatchingNPIOrgs(orgs, "FOO FOO BAR", false)
	th.Assert(t, (err == nil), "Error getting matches from list")
	th.Assert(t, (len(matches) == 4), "There should have been 3 matchs returned got: "+strconv.Itoa(len(matches)))
	th.Assert(t, (len(confidences) == 4), "There should have been 3 confidences returned "+strconv.Itoa(len(confidences)))
	confidence := fmt.Sprintf("%f", confidences[matches[0]])
	// FOO FOO BAR and primary name FOO FOO BAR have confidene of 1
	th.Assert(t, (confidence == "1.000000"), "Exact match confidence should have been 1.000000 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[1]])
	// FOO FOO BAR and secondary name FOO FOO BAR BAZ have confidene of .75
	th.Assert(t, (confidence == "0.750000"), "Exact match confidence should have been 0.750000 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[2]])
	// FOO FOO BAR and secondary name FOO FOO BAR have confidene of 1.000000
	th.Assert(t, (confidence == "1.000000"), "Exact match confidence should have been 1.000000 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[3]])
	// FOO FOO BAR and secondary name FOO FOO BAR have confidene of 1.000000
	th.Assert(t, (confidence == "1.000000"), "Exact match confidence should have been 1.000000 confidence got "+confidence)
}
