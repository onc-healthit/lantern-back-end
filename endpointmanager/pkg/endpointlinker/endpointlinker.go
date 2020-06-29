package endpointlinker

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/gonum/stat"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/pkg/errors"
)

func NormalizeOrgName(orgName string) (string, error) {
	orgName = strings.ReplaceAll(orgName, "-", " ")
	orgName = strings.ReplaceAll(orgName, "/", " ")
	orgName = strings.ReplaceAll(orgName, ",", " ")
	// Regex for only letters
	reg, err := regexp.Compile(`[^a-zA-Z0-9\s]+`)
	if err != nil {
		return "", errors.Wrap(err, "error compiling regex for normalizing organization name")
	}
	characterStrippedName := reg.ReplaceAllString(orgName, "")
	return strings.ToUpper(characterStrippedName), nil
}

func calculateWeightedJaccardIndex(string1 string, string2 string, tokenVal map[string]float64) float64 {
	// https://www.statisticshowto.datasciencecentral.com/jaccard-index/
	// Tokens given weights based on their freqeuncy,
	// Find the weighted value of common tokens and divide it by the total weight of all the tokens
	string1Tokens := strings.Fields(string1)
	string2Tokens := strings.Fields(string2)
	// By taking the total tokens weight and subtracting by the intersection weight we allow for there to be
	// repeated tokens. ie: foo foo bar and foo bar would not be considered identical
	set1Map := make(map[string]int)
	intersectCount := 0.0
	denom := 0.0
	for _, name := range string1Tokens {
		denom = denom + tokenVal[name]
		if _, exists := set1Map[name]; !exists {
			set1Map[name] = 1
		} else {
			set1Map[name]++
		}
	}
	for _, name := range string2Tokens {
		denom = denom + tokenVal[name]
		if set1Map[name] > 0 {
			intersectCount = intersectCount + tokenVal[name]
			set1Map[name]--
		}
	}
	denom = denom - intersectCount

	if denom == 0 {
		denom = 1
	}
	return float64(intersectCount / denom)
}

// This function is available for making the matching algorithm easier to tune
func verbosePrint(message string, verbose bool) {
	if verbose {
		println(message)
	}
}

