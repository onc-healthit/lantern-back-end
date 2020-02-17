package endpointlinker

import (
	"strings"
	"log"
	"regexp"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

func NormalizeOrgName(orgName string) string{
	// Regex for only letters
	orgName = strings.ReplaceAll(orgName, "-", " ")
	reg, err := regexp.Compile(`[^a-zA-Z0-9\s]+`)
	if err != nil {
		log.Fatal(err)
	}
	characterStrippedName := reg.ReplaceAllString(orgName, "")
	return strings.ToUpper(characterStrippedName)
}

func intersectionCount(set1 []string, set2 []string) int{
	set1Map := make(map[string]int)
	intersectionCount := 0
	for _, name := range set1 {
		if _, exists := set1Map[name]; !exists {
			set1Map[name] = 1
		}else{
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

func CalculateJaccardIndex(string1 string, string2 string) float64 {
	// Find the number of common tokens and divide it by the total number of unique tokens
	string1Tokens := strings.Fields(string1)
	string2Tokens := strings.Fields(string2)
	intersectionCount := intersectionCount(string1Tokens, string2Tokens)
	string1UniqueTokens := len(string1Tokens)
	string2UniqueTokens := len(string2Tokens)
	denom := float64(string1UniqueTokens + string2UniqueTokens - intersectionCount)
	if denom == 0 {
		denom = 1
	}
	return float64(intersectionCount)/denom
}

func verbosePrint(message string, verbose bool) {
	if verbose == true {
		println(message)
	}
}

func GetIdsOfMatchingNPIOrgs(npiOrgNames []endpointmanager.NPIOrganization, normalizedEndpointName string, verbose bool ) ([]int, error){
	JACARD_THRESHOLD := .75

	matches := []int{}
	verbosePrint(normalizedEndpointName + " Matched To:", verbose)
	for _, npiOrg := range npiOrgNames {
		consideredMatch := false
		jacard1 := CalculateJaccardIndex(normalizedEndpointName, npiOrg.NormalizedName)
		jacard2 := CalculateJaccardIndex(normalizedEndpointName, npiOrg.NormalizedSecondaryName)
		if (jacard1 == 1){
			consideredMatch = true
			matches = append(matches, npiOrg.ID)
		}else if (jacard1 >= JACARD_THRESHOLD) {
			consideredMatch = true
			verbosePrint(normalizedEndpointName + "=>" + npiOrg.NormalizedName, verbose)
		}
		if (jacard2 == 1){
			consideredMatch = true
		}else if (jacard2 >= JACARD_THRESHOLD) {
			consideredMatch = true
			verbosePrint(normalizedEndpointName + "=>" + npiOrg.NormalizedSecondaryName, verbose)
		}
		if consideredMatch == true {
			matches = append(matches, npiOrg.ID)
		}
	}
	return matches, nil
}