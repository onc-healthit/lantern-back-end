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

func Test_calculateWeightedJaccardIndex(t *testing.T) {
	jaccardIndex := calculateWeightedJaccardIndex("FOO BAR", "FOO BAR", tokenValues)
	ind := strconv.FormatFloat(jaccardIndex, 'f', -1, 64)
	th.Assert(t, (jaccardIndex == 1), "Jaccard index expected to be 1, was "+ind)

	jaccardIndex = calculateWeightedJaccardIndex("FOO BAZ BAR", "FOO BAR", tokenValues)
	ind = strconv.FormatFloat(jaccardIndex, 'f', -1, 64)
	th.Assert(t, (jaccardIndex == .6666666666666666), "Jaccard index expected to be .6666666666666666, was "+ind)

	jaccardIndex = calculateWeightedJaccardIndex("FOO FOO BAR", "FOO BAR", tokenValues)
	ind = strconv.FormatFloat(jaccardIndex, 'f', -1, 64)
	th.Assert(t, (jaccardIndex == .6666666666666666), "Jaccard index expected to be .6666666666666666, was "+ind)
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
	// FOO FOO BAR BAR BAZ BAZ BAM and primary name FOO FOO BAR BAR BAZ BAZ BAM have confidence of 1 * .99
	th.Assert(t, (confidence == "0.990000"), "Exact match confidence should have been 0.990000 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[1]])
	// FOO FOO BAR BAR BAZ BAZ BAM and secondary name FOO FOO BAR BAR BAZ BAZ BAM BAM have confidence of .875 * .99
	th.Assert(t, (confidence == "0.866250"), "Exact match confidence should have been 0.866250 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[2]])
	// FOO FOO BAR BAR BAZ BAZ BAM and secondary name FOO FOO BAR BAR BAZ BAZ BAM have confidence of 1.000000 * .99
	th.Assert(t, (confidence == "0.990000"), "Exact match confidence should have been 0.990000 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[3]])
	// FOO FOO BAR BAR BAZ BAZ BAM and secondary name FOO FOO BAR BAR BAZ BAZ BAM have confidence of 1.000000 * .99
	th.Assert(t, (confidence == "0.990000"), "Exact match confidence should have been 0.990000 confidence got "+confidence)
	confidence = fmt.Sprintf("%f", confidences[matches[4]])
	// FOO FOO BAR BAR BAZ BAZ BAM and primary name FOO FOO BAR BAR BAZ BAZ BAM BAM have confidence of .875 * .99
	th.Assert(t, (confidence == "0.866250"), "Exact match confidence should have been 0.866250 confidence got "+confidence)

	// Test the case where the primary name and secondary name both pass threshold but one is greater than the other
	matches, confidences, err = getIdsOfMatchingNPIOrgs(orgs, "ONE TWO THREE FOUR FIVE SIX SEVEN EIGHT", false, tokenValues, .85)
	th.Assert(t, (err == nil), "Error getting matches from list")
	th.Assert(t, (len(matches) == 1), "There should have been 1 matchs returned got: "+strconv.Itoa(len(matches)))
	th.Assert(t, (len(confidences) == 1), "There should have been 1 confidences returned "+strconv.Itoa(len(confidences)))
	confidence = fmt.Sprintf("%f", confidences[matches[0]])
	// ONE TWO THREE FOUR FIVE SIX SEVEN EIGHT and secondary name ONE TWO THREE FOUR FIVE SIX SEVEN should have confidence of .875000 * .99
	// .875 *.99 > than primary name ONE TWO THREE FOUR FIVE SIX match of .75 * .99
	th.Assert(t, (confidence == "0.866250"), "Exact match confidence should have been 0.866250 confidence got "+confidence)
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
		ListSource:        "https://open.epic.com/Endpoints/DSTU2"}

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
	expectedConf := 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactSecondaryNameOrg
	expectedConf = .875 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrg
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrgNoPrimaryName
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactPrimaryNameOrgName
	expectedConf = .875 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))

	// expect some matches with varying confidences to "FOO FOO BAR BAR BAZ BAZ BAM BAM"
	ep.OrganizationNames = []string{"FOO FOO BAR BAR BAZ BAZ BAM BAM"}
	matches, confidences, err = matchByName(ep, orgs, false, tokenValues, .85)
	expected = 5
	th.Assert(t, err == nil, err)
	th.Assert(t, len(matches) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	th.Assert(t, len(confidences) == expected, fmt.Sprintf("expected %d matches. got %d.", expected, len(matches)))
	org = exactPrimaryNameOrg
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactSecondaryNameOrg
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrg
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrgNoPrimaryName
	expectedConf = .875 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactPrimaryNameOrgName
	expectedConf = 1.0 * .99
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
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactSecondaryNameOrg
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrg
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrgNoPrimaryName
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactPrimaryNameOrgName
	expectedConf = 1.0 * .99
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
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactSecondaryNameOrg
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrg
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = exactSecondaryNameOrgNoPrimaryName
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
	org = nonExactPrimaryNameOrgName
	expectedConf = 1.0 * .99
	th.Assert(t, confidences[org.NPI_ID] == expectedConf, fmt.Sprintf("Expected %s/%s to match %v with confidence %f. got %f", org.NormalizedName, org.NormalizedSecondaryName, ep.OrganizationNames, expectedConf, confidences[org.NPI_ID]))
}

func Test_countTokens(t *testing.T) {
	var npiOrgs []*endpointmanager.NPIOrganization
	npiOrgs = append(npiOrgs, exactPrimaryNameOrg)
	npiOrgs = append(npiOrgs, nonExactSecondaryNameOrg)
	npiOrgs = append(npiOrgs, exactSecondaryNameOrg)
	npiOrgs = append(npiOrgs, exactSecondaryNameOrgNoPrimaryName)
	npiOrgs = append(npiOrgs, nonExactPrimaryNameOrgName)
	npiOrgs = append(npiOrgs, nonExactPrimaryAndSecondaryOrgName)

	var FHIREndpoints []*endpointmanager.FHIREndpoint

	var ep1 = &endpointmanager.FHIREndpoint{
		ID:                1,
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{"FOO FOO BAR BAR BAZ BAZ BAM", "BLAH", "ONE TWO THREE FOUR FIVE"},
		NPIIDs:            []string{"1", "2", "3"},
		ListSource:        "https://open.epic.com/Endpoints/DSTU2"}

	var ep2 = &endpointmanager.FHIREndpoint{
		ID:                1,
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{"FOO FOO", "BAM BLAH", "FOO FOO BAR BAR BAZ BAZ BAM BAM BAM", "SYSTEM SYSTEM SERVICES"},
		NPIIDs:            []string{"1", "2", "3"},
		ListSource:        "https://open.epic.com/Endpoints/DSTU2"}

	var ep3 = &endpointmanager.FHIREndpoint{
		ID:                1,
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{"BLAH EIGHT", "EIGHT NINE TEN", "FOO NOTHING BAM BAM", "SYSTEM SERVICES BLAH SERVICES"},
		NPIIDs:            []string{"1", "2", "3"},
		ListSource:        "https://open.epic.com/Endpoints/DSTU2"}

	FHIREndpoints = append(FHIREndpoints, ep1)
	FHIREndpoints = append(FHIREndpoints, ep2)
	FHIREndpoints = append(FHIREndpoints, ep3)
	tokenCountsAll, NPITokenCounts, EndpointTokenCounts, firstKey := countTokens(npiOrgs, FHIREndpoints)

	expectedFirstKey := "FOO"

	th.Assert(t, expectedFirstKey == firstKey, fmt.Sprintf("expected first key to be 'FOO', but got %s.", firstKey))

	th.Assert(t, tokenCountsAll["FOO"] == 21, fmt.Sprintf("expected FOO count to be 21, but got %v.", tokenCountsAll["FOO"]))
	th.Assert(t, NPITokenCounts["FOO"] == 14, fmt.Sprintf("expected FOO NPI count to be 14, but got %v.", NPITokenCounts["FOO"]))
	th.Assert(t, EndpointTokenCounts["FOO"] == 7, fmt.Sprintf("expected FOO Endpoint count to be 7, but got %v.", EndpointTokenCounts["FOO"]))
	th.Assert(t, NPITokenCounts["SEVEN"] == 1, fmt.Sprintf("expected SEVEN NPI count to be 1, but got %v.", NPITokenCounts["SEVEN"]))
	th.Assert(t, EndpointTokenCounts["SEVEN"] == 0, fmt.Sprintf("expected SEVEN Endpoint count to be 0, but got %v.", EndpointTokenCounts["SEVEN"]))
	th.Assert(t, NPITokenCounts["BLAH"] == 0, fmt.Sprintf("expected BLAH NPI count to be 0, but got %v.", NPITokenCounts["BLAH"]))
	th.Assert(t, EndpointTokenCounts["BLAH"] == 4, fmt.Sprintf("expected BLAH Endpoint count to be 4, but got %v.", EndpointTokenCounts["BLAH"]))
	th.Assert(t, tokenCountsAll[""] == 0, fmt.Sprintf("expected empty string count to be 0, but got %v.", tokenCountsAll[""]))
	th.Assert(t, NPITokenCounts[""] == 0, fmt.Sprintf("expected empty string count to be 0, but got %v.", NPITokenCounts[""]))
	th.Assert(t, EndpointTokenCounts[""] == 0, fmt.Sprintf("expected empty string count to be 0, but got %v.", EndpointTokenCounts[""]))

}

func Test_computeTokenValues(t *testing.T) {
	var tokenCountsAll = map[string]int{
		"FOO":            600,
		"BAR":            18,
		"BAZ":            70,
		"BAM":            50,
		"NOTHING":        5,
		"SHOULD":         2,
		"MATCH":          45,
		"THIS":           300,
		"EMPTY":          9,
		"CORP":           400,
		"BLAH":           30,
		"ONLYINENDPOINT": 25,
		"ONLYINNPI":      58}

	var tokenCountsNPI = map[string]int{
		"FOO":       250,
		"BAR":       16,
		"BAZ":       30,
		"BAM":       25,
		"NOTHING":   3,
		"SHOULD":    1,
		"MATCH":     5,
		"THIS":      150,
		"EMPTY":     8,
		"CORP":      200,
		"BLAH":      16,
		"ONLYINNPI": 58}

	var tokenCountsEndpoints = map[string]int{
		"FOO":            350,
		"BAR":            2,
		"BAZ":            40,
		"BAM":            25,
		"NOTHING":        2,
		"SHOULD":         1,
		"MATCH":          40,
		"THIS":           150,
		"EMPTY":          1,
		"CORP":           200,
		"BLAH":           14,
		"ONLYINENDPOINT": 25}

	firstKey := "FOO"
	randTokenMean := 5
	randStandardDev := 5

	corpVal := 1.0 - (400.0 / 600.0)
	shouldVal := 1.0 - (2.0 / 600.0)
	nothingVal := 1.0 - (5.0 / 600.0)
	emptyVal := 1.0 - (9.0 / 600.0)
	barVal := 1.0 - (18.0 / 600.0)
	blahVal := 1.0 - (30.0 / 600.0)
	matchVal := 1.0 - (45.0 / 600.0)
	thisVal := 1.0 - (300.0 / 600.0)
	onlyinnpiVal := 1.0 - (58.0 / 600.0)
	onlyinendpointVal := 1.0 - (25.0 / 600.0)

	tokenVals := computeTokenValues(tokenCountsAll, tokenCountsNPI, tokenCountsEndpoints, firstKey, randTokenMean, randStandardDev)

	// CORP token found in fluff dictionary (multiply by 0.2)
	tokenValResult := fmt.Sprintf("%f", tokenVals["CORP"])
	expectedResult := fmt.Sprintf("%f", corpVal*0.2)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("CORP expected %s value. got %s.", expectedResult, tokenValResult))

	// SHOULD token count < mean (multiply by 2.5)
	tokenValResult = fmt.Sprintf("%f", tokenVals["SHOULD"])
	expectedResult = fmt.Sprintf("%f", shouldVal*2.5)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("SHOULD expected %s value. got %s.", expectedResult, tokenValResult))

	// NOTHING token count < mean + (standardDev/3) (multiply by 1.6)
	tokenValResult = fmt.Sprintf("%f", tokenVals["NOTHING"])
	expectedResult = fmt.Sprintf("%f", nothingVal*1.6)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("NOTHING expected %s value. got %s.", expectedResult, tokenValResult))

	// EMPTY token count < mean + standardDev (multiply by 1.3)
	tokenValResult = fmt.Sprintf("%f", tokenVals["EMPTY"])
	expectedResult = fmt.Sprintf("%f", emptyVal*1.3)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("EMPTY expected %s value. got %s.", expectedResult, tokenValResult))

	// BAR token count < mean + (standardDev*3) (multiply by 1.0)
	tokenValResult = fmt.Sprintf("%f", tokenVals["BAR"])
	expectedResult = fmt.Sprintf("%f", barVal*1.0)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("BAR expected %s value. got %s.", expectedResult, tokenValResult))

	// BLAH token count < mean + (standardDev*6) (multiply by 0.8)
	tokenValResult = fmt.Sprintf("%f", tokenVals["BLAH"])
	expectedResult = fmt.Sprintf("%f", blahVal*0.8)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("BLAH expected %s value. got %s.", expectedResult, tokenValResult))

	// MATCH token count < mean + (standardDev*9) (multiply by 0.6)
	tokenValResult = fmt.Sprintf("%f", tokenVals["MATCH"])
	expectedResult = fmt.Sprintf("%f", matchVal*0.6)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("MATCH expected %s value. got %s.", expectedResult, tokenValResult))

	// THIS token count > mean + (standardDev*9) (multiply by 0.4)
	tokenValResult = fmt.Sprintf("%f", tokenVals["THIS"])
	expectedResult = fmt.Sprintf("%f", thisVal*0.4)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("THIS expected %s value. got %s.", expectedResult, tokenValResult))

	// ONLYINNPI token count > mean + (standardDev*9) and only found in NPI token count (multiply by 0.4 and then 0.3)
	tokenValResult = fmt.Sprintf("%f", tokenVals["ONLYINNPI"])
	expectedResult = fmt.Sprintf("%f", onlyinnpiVal*0.4*0.3)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("ONLYINNPI expected %s value. got %s.", expectedResult, tokenValResult))

	// ONLYINENDPOINT token count < mean + (standardDev*6) and only found in Endpoint token count (multiply by 0.8 and then 0.3)
	tokenValResult = fmt.Sprintf("%f", tokenVals["ONLYINENDPOINT"])
	expectedResult = fmt.Sprintf("%f", onlyinendpointVal*0.8*0.3)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("ONLYINENDPOINT expected %s value. got %s.", expectedResult, tokenValResult))

	// Word that does not exist in token counts should return 0 for a value
	tokenValResult = fmt.Sprintf("%f", tokenVals["NOTINLIST"])
	expectedResult = fmt.Sprintf("%f", 0.0)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("NOTINLIST expected %s value. got %s.", expectedResult, tokenValResult))

}

