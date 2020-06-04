package endpointlinker

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/pkg/errors"
)

func NormalizeOrgName(orgName string) (string, error) {
	// Regex for only letters
	orgName = strings.ReplaceAll(orgName, "-", " ")
	orgName = strings.ReplaceAll(orgName, "/", " ")
	reg, err := regexp.Compile(`[^a-zA-Z0-9\s]+`)
	if err != nil {
		return "", errors.Wrap(err, "error compiling regex for normalizing organization name")
	}
	characterStrippedName := reg.ReplaceAllString(orgName, "")
	return strings.ToUpper(characterStrippedName), nil
}

func intersectionCount(set1 []string, set2 []string, tokenVal map[string]float64) (float64, float64) {
	set1Map := make(map[string]int)
	intersectCount := 0.0
	denom := 0.0
	for _, name := range set1 {
		if _, exists := set1Map[name]; !exists {
			set1Map[name] = 1
			denom = denom + tokenVal[name]
		} else {
			set1Map[name] += 1
			denom = denom + tokenVal[name]
		}
	}
	for _, name := range set2 {
		denom = denom + tokenVal[name]
		if set1Map[name] > 0 {
			intersectCount = intersectCount + tokenVal[name]
			set1Map[name] -= 1
		}
	}
	denom = denom - intersectCount
	return intersectCount, denom
}

func calculateJaccardIndex(string1 string, string2 string, tokenVal map[string]float64) float64 {
	// https://www.statisticshowto.datasciencecentral.com/jaccard-index/
	// Find the number of common tokens and divide it by the total number of unique tokens
	string1Tokens := strings.Fields(string1)
	string2Tokens := strings.Fields(string2)
	intersectionCount, denom := intersectionCount(string1Tokens, string2Tokens, tokenVal)
	// By taking the total tokens count and subtracting by the intersection we allow for there to be
	// repeated tokens. ie: foo foo bar and foo bar would not be considered identical

	if denom == 0 {
		denom = 1
	}
	return float64(intersectionCount / denom)
}

// This function is available for making the matching algorithm easier to tune
func verbosePrint(message string, verbose bool) {
	if verbose {
		println(message)
	}
}

func getIdsOfMatchingNPIOrgs(npiOrgNames []*endpointmanager.NPIOrganization, normalizedEndpointName string, verbose bool, tokenVal map[string]float64, jaccard_threshold float64) ([]string, map[string]float64, error) {
	JACCARD_THRESHOLD := jaccard_threshold

	matches := []string{}
	confidenceMap := make(map[string]float64)

	verbosePrint(normalizedEndpointName+" Matched To:", verbose)
	for _, npiOrg := range npiOrgNames {
		consideredMatch := false
		confidence := 0.0
		jaccard1 := calculateJaccardIndex(normalizedEndpointName, npiOrg.NormalizedName, tokenVal)
		jaccard2 := calculateJaccardIndex(normalizedEndpointName, npiOrg.NormalizedSecondaryName, tokenVal)
		if jaccard1 >= .99 {
			confidence = jaccard1
			consideredMatch = true
			verbosePrint("Exact Match Primary Name: "+normalizedEndpointName, verbose)
		} else if jaccard2 >= .99 {
			confidence = jaccard2
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
			// multiply confidence by .9 for all name matches to demonstrate that these matches are not as good as the id matches
			confidence = confidence * .9
			confidenceMap[npiOrg.NPI_ID] = confidence
			matches = append(matches, npiOrg.NPI_ID)
		}
	}
	return matches, confidenceMap, nil
}

// updates allMatches and allConfidences. Because allMatches is a slice and we use 'append', a new slice might be allocated,
// so we need to return allMatches in case a new slice is created. Also return allConfidences to make the function work more
// intuitively.
func mergeMatches(allMatches []string, allConfidences map[string]float64, matches []string, confidences map[string]float64) ([]string, map[string]float64) {
	for _, match := range matches {
		if !helpers.StringArrayContains(allMatches, match) {
			allMatches = append(allMatches, match)
			allConfidences[match] = confidences[match]
		} else {
			if confidences[match] > allConfidences[match] {
				allConfidences[match] = confidences[match]
			}
		}
	}

	return allMatches, allConfidences
}

func matchByID(ctx context.Context, endpoint *endpointmanager.FHIREndpoint, store *postgresql.Store, verbose bool) ([]string, map[string]float64, error) {
	matches := make([]string, 0)
	confidences := make(map[string]float64)
	for _, npiID := range endpoint.NPIIDs {
		npiOrg, err := store.GetNPIOrganizationByNPIID(ctx, npiID)
		if err == sql.ErrNoRows {
			// do nothing
		} else if err != nil {
			return matches, confidences, errors.Wrap(err, "error retrieving referenced NPI organization")
		} else {
			matches = append(matches, npiOrg.NPI_ID)
			confidences[npiOrg.NPI_ID] = 1
		}
	}
	return matches, confidences, nil
}

