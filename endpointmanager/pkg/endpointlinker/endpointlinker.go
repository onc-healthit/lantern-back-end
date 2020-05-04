package endpointlinker

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/pkg/errors"
)

func NormalizeOrgName(orgName string) string {
	// Regex for only letters
	orgName = strings.ReplaceAll(orgName, "-", " ")
	reg, err := regexp.Compile(`[^a-zA-Z0-9\s]+`)
	if err != nil {
		log.Fatal(err)
	}
	characterStrippedName := reg.ReplaceAllString(orgName, "")
	return strings.ToUpper(characterStrippedName)
}

func intersectionCount(set1 []string, set2 []string) int {
	set1Map := make(map[string]int)
	intersectionCount := 0
	for _, name := range set1 {
		if _, exists := set1Map[name]; !exists {
			set1Map[name] = 1
		} else {
			set1Map[name] += 1
		}
	}
	for _, name := range set2 {
		if set1Map[name] > 0 {
			intersectionCount += 1
			set1Map[name] -= 1
		}
	}
	return intersectionCount
}

func calculateJaccardIndex(string1 string, string2 string) float64 {
	// https://www.statisticshowto.datasciencecentral.com/jaccard-index/
	// Find the number of common tokens and divide it by the total number of unique tokens
	string1Tokens := strings.Fields(string1)
	string2Tokens := strings.Fields(string2)
	intersectionCount := intersectionCount(string1Tokens, string2Tokens)
	// By taking the total tokens count and subtracting by the intersection we allow for there to be
	// repeated tokens. ie: foo foo bar and foo bar would not be considered identical
	string1TokensCount := len(string1Tokens)
	string2TokensCount := len(string2Tokens)
	denom := float64(string1TokensCount + string2TokensCount - intersectionCount)
	if denom == 0 {
		denom = 1
	}
	return float64(intersectionCount) / denom
}

// This function is available for making the matching algorithm easier to tune
func verbosePrint(message string, verbose bool) {
	if verbose {
		println(message)
	}
}

func getIdsOfMatchingNPIOrgs(npiOrgNames []endpointmanager.NPIOrganization, normalizedEndpointName string, verbose bool) ([]int, map[int]float64, error) {
	JACCARD_THRESHOLD := .75

	matches := []int{}
	confidenceMap := make(map[int]float64)

	verbosePrint(normalizedEndpointName+" Matched To:", verbose)
	for _, npiOrg := range npiOrgNames {
		consideredMatch := false
		confidence := 0.0
		jaccard1 := calculateJaccardIndex(normalizedEndpointName, npiOrg.NormalizedName)
		jaccard2 := calculateJaccardIndex(normalizedEndpointName, npiOrg.NormalizedSecondaryName)
		if jaccard1 == 1 {
			confidence = 1
			consideredMatch = true
			verbosePrint("Exact Match Primary Name: "+normalizedEndpointName, verbose)
		} else if jaccard2 == 1 {
			confidence = 1
			consideredMatch = true
			verbosePrint("Exact Match Secondary Name: "+normalizedEndpointName, verbose)
		} else if jaccard1 >= JACCARD_THRESHOLD && jaccard1 > jaccard2 {
			confidence = jaccard1
			consideredMatch = true
			verbosePrint(normalizedEndpointName+"=>"+npiOrg.NormalizedName+" Match Score: "+fmt.Sprintf("%f", jaccard1), verbose)
		} else if jaccard2 >= JACCARD_THRESHOLD {
			consideredMatch = true
			confidence = jaccard2
			verbosePrint(normalizedEndpointName+"=>"+npiOrg.NormalizedSecondaryName+" Match Score: "+fmt.Sprintf("%f", jaccard2), verbose)
		}
		if consideredMatch {
			confidenceMap[npiOrg.ID] = confidence
			matches = append(matches, npiOrg.ID)
		}
	}
	return matches, confidenceMap, nil
}

// updates allMatches and allConfidences. allConfidences gets updated within the function. Because allMatches is a
// slice and we use 'append', a new slice might be allocated, so we need to return allMatches in case a new slice is
// created.
func mergeMatches(allMatches []int, allConfidences map[int]float64, matches []int, confidences map[int]float64) []int {
	for _, match := range matches {
		if !helpers.IntArrayContains(allMatches, match) {
			allMatches = append(allMatches, match)
			allConfidences[match] = confidences[match]
		} else {
			if confidences[match] > allConfidences[match] {
				allConfidences[match] = confidences[match]
			}
		}
	}

	return allMatches
}

func LinkAllOrgsAndEndpoints(ctx context.Context, store *postgresql.Store, verbose bool) error {
	fhirEndpointOrgNames, err := store.GetAllFHIREndpointOrgNames(ctx)
	if err != nil {
		return errors.Wrap(err, "Error getting endpoint org names")
	}

	npiOrgNames, err := store.GetAllNPIOrganizationNormalizedNames(ctx)
	if err != nil {
		return errors.Wrap(err, "Error getting normalized org names")
	}

	matchCount := 0
	unmatchable := []string{}
	// Iterate through fhir endpoints
	for _, endpoint := range fhirEndpointOrgNames {
		allMatches := make([]int, 0)
		allConfidences := make(map[int]float64)

		for _, name := range endpoint.OrganizationNames {
			normalizedEndpointName := NormalizeOrgName(name)
			matches, confidences, err := getIdsOfMatchingNPIOrgs(npiOrgNames, normalizedEndpointName, verbose)
			if err != nil {
				return errors.Wrap(err, "Error getting matching NPI org IDs")
			}

			allMatches = mergeMatches(allMatches, allConfidences, matches, confidences)
		}

		if len(allMatches) > 0 {
			matchCount++
			// Iterate over matches and add to linking table
			for _, match := range allMatches {
				err = store.LinkNPIOrganizationToFHIREndpoint(ctx, match, endpoint.ID, allConfidences[match])
				if err != nil {
					return errors.Wrap(err, "Error linking org to FHIR endpoint")
				}
			}
		} else {
			if verbose {
				unmatchable = append(unmatchable, endpoint.URL)
			}
		}
	}

	verbosePrint("Match Total: "+strconv.Itoa(matchCount)+"/"+strconv.Itoa(len(fhirEndpointOrgNames)), verbose)

	verbosePrint("UNMATCHABLE ENDPOINT ORG NAMES", verbose)
	if verbose {
		for _, name := range unmatchable {
			verbosePrint(name, verbose)
		}
	}

	return nil
}