func Test_getTokenVals(t *testing.T) {
	var npiOrgs []*endpointmanager.NPIOrganization
	npiOrgs = append(npiOrgs, exactPrimaryNameOrg)
	npiOrgs = append(npiOrgs, nonExactSecondaryNameOrg)
	npiOrgs = append(npiOrgs, exactSecondaryNameOrg)
	npiOrgs = append(npiOrgs, exactSecondaryNameOrgNoPrimaryName)
	npiOrgs = append(npiOrgs, nonExactPrimaryNameOrgName)
	npiOrgs = append(npiOrgs, nonExactPrimaryAndSecondaryOrgName)

	var FHIREndpoints []*endpointmanager.FHIREndpoint

	var ep1 = &endpointmanager.FHIREndpoint{
		ID:                1,
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{"FOO FOO BAR BAR BAZ BAZ BAM", "BLAH", "ONE TWO THREE FOUR FIVE"},
		NPIIDs:            []string{"1", "2", "3"},
		ListSource:        "https://open.epic.com/Endpoints/DSTU2"}

	var ep2 = &endpointmanager.FHIREndpoint{
		ID:                1,
		URL:               "example.com/FHIR/DSTU2",
		OrganizationNames: []string{"FOO FOO", "BAM BLAH", "FOO FOO BAR BAR BAZ BAZ BAM BAM BAM", "SYSTEM SYSTEM SERVICES"},
		NPIIDs:            []string{"1", "2", "3"},
		ListSource:        "https://open.epic.com/Endpoints/DSTU2"}

	FHIREndpoints = append(FHIREndpoints, ep1)
	FHIREndpoints = append(FHIREndpoints, ep2)

	// Total value of tokens = 105, total value of unique tokens = 19, mean = 6 standard deviation = 6
	// Divide by 20 as this is the highest token count
	fooValue := 0.0 // (1.0 - 20.0/20.0)
	bamValue := 1.0 - (16.0 / 20.0)
	fiveValue := 1.0 - (3.0 / 20.0)
	sevenValue := 1.0 - (1.0 / 20.0)
	blahValue := 1.0 - (2.0 / 20.0)
	systemValue := 1.0 - (2.0 / 20.0)

	tokenVals := getTokenVals(npiOrgs, FHIREndpoints)

	// Foo with count of 20 has value 0.0, and 20 < mean + standardDev*3 so multiplied by 1.0
	tokenValResult := fmt.Sprintf("%f", tokenVals["FOO"])
	expectedResult := fmt.Sprintf("%f", fooValue*1.0)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("expected %s value. got %s.", expectedResult, tokenValResult))
	// Bam with count of 16 has value 0.2, and 16 < mean + standardDev*3 so multiplied by 1.0
	tokenValResult = fmt.Sprintf("%f", tokenVals["BAM"])
	expectedResult = fmt.Sprintf("%f", bamValue*1.0)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("expected %s value. got %s.", expectedResult, tokenValResult))
	// Five with count of 3 has value 0.85, and 3 < mean so multiplied by 2.5
	tokenValResult = fmt.Sprintf("%f", tokenVals["FIVE"])
	expectedResult = fmt.Sprintf("%f", fiveValue*2.5)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("expected %s value. got %s.", expectedResult, tokenValResult))
	// System with count of 2 has value 0.9, and "SYSTEM" is in fluff dictionary so multiplied by 0.2, and "SYSTEM" in fhir endpoint tokens but not npi organization tokens so multiply by 0.3
	tokenValResult = fmt.Sprintf("%f", tokenVals["SYSTEM"])
	expectedResult = fmt.Sprintf("%f", systemValue*0.2*0.3)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("expected %s value. got %s.", expectedResult, tokenValResult))
	// Blah with count of 2 has value 0.9, and 2 < mean so multiplied by 2.5, and "BLAH" is in in fhir endpoint tokens but not npi organization tokens so multiply by 0.3
	tokenValResult = fmt.Sprintf("%f", tokenVals["BLAH"])
	expectedResult = fmt.Sprintf("%f", blahValue*2.5*0.3)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("expected %s value. got %s.", expectedResult, tokenValResult))
	// Seven with count of 1 has value 0.95, and 1 < mean so multiplied by 2.5, and "SEVEN" is in in npi organization tokens but not fhir endpoint tokens so multiply by 0.3
	tokenValResult = fmt.Sprintf("%f", tokenVals["SEVEN"])
	expectedResult = fmt.Sprintf("%f", sevenValue*2.5*0.3)
	th.Assert(t, tokenValResult == expectedResult, fmt.Sprintf("expected %s value. got %s.", expectedResult, tokenValResult))
}
