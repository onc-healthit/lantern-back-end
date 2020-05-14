package endpointlinker

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

var exactPrimaryNameOrg = &endpointmanager.NPIOrganization{
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
var nonExactSecondaryNameOrg = &endpointmanager.NPIOrganization{
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
var exactSecondaryNameOrg = &endpointmanager.NPIOrganization{
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
var exactSecondaryNameOrgNoPrimaryName = &endpointmanager.NPIOrganization{
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
var nonExactPrimaryNameOrgName = &endpointmanager.NPIOrganization{
	ID:                      6,
	NPI_ID:                  "6",
	Name:                    "Foo Bar",
	SecondaryName:           "foo bar baz",
	NormalizedName:          "FOO FOO BAR BAZ",
	NormalizedSecondaryName: "NOTHING SHOULD MATCH THIS",
	Location: &endpointmanager.Location{
		Address1: "somerandomstring",
		Address2: "Foo Bar",
		City:     "A City",
		State:    "AK",
		ZipCode:  "00000"},
	Taxonomy: "208D00000X"}
var nonMatchingOrg = &endpointmanager.NPIOrganization{
	ID:                      7,
	NPI_ID:                  "7",
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
var nonExactPrimaryAndSecondaryOrgName = &endpointmanager.NPIOrganization{
	ID:                      8,
	NPI_ID:                  "8",
	Name:                    "Foo Bar",
	SecondaryName:           "foo bar baz",
	NormalizedName:          "ONE TWO THREE FOUR FIVE SIX",
	NormalizedSecondaryName: "ONE TWO THREE FOUR FIVE SIX SEVEN",
	Location: &endpointmanager.Location{
		Address1: "somerandomstring",
		Address2: "Foo Bar",
		City:     "A City",
		State:    "AK",
		ZipCode:  "00000"},
	Taxonomy: "208D00000X"}

func Test_NormalizeOrgName(t *testing.T) {
	orgName := "AMBULANCE & and-chair. SERVICE!"
	expected := "AMBULANCE  AND CHAIR SERVICE"
	normalized, err := NormalizeOrgName(orgName)
	th.Assert(t, err == nil, err)
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
	var orgs []*endpointmanager.NPIOrganization

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
	orgs = append(orgs, nonExactPrimaryNameOrgName)
	orgs = append(orgs, nonExactPrimaryAndSecondaryOrgName)

	matches, confidences, err = getIdsOfMatchingNPIOrgs(orgs, "FOO FOO BAR", false)
	th.Assert(t, (err == nil), "Error getting matches from list")
	th.Assert(t, (len(matches) == 5), "There should have been 6 matchs returned got: "+strconv.Itoa(len(matches)))
	th.Assert(t, (len(confidences) == 5), "There should have been 6 confidences returned "+strconv.Itoa(len(confidences)))
	confidence := fmt.Sprintf("%f", confidences[matches[0]])
	// FOO FOO BAR and primary name FOO FOO BAR have confidence of 1
	th.Assert(t, (confidence == "1.000000"), "Exact match confidence should have been 1.000000 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[1]])
	// FOO FOO BAR and secondary name FOO FOO BAR BAZ have confidence of .75
	th.Assert(t, (confidence == "0.750000"), "Exact match confidence should have been 0.750000 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[2]])
	// FOO FOO BAR and secondary name FOO FOO BAR have confidence of 1.000000
	th.Assert(t, (confidence == "1.000000"), "Exact match confidence should have been 1.000000 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[3]])
	// FOO FOO BAR and secondary name FOO FOO BAR have confidence of 1.000000
	th.Assert(t, (confidence == "1.000000"), "Exact match confidence should have been 1.000000 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[4]])
	// FOO FOO BAR and primary name FOO FOO BAR BAZ have confidence of .75
	th.Assert(t, (confidence == "0.750000"), "Exact match confidence should have been 0.750000 confidence got "+confidence)

	// Test the case where the primary name and secondary name both pass threshold but one is greater than the other
	matches, confidences, err = getIdsOfMatchingNPIOrgs(orgs, "ONE TWO THREE FOUR FIVE SIX SEVEN EIGHT", false)
	th.Assert(t, (err == nil), "Error getting matches from list")
	th.Assert(t, (len(matches) == 1), "There should have been 6 matchs returned got: "+strconv.Itoa(len(matches)))
	th.Assert(t, (len(confidences) == 1), "There should have been 6 confidences returned "+strconv.Itoa(len(confidences)))
	confidence = fmt.Sprintf("%f", confidences[matches[0]])
	// ONE TWO THREE FOUR FIVE SIX SEVEN EIGHT and secondary name ONE TWO THREE FOUR FIVE SIX SEVEN should have confidence of .875000
	// .875 > than primary name ONE TWO THREE FOUR FIVE SIX match of .75
	th.Assert(t, (confidence == "0.875000"), "Exact match confidence should have been 0.875000 confidence got "+confidence)
}

func Test_mergeMatches(t *testing.T) {
	var allMatches []string
	var allConfidences map[string]float64
	var matches []string
	var confidences map[string]float64

	// test with uninitialized
	allMatches, allConfidences = mergeMatches(allMatches, allConfidences, matches, confidences)
	expected := 0
	th.Assert(t, len(allMatches) == expected, "expected no matches")
	th.Assert(t, len(allConfidences) == expected, "expected no matches")

	// test with initialized
	allMatches = make([]string, 0, 5)
	allConfidences = make(map[string]float64)
	matches = make([]string, 0, 5)
	confidences = make(map[string]float64)
	allMatches, allConfidences = mergeMatches(allMatches, allConfidences, matches, confidences)
	expected = 0
	th.Assert(t, len(allMatches) == expected, "expected no matches")
	th.Assert(t, len(allConfidences) == expected, "expected no matches")

	// test merging values into empty 'allMatches' and 'allConfidences'
	matches = []string{"1", "2", "3"}
	confidences = make(map[string]float64)
	confidences["1"] = 1.0
	confidences["2"] = .75
	confidences["3"] = .5
	allMatches, allConfidences = mergeMatches(allMatches, allConfidences, matches, confidences)
	expected = 3
	th.Assert(t, len(allMatches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(allMatches)))
	th.Assert(t, len(allConfidences) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(allMatches)))
	id := "1"
	conf := 1.0
	th.Assert(t, allConfidences[id] == conf, fmt.Sprintf("expected confidence for %s to be %f. got %f", id, conf, allConfidences[id]))
	id = "2"
	conf = .75
	th.Assert(t, allConfidences[id] == conf, fmt.Sprintf("expected confidence for %s to be %f. got %f", id, conf, allConfidences[id]))
	id = "3"
	conf = .5
	th.Assert(t, allConfidences[id] == conf, fmt.Sprintf("expected confidence for %s to be %f. got %f", id, conf, allConfidences[id]))

	// test adding new
	matches = []string{"4"}
	confidences = make(map[string]float64)
	confidences["4"] = .6
	allMatches, allConfidences = mergeMatches(allMatches, allConfidences, matches, confidences)
	expected = 4
	th.Assert(t, len(allMatches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(allMatches)))
	th.Assert(t, len(allConfidences) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(allMatches)))
	id = "1"
	conf = 1.0
	th.Assert(t, allConfidences[id] == conf, fmt.Sprintf("expected confidence for %s to be %f. got %f", id, conf, allConfidences[id]))
	id = "2"
	conf = .75
	th.Assert(t, allConfidences[id] == conf, fmt.Sprintf("expected confidence for %s to be %f. got %f", id, conf, allConfidences[id]))
	id = "3"
	conf = .5
	th.Assert(t, allConfidences[id] == conf, fmt.Sprintf("expected confidence for %s to be %f. got %f", id, conf, allConfidences[id]))
	id = "4"
	conf = .6
	th.Assert(t, allConfidences[id] == conf, fmt.Sprintf("expected confidence for %s to be %f. got %f", id, conf, allConfidences[id]))

	// test updating with both higher, lower, same, and new confidences
	matches = []string{"2", "3", "4", "5"}
	confidences = make(map[string]float64)
	confidences["2"] = .5
	confidences["3"] = 1.0
	confidences["4"] = .6
	confidences["5"] = .8
	allMatches, allConfidences = mergeMatches(allMatches, allConfidences, matches, confidences)
	expected = 5
	th.Assert(t, len(allMatches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(allMatches)))
	th.Assert(t, len(allConfidences) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(allMatches)))
	id = "1"
	conf = 1.0
	th.Assert(t, allConfidences[id] == conf, fmt.Sprintf("expected confidence for %s to be %f. got %f", id, conf, allConfidences[id]))
	id = "2"
	conf = .75
	th.Assert(t, allConfidences[id] == conf, fmt.Sprintf("expected confidence for %s to be %f. got %f", id, conf, allConfidences[id]))
	id = "3"
	conf = 1.0
	th.Assert(t, allConfidences[id] == conf, fmt.Sprintf("expected confidence for %s to be %f. got %f", id, conf, allConfidences[id]))
	id = "4"
	conf = .6
	th.Assert(t, allConfidences[id] == conf, fmt.Sprintf("expected confidence for %s to be %f. got %f", id, conf, allConfidences[id]))
	id = "5"
	conf = .8
	th.Assert(t, allConfidences[id] == conf, fmt.Sprintf("expected confidence for %s to be %f. got %f", id, conf, allConfidences[id]))
}

func Test_matchByName(t *testing.T) {
	var orgs []*endpointmanager.NPIOrganization

	var ep = &endpointmanager.FHIREndpoint{
		ID:                1,
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{"FOO FOO BAR"},
		NPIIDs:            []string{"1", "2", "3"},
		ListSource:        "https://open.epic.com/MyApps/EndpointsJson"}

	// test with no orgs
	matches, confidences, err := matchByName(ep, orgs, false)
	expected := 0
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, "expected no matches")
	th.Assert(t, len(confidences) == expected, "expected no confidences")

	orgs = append(orgs, nonMatchingOrg)

	// test with non matching org
	matches, confidences, err = matchByName(ep, orgs, false)
	expected = 0
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, "expected no matches")
	th.Assert(t, len(confidences) == expected, "expected no confidences")

	orgs = append(orgs, exactPrimaryNameOrg)
	orgs = append(orgs, nonExactSecondaryNameOrg)
	orgs = append(orgs, exactSecondaryNameOrg)
	orgs = append(orgs, exactSecondaryNameOrgNoPrimaryName)
	orgs = append(orgs, nonExactPrimaryNameOrgName)
	orgs = append(orgs, nonExactPrimaryAndSecondaryOrgName)

	// expect some matches with varying confidences to "FOO FOO BAR"
	matches, confidences, err = matchByName(ep, orgs, false)
	expected = 5
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	th.Assert(t, len(confidences) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	org := exactPrimaryNameOrg
	expectedConf := 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactSecondaryNameOrg
	expectedConf = .75
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrg
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrgNoPrimaryName
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactPrimaryNameOrgName
	expectedConf = .75
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))

	// expect some matches with varying confidences to "FOO FOO BAR BAZ"
	ep.OrganizationNames = []string{"FOO FOO BAR BAZ"}
	matches, confidences, err = matchByName(ep, orgs, false)
	expected = 5
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	th.Assert(t, len(confidences) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	org = exactPrimaryNameOrg
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactSecondaryNameOrg
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrg
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrgNoPrimaryName
	expectedConf = .75
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactPrimaryNameOrgName
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))

	// check that highest confidence value is used
	// expect some matches with varying confidences to "FOO FOO BAR BAZ" and "FOO FOO BAR"
	ep.OrganizationNames = []string{"FOO FOO BAR BAZ", "FOO FOO BAR"}
	matches, confidences, err = matchByName(ep, orgs, false)
	expected = 5
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	th.Assert(t, len(confidences) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	org = exactPrimaryNameOrg
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactSecondaryNameOrg
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrg
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrgNoPrimaryName
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactPrimaryNameOrgName
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))

	// checking non-existent org name causes no issues
	// expect some matches with varying confidences to "FOO FOO BAR BAZ" and "FOO FOO BAR" and "BLAH"
	ep.OrganizationNames = []string{"FOO FOO BAR BAZ", "FOO FOO BAR", "BLAH"}
	matches, confidences, err = matchByName(ep, orgs, false)
	expected = 5
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	th.Assert(t, len(confidences) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	org = exactPrimaryNameOrg
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactSecondaryNameOrg
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrg
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrgNoPrimaryName
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactPrimaryNameOrgName
	expectedConf = 1.0
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
}