func getIdsOfMatchingNPIOrgs(npiOrgNames []*endpointmanager.NPIOrganization, normalizedEndpointName string, verbose bool, tokenVal map[string]float64, jaccardThreshold float64) ([]string, map[string]float64, error) {
	JACCARD_THRESHOLD := jaccardThreshold

	matches := []string{}
	confidenceMap := make(map[string]float64)

	verbosePrint(normalizedEndpointName+" Matched To:", verbose)
	for _, npiOrg := range npiOrgNames {
		consideredMatch := false
		confidence := 0.0
		jaccard1 := calculateWeightedJaccardIndex(normalizedEndpointName, npiOrg.NormalizedName, tokenVal)
		jaccard2 := calculateWeightedJaccardIndex(normalizedEndpointName, npiOrg.NormalizedSecondaryName, tokenVal)
		if jaccard1 == 1.0 {
			confidence = jaccard1
			consideredMatch = true
			verbosePrint("Exact Match Primary Name: "+normalizedEndpointName, verbose)
		} else if jaccard2 == 1.0 {
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
			// multiply confidence by .99 for all name matches to demonstrate that these matches are not as good as the id matches
			confidence = confidence * .99
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

func matchByName(endpoint *endpointmanager.FHIREndpoint, npiOrgNames []*endpointmanager.NPIOrganization, verbose bool, tokenVal map[string]float64, jaccardThreshold float64) ([]string, map[string]float64, error) {
	allMatches := make([]string, 0)
	allConfidences := make(map[string]float64)
	for _, name := range endpoint.OrganizationNames {
		normalizedEndpointName, err := NormalizeOrgName(name)
		if err != nil {
			return allMatches, allConfidences, errors.Wrap(err, "Error getting normalizing endpoint organizaton name")
		}
		matches, confidences, err := getIdsOfMatchingNPIOrgs(npiOrgNames, normalizedEndpointName, verbose, tokenVal, jaccardThreshold)
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

func countTokens(npiOrg []*endpointmanager.NPIOrganization, FHIREndpoints []*endpointmanager.FHIREndpoint) (map[string]int, map[string]int, map[string]int, string) {
	tokenCounterAll := make(map[string]int)
	tokenCounterNPI := make(map[string]int)
	tokenCounterEndpoints := make(map[string]int)
	firstKey := ""
	// Counts all the tokens in the primary and secondary normalized names for NPI organizations
	for _, organization := range npiOrg {
		orgNameTokens := strings.Fields(organization.NormalizedName)
		secondaryOrgNameTokens := strings.Fields(organization.NormalizedSecondaryName)
		tokenCounterAll, tokenCounterNPI, firstKey = counter(orgNameTokens, tokenCounterAll, tokenCounterNPI, firstKey)
		tokenCounterAll, tokenCounterNPI, firstKey = counter(secondaryOrgNameTokens, tokenCounterAll, tokenCounterNPI, firstKey)
	}

	// Counts all tokens in the FHIR endpoint names
	for _, endpoint := range FHIREndpoints {
		for _, name := range endpoint.OrganizationNames {
			endpointName, _ := NormalizeOrgName(name)
			endpointNameTokens := strings.Fields(endpointName)
			tokenCounterAll, tokenCounterEndpoints, firstKey = counter(endpointNameTokens, tokenCounterAll, tokenCounterEndpoints, firstKey)
		}
	}
	return tokenCounterAll, tokenCounterNPI, tokenCounterEndpoints, firstKey
}

func counter(tokenStrings []string, tokenCounterAll map[string]int, tokenCounterList map[string]int, firstKey string) (map[string]int, map[string]int, string) {
	for _, token := range tokenStrings {

		tokenCounterAll[token]++
		tokenCounterList[token]++

		if tokenCounterAll[token] >= tokenCounterAll[firstKey] {
			firstKey = token
		}
	}

	return tokenCounterAll, tokenCounterList, firstKey
}

func computeTokenValues(tokenCounts map[string]int, tokenCountsNPI map[string]int, tokenCountsEndpoints map[string]int, firstKey string, tokenMean int, tokenStandardDev int) map[string]float64 {
	tokenVal := make(map[string]float64)
	fluffDictionary := makeFluffDictionary()
	for key, value := range tokenCounts {
		tokenVal[key] = 1.0 - (float64(value) / float64(tokenCounts[firstKey]))

		// token count ranges that multiply token value to adjust weight of tokens according to their rarity
		if fluffDictionary[key] {
			tokenVal[key] *= 0.2
		} else if value < tokenMean {
			tokenVal[key] *= 2.5
		} else if value < tokenMean+(tokenStandardDev/3) {
			tokenVal[key] *= 1.6
		} else if value < tokenMean+(tokenStandardDev) {
			tokenVal[key] *= 1.3
		} else if value < tokenMean+(tokenStandardDev*3) {
			tokenVal[key] *= 1
		} else if value < tokenMean+(tokenStandardDev*6) {
			tokenVal[key] *= 0.8
		} else if value < tokenMean+(tokenStandardDev*9) {
			tokenVal[key] *= 0.6
		} else {
			tokenVal[key] *= 0.4
		}

		//Multiplies token value to adjust weight of tokens that appear in one list (NPI organization/FHIR endpoints) but not the other
		if (tokenCountsNPI[key] == 0 && tokenCountsEndpoints[key] != 0) || (tokenCountsNPI[key] != 0 && tokenCountsEndpoints[key] == 0) {
			tokenVal[key] *= 0.3
		}

	}

	return tokenVal
}

func makeFluffDictionary() map[string]bool {
	// The contents of this list were determined by comparing similar organization names from NPPES and fhir endpoints
	// Creates higher matching scores for organizations who differ slightly by the addition or removal of these words
	var fluffDictionary = make(map[string]bool)

	fluffDictionary["LLC"] = true
	fluffDictionary["EMS"] = true
	fluffDictionary["DR"] = true
	fluffDictionary["PA"] = true
	fluffDictionary["MD"] = true
	fluffDictionary["LTD"] = true
	fluffDictionary["PC"] = true
	fluffDictionary["DPM"] = true
	fluffDictionary["LLP"] = true
	fluffDictionary["AND"] = true
	fluffDictionary["OF"] = true
	fluffDictionary["IN"] = true
	fluffDictionary["THE"] = true
	fluffDictionary["MCC"] = true
	fluffDictionary["MMC"] = true
	fluffDictionary["TO"] = true
	fluffDictionary["PLC"] = true
	fluffDictionary["PLLC"] = true
	fluffDictionary["SYSTEM"] = true
	fluffDictionary["SERVICES"] = true
	fluffDictionary["DPMPC"] = true
	fluffDictionary["MDSC"] = true
	fluffDictionary["CORP"] = true
	fluffDictionary["HSHS"] = true
	fluffDictionary["ST"] = true
	fluffDictionary["CARE"] = true
	fluffDictionary["INC"] = true
	fluffDictionary["CLINIC"] = true
	fluffDictionary["GROUP"] = true
	fluffDictionary["CENTERS"] = true
	fluffDictionary["CENTER"] = true

	return fluffDictionary
}

func getTokenVals(npiOrg []*endpointmanager.NPIOrganization, FHIREndpoints []*endpointmanager.FHIREndpoint) map[string]float64 {
	tokenCounterAll, tokenCounterNPI, tokenCounterEndpoints, firstKey := countTokens(npiOrg, FHIREndpoints)
	var tokenCountsData []float64
	for _, count := range tokenCounterAll {
		tokenCountsData = append(tokenCountsData, float64(count))
	}
	tokenMean, tokenStandardDev := stat.MeanStdDev(tokenCountsData, nil)
	tokenVal := computeTokenValues(tokenCounterAll, tokenCounterNPI, tokenCounterEndpoints, firstKey, int(math.Round(tokenMean)), int(math.Round(tokenStandardDev)))
	return tokenVal
}

func LinkAllOrgsAndEndpoints(ctx context.Context, store *postgresql.Store, verbose bool) error {
	jaccardThreshold := .85
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
		nameMatches, nameConfidences, err := matchByName(endpoint, npiOrgNames, verbose, tokenVal, jaccardThreshold)
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
