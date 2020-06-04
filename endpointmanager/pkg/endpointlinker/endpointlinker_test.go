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
	NormalizedName:          "FOO FOO BAR BAR BAZ BAZ BAM",
	NormalizedSecondaryName: "FOO FOO BAR BAR BAZ BAZ BAM BAM",
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
	NormalizedSecondaryName: "FOO FOO BAR BAR BAZ BAZ BAM BAM",
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
	NormalizedName:          "FOO FOO BAR BAR BAZ BAZ BAM BAM",
	NormalizedSecondaryName: "FOO FOO BAR BAR BAZ BAZ BAM",
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
	NormalizedSecondaryName: "FOO FOO BAR BAR BAZ BAZ BAM",
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
	NormalizedName:          "FOO FOO BAR BAR BAZ BAZ BAM BAM",
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

var tokenValues = map[string]float64{
	"FOO":                    1.0,
	"BAR":                    1.0,
	"BAZ":                    1.0,
	"BAM":                    1.0,
	"NOTHING":                1.0,
	"SHOULD":                 1.0,
	"MATCH":                  1.0,
	"THIS":                   1.0,
	"NOTHINGSHOULDMATCHTHIS": 1.0,
	"ONE":                    1.0,
	"TWO":                    1.0,
	"THREE":                  1.0,
	"FOUR":                   1.0,
	"FIVE":                   1.0,
	"SIX":                    1.0,
	"SEVEN":                  1.0,
	"EIGHT":                  1.0}

func Test_NormalizeOrgName(t *testing.T) {
	orgName := "AMBULANCE & and-chair. SERVICE!"
	expected := "AMBULANCE  AND CHAIR SERVICE"
	normalized, err := NormalizeOrgName(orgName)
	th.Assert(t, err == nil, err)
	th.Assert(t, (normalized == expected), "Organization name normalization failed. Expected: "+expected+" Got: "+normalized)
}

func Test_calculateJaccardIndex(t *testing.T) {
	jaccardIndex := calculateJaccardIndex("FOO BAR", "FOO BAR", tokenValues)
	ind := strconv.FormatFloat(jaccardIndex, 'f', -1, 64)
	th.Assert(t, (jaccardIndex == 1), "Jaccard index expected to be 1, was "+ind)

	jaccardIndex = calculateJaccardIndex("FOO BAZ BAR", "FOO BAR", tokenValues)
	ind = strconv.FormatFloat(jaccardIndex, 'f', -1, 64)
	th.Assert(t, (jaccardIndex == .6666666666666666), "Jaccard index expected to be .6666666666666666, was "+ind)

	jaccardIndex = calculateJaccardIndex("FOO FOO BAR", "FOO BAR", tokenValues)
	ind = strconv.FormatFloat(jaccardIndex, 'f', -1, 64)
	th.Assert(t, (jaccardIndex == .6666666666666666), "Jaccard index expected to be .6666666666666666, was "+ind)
}

func Test_IntersectionCount(t *testing.T) {
	emptyListIntersections, emptyListDenom := intersectionCount([]string{}, []string{}, tokenValues)
	th.Assert(t, (int(emptyListIntersections) == 0 && int(emptyListDenom) == 0), "Intersection count of empty lists should be zero, got "+strconv.Itoa(int(emptyListIntersections))+", Denominator of empty lists should be zero, got "+strconv.Itoa(int(emptyListDenom)))

	emptyListIntersections, emptyListDenom = intersectionCount([]string{"FOO"}, []string{}, tokenValues)
	th.Assert(t, (int(emptyListIntersections) == 0 && int(emptyListDenom) == 1), "Intersection count of empty lists should be zero, got "+strconv.Itoa(int(emptyListIntersections))+", Denominator of empty lists should be one, got "+strconv.Itoa(int(emptyListDenom)))

	emptyListIntersections, emptyListDenom = intersectionCount([]string{"FOO"}, []string{"BAR"}, tokenValues)
	th.Assert(t, (int(emptyListIntersections) == 0 && int(emptyListDenom) == 2), "Intersection count of empty lists should be zero, got "+strconv.Itoa(int(emptyListIntersections))+", Denominator of empty lists should be two, got "+strconv.Itoa(int(emptyListDenom)))

	nonEmptyListIntersections, nonEmptyListDenom := intersectionCount([]string{"FOO"}, []string{"FOO"}, tokenValues)
	th.Assert(t, (int(nonEmptyListIntersections) == 1 && int(nonEmptyListDenom) == 1), "Intersection count of non-empty lists should be one, got "+strconv.Itoa(int(nonEmptyListIntersections))+", Denominator of non-empty lists should be one, got "+strconv.Itoa(int(nonEmptyListDenom)))

	nonEmptyListIntersections, nonEmptyListDenom = intersectionCount([]string{"FOO", "BAR"}, []string{"BAR"}, tokenValues)
	th.Assert(t, (int(nonEmptyListIntersections) == 1 && int(nonEmptyListDenom) == 2), "Intersection count of non-empty lists should be one, got "+strconv.Itoa(int(nonEmptyListIntersections))+", Denominator of non-empty lists should be two, got "+strconv.Itoa(int(nonEmptyListDenom)))

	nonEmptyListIntersections, nonEmptyListDenom = intersectionCount([]string{"FOO", "BAR"}, []string{"BAR", "FOO"}, tokenValues)
	th.Assert(t, (int(nonEmptyListIntersections) == 2 && int(nonEmptyListDenom) == 2), "Intersection count of non-empty lists should be two, got "+strconv.Itoa(int(nonEmptyListIntersections))+", Denominator of non-empty lists should be two, got "+strconv.Itoa(int(nonEmptyListDenom)))

	nonEmptyListIntersections, nonEmptyListDenom = intersectionCount([]string{"FOO", "BAR", "FOO", "FOO"}, []string{"BAR", "FOO", "FOO"}, tokenValues)
	th.Assert(t, (int(nonEmptyListIntersections) == 3 && int(nonEmptyListDenom) == 4), "Intersection count of non-empty lists should be three, got "+strconv.Itoa(int(nonEmptyListIntersections))+", Denominator of non-empty lists should be four, got "+strconv.Itoa(int(nonEmptyListDenom)))
}