func matchByName(endpoint *endpointmanager.FHIREndpoint, npiOrgNames []*endpointmanager.NPIOrganization, verbose bool, tokenVal map[string]float64, jaccard_threshold float64) ([]string, map[string]float64, error) {
	allMatches := make([]string, 0)
	allConfidences := make(map[string]float64)
	for _, name := range endpoint.OrganizationNames {
		normalizedEndpointName, err := NormalizeOrgName(name)
		if err != nil {
			return allMatches, allConfidences, errors.Wrap(err, "Error getting normalizing endpoint organizaton name")
		}
		matches, confidences, err := getIdsOfMatchingNPIOrgs(npiOrgNames, normalizedEndpointName, verbose, tokenVal, jaccard_threshold)
		if err != nil {
			return allMatches, allConfidences, errors.Wrap(err, "Error getting matching NPI org IDs")
		}

		allMatches, allConfidences = mergeMatches(allMatches, allConfidences, matches, confidences)
	}
	return allMatches, allConfidences, nil
}

func addMatch(ctx context.Context, store *postgresql.Store, orgID string, endpoint *endpointmanager.FHIREndpoint, confidence float64) error {
	_, _, storedConfidence, err := store.GetNPIOrganizationFHIREndpointLink(ctx, orgID, endpoint.URL)
	if err == sql.ErrNoRows {
		err = store.LinkNPIOrganizationToFHIREndpoint(ctx, orgID, endpoint.URL, confidence)
		if err != nil {
			return errors.Wrap(err, "Error linking org to FHIR endpoint")
		}
	} else if err != nil {
		return err
	} else {
		if storedConfidence < confidence {
			err = store.UpdateNPIOrganizationFHIREndpointLink(ctx, orgID, endpoint.URL, confidence)
			if err != nil {
				return errors.Wrap(err, "Error update org to FHIR endpoint link")
			}
		}
	}
	return nil
}

func getTokenVals(npiOrg []*endpointmanager.NPIOrganization, FHIREndpoints []*endpointmanager.FHIREndpoint) map[string]float64 {
	tokenCounterAll := make(map[string]int)
	tokenCounterNPI := make(map[string]int)
	tokenCounterEndpoints := make(map[string]int)
	firstKey := ""
	var totalTokens int
	var totalUniqueTokens int
	for _, organization := range npiOrg {
		orgNameTokens := strings.Fields(organization.NormalizedName)
		secondaryOrgNameTokens := strings.Fields(organization.NormalizedSecondaryName)
		for _, orgToken := range orgNameTokens {
			if _, contains := tokenCounterAll[orgToken]; !contains {
				totalUniqueTokens += 1
			}
			tokenCounterAll[orgToken] += 1
			tokenCounterNPI[orgToken] += 1
			totalTokens++

			if tokenCounterAll[orgToken] >= tokenCounterAll[firstKey] {
				firstKey = orgToken
			}
		}
		for _, orgSecondaryToken := range secondaryOrgNameTokens {
			if _, contains := tokenCounterAll[orgSecondaryToken]; !contains {
				totalUniqueTokens += 1
			}
			tokenCounterAll[orgSecondaryToken] += 1
			tokenCounterNPI[orgSecondaryToken] += 1
			totalTokens++

			if tokenCounterAll[orgSecondaryToken] >= tokenCounterAll[firstKey] {
				firstKey = orgSecondaryToken
			}
		}
	}
	for _, endpoint := range FHIREndpoints {
		for _, name := range endpoint.OrganizationNames {
			endpointName, _ := NormalizeOrgName(name)
			endpointNameTokens := strings.Fields(endpointName)
			for _, endpointToken := range endpointNameTokens {
				if _, contains := tokenCounterAll[endpointToken]; !contains {
					totalUniqueTokens += 1
				}
				tokenCounterAll[endpointToken] += 1
				tokenCounterEndpoints[endpointToken] += 1
				totalTokens++

				if tokenCounterAll[endpointToken] >= tokenCounterAll[firstKey] {
					firstKey = endpointToken
				}
			}
		}
	}

	tokenMean := int(math.Round(float64(totalTokens) / float64(totalUniqueTokens)))
	tokenStandardDev := calculateStandardDev(tokenCounterAll, tokenMean, totalUniqueTokens)
	tokenVal := computeTokenValues(tokenCounterAll, tokenCounterNPI, tokenCounterEndpoints, firstKey, tokenMean, tokenStandardDev)
	return tokenVal
}

func calculateStandardDev(tokenCounterAll map[string]int, tokenMean int, totalUniqueTokens int) int {
	squaredDiffSum := 0.0
	for _, value := range tokenCounterAll {
		squaredDiffSum += math.Pow(float64(value-tokenMean), 2.0)
	}

	sqrtStandardDev := math.Sqrt(squaredDiffSum / float64(totalUniqueTokens))
	return int(math.Round(sqrtStandardDev))
}

func computeTokenValues(tokenCounts map[string]int, tokenCountsNPI map[string]int, tokenCountsEndpoints map[string]int, firstKey string, tokenMean int, tokenStandardDev int) map[string]float64 {
	tokenVal := make(map[string]float64)
	fluffDictionary := makeFluffDictionary()
	for key, value := range tokenCounts {
		tokenVal[key] = 1.0 - (float64(value) / float64(tokenCounts[firstKey]))

		if fluffDictionary[key] {
			tokenVal[key] *= 0.2
		} else if value < tokenMean {
			tokenVal[key] *= 2.5
		} else if value < tokenMean+(tokenStandardDev/3) {
			tokenVal[key] *= 1.6
		} else if value < tokenMean+(tokenStandardDev) {
			tokenVal[key] *= 1.3
		} else if value < tokenMean+(tokenStandardDev*3) {
			tokenVal[key] *= 1.0
		} else if value < tokenMean+(tokenStandardDev*6) {
			tokenVal[key] *= 0.8
		} else if value < tokenMean+(tokenStandardDev*9) {
			tokenVal[key] *= 0.6
		} else {
			tokenVal[key] *= 0.4
		}

		if tokenCountsNPI[key] == 0 && tokenCountsEndpoints[key] != 0 {
			tokenVal[key] *= 0.1
			continue
		} else if tokenCountsNPI[key] != 0 && tokenCountsEndpoints[key] == 0 {
			tokenVal[key] *= 0.3
			continue
		}

	}

	return tokenVal
}

func makeFluffDictionary() map[string]bool {
	var fluffDictionary = make(map[string]bool)

	fluffDictionary["LLC"] = true
	fluffDictionary["EMS"] = true
	fluffDictionary["DR"] = true
	fluffDictionary["PA"] = true
	fluffDictionary["MD"] = true
	fluffDictionary["LLC"] = true
	fluffDictionary["LTD"] = true
	fluffDictionary["PC"] = true
	fluffDictionary["DPM"] = true
	fluffDictionary["LLP"] = true
	fluffDictionary["AND"] = true
	fluffDictionary["OF"] = true
	fluffDictionary["IN"] = true
	fluffDictionary["THE"] = true
	fluffDictionary["LLP"] = true
	fluffDictionary["MCC"] = true
	fluffDictionary["MMC"] = true
	fluffDictionary["TO"] = true
	fluffDictionary["PLC"] = true
	fluffDictionary["PLLC"] = true
	fluffDictionary["SYSTEM"] = true
	fluffDictionary["SERVICES"] = true
	fluffDictionary["REGIONAL"] = true
	fluffDictionary["DPMPC"] = true
	fluffDictionary["MDSC"] = true

	return fluffDictionary
}

func LinkAllOrgsAndEndpoints(ctx context.Context, store *postgresql.Store, verbose bool) error {
	jaccard_threshold := .85
	fhirEndpoints, err := store.GetAllFHIREndpoints(ctx)
	if err != nil {
		return errors.Wrap(err, "Error getting endpoint org names")
	}

	npiOrgNames, err := store.GetAllNPIOrganizationNormalizedNames(ctx)
	if err != nil {
		return errors.Wrap(err, "Error getting normalized org names")
	}

	tokenVal := getTokenVals(npiOrgNames, fhirEndpoints)

	matchCount := 0
	unmatchable := []string{}
	// Iterate through fhir endpoints
	for _, endpoint := range fhirEndpoints {
		allMatches := make([]string, 0)
		allConfidences := make(map[string]float64)

		idMatches, idConfidences, err := matchByID(ctx, endpoint, store, verbose)
		if err != nil {
			return errors.Wrap(err, "error matching endpoint to NPI organization by ID")
		}
		nameMatches, nameConfidences, err := matchByName(endpoint, npiOrgNames, verbose, tokenVal, jaccard_threshold)
		if err != nil {
			return errors.Wrap(err, "error matching endpoint to NPI organization by name")
		}
		allMatches, allConfidences = mergeMatches(allMatches, allConfidences, idMatches, idConfidences)
		allMatches, allConfidences = mergeMatches(allMatches, allConfidences, nameMatches, nameConfidences)

		if len(allMatches) > 0 {
			matchCount++
			// Iterate over matches and add to linking table
			for _, match := range allMatches {
				err = addMatch(ctx, store, match, endpoint, allConfidences[match])
				if err != nil {
					return errors.Wrap(err, "Error linking org to FHIR endpoint")
				}
			}
		} else {
			if verbose {
				unmatchable = append(unmatchable, endpoint.OrganizationNames...)
			}
		}
	}

	verbosePrint("Match Total: "+strconv.Itoa(matchCount)+"/"+strconv.Itoa(len(fhirEndpoints)), verbose)

	verbosePrint("UNMATCHABLE ENDPOINT ORG NAMES", verbose)
	if verbose {
		for _, name := range unmatchable {
			verbosePrint(name, verbose)
		}
	}

	return nil
}