func Test_getIdsOfMatchingNPIOrgs(t *testing.T) {
	var orgs []*endpointmanager.NPIOrganization

	matches, confidences, err := getIdsOfMatchingNPIOrgs(orgs, "FOO BAR", false, tokenValues, .85)
	th.Assert(t, (err == nil), "Error getting matches from empty list")
	th.Assert(t, (len(matches) == 0), "There should not have been any matches returned got: "+strconv.Itoa(len(matches)))
	th.Assert(t, (len(confidences) == 0), "There should not have been any confidences returned"+strconv.Itoa(len(matches)))

	orgs = append(orgs, nonMatchingOrg)
	matches, confidences, err = getIdsOfMatchingNPIOrgs(orgs, "FOO BAR", false, tokenValues, .85)
	th.Assert(t, (err == nil), "Error getting matches from list")
	th.Assert(t, (len(matches) == 0), "There should not have been any matches returned got: "+strconv.Itoa(len(matches)))
	th.Assert(t, (len(confidences) == 0), "There should not have been any confidences returned"+strconv.Itoa(len(matches)))

	orgs = append(orgs, exactPrimaryNameOrg)
	orgs = append(orgs, nonExactSecondaryNameOrg)
	orgs = append(orgs, exactSecondaryNameOrg)
	orgs = append(orgs, exactSecondaryNameOrgNoPrimaryName)
	orgs = append(orgs, nonExactPrimaryNameOrgName)
	orgs = append(orgs, nonExactPrimaryAndSecondaryOrgName)

	matches, confidences, err = getIdsOfMatchingNPIOrgs(orgs, "FOO FOO BAR BAR BAZ BAZ BAM", false, tokenValues, .85)
	th.Assert(t, (err == nil), "Error getting matches from list")
	th.Assert(t, (len(matches) == 5), "There should have been 5 matches returned got: "+strconv.Itoa(len(matches)))
	th.Assert(t, (len(confidences) == 5), "There should have been 5 confidences returned "+strconv.Itoa(len(confidences)))
	confidence := fmt.Sprintf("%f", confidences[matches[0]])
	// FOO FOO BAR BAR BAZ BAZ BAM and primary name FOO FOO BAR BAR BAZ BAZ BAM have confidence of 1 * .9
	th.Assert(t, (confidence == "0.900000"), "Exact match confidence should have been 0.900000 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[1]])
	// FOO FOO BAR BAR BAZ BAZ BAM and secondary name FOO FOO BAR BAR BAZ BAZ BAM BAM have confidence of .875 * .9
	th.Assert(t, (confidence == "0.787500"), "Exact match confidence should have been 0.787500 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[2]])
	// FOO FOO BAR BAR BAZ BAZ BAM and secondary name FOO FOO BAR BAR BAZ BAZ BAM have confidence of 1.000000 * .9
	th.Assert(t, (confidence == "0.900000"), "Exact match confidence should have been 0.900000 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[3]])
	// FOO FOO BAR BAR BAZ BAZ BAM and secondary name FOO FOO BAR BAR BAZ BAZ BAM have confidence of 1.000000 * .9
	th.Assert(t, (confidence == "0.900000"), "Exact match confidence should have been 0.900000 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[4]])
	// FOO FOO BAR BAR BAZ BAZ BAM and primary name FOO FOO BAR BAR BAZ BAZ BAM BAM have confidence of .875 * .9
	th.Assert(t, (confidence == "0.787500"), "Exact match confidence should have been 0.787500 confidence got "+confidence)

	// Test the case where the primary name and secondary name both pass threshold but one is greater than the other
	matches, confidences, err = getIdsOfMatchingNPIOrgs(orgs, "ONE TWO THREE FOUR FIVE SIX SEVEN EIGHT", false, tokenValues, .85)
	th.Assert(t, (err == nil), "Error getting matches from list")
	th.Assert(t, (len(matches) == 1), "There should have been 1 matchs returned got: "+strconv.Itoa(len(matches)))
	th.Assert(t, (len(confidences) == 1), "There should have been 1 confidences returned "+strconv.Itoa(len(confidences)))
	confidence = fmt.Sprintf("%f", confidences[matches[0]])
	// ONE TWO THREE FOUR FIVE SIX SEVEN EIGHT and secondary name ONE TWO THREE FOUR FIVE SIX SEVEN should have confidence of .875000 * .9
	// .875 *.9 > than primary name ONE TWO THREE FOUR FIVE SIX match of .75 * .9
	th.Assert(t, (confidence == "0.787500"), "Exact match confidence should have been 0.787500 confidence got "+confidence)
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
		OrganizationNames: []string{"FOO FOO BAR BAR BAZ BAZ BAM"},
		NPIIDs:            []string{"1", "2", "3"},
		ListSource:        "https://open.epic.com/MyApps/EndpointsJson"}

	// test with no orgs
	matches, confidences, err := matchByName(ep, orgs, false, tokenValues, .85)
	expected := 0
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, "expected no matches")
	th.Assert(t, len(confidences) == expected, "expected no confidences")

	orgs = append(orgs, nonMatchingOrg)

	// test with non matching org
	matches, confidences, err = matchByName(ep, orgs, false, tokenValues, .85)
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

	// expect some matches with varying confidences to "FOO FOO BAR BAR BAZ BAZ BAM"
	matches, confidences, err = matchByName(ep, orgs, false, tokenValues, .85)
	expected = 5
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	th.Assert(t, len(confidences) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	org := exactPrimaryNameOrg
	expectedConf := 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactSecondaryNameOrg
	expectedConf = .875 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrg
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrgNoPrimaryName
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactPrimaryNameOrgName
	expectedConf = .875 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))

	// expect some matches with varying confidences to "FOO FOO BAR BAR BAZ BAZ BAM BAM"
	ep.OrganizationNames = []string{"FOO FOO BAR BAR BAZ BAZ BAM BAM"}
	matches, confidences, err = matchByName(ep, orgs, false, tokenValues, .85)
	expected = 5
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	th.Assert(t, len(confidences) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	org = exactPrimaryNameOrg
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactSecondaryNameOrg
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrg
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrgNoPrimaryName
	expectedConf = .875 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactPrimaryNameOrgName
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))

	// check that highest confidence value is used
	// expect some matches with varying confidences to "FOO FOO BAR BAR BAZ BAZ BAM BAM" and "FOO FOO BAR BAR BAZ BAZ BAM"
	ep.OrganizationNames = []string{"FOO FOO BAR BAR BAZ BAZ BAM BAM", "FOO FOO BAR BAR BAZ BAZ BAM"}
	matches, confidences, err = matchByName(ep, orgs, false, tokenValues, .85)
	expected = 5
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	th.Assert(t, len(confidences) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	org = exactPrimaryNameOrg
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactSecondaryNameOrg
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrg
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrgNoPrimaryName
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactPrimaryNameOrgName
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))

	// checking non-existent org name causes no issues
	// expect some matches with varying confidences to "FOO FOO BAR BAR BAZ BAZ BAM BAM" and "FOO FOO BAR BAR BAZ BAZ BAM" and "BLAH"
	ep.OrganizationNames = []string{"FOO FOO BAR BAR BAZ BAZ BAM BAM", "FOO FOO BAR BAR BAZ BAZ BAM", "BLAH"}
	matches, confidences, err = matchByName(ep, orgs, false, tokenValues, .85)
	expected = 5
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	th.Assert(t, len(confidences) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	org = exactPrimaryNameOrg
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactSecondaryNameOrg
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrg
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrgNoPrimaryName
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactPrimaryNameOrgName
	expectedConf = 1.0 * .9
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
}
