package main

import (
	"strconv"
	"context"
	"log"
	"github.com/spf13/viper"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointlinker"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

func failOnError(errString string, err error) {
	if err != nil {
		log.Fatalf("%s %s", errString, err)
	}
}

func main() {
	JACARD_THRESHOLD := .75

	err := config.SetupConfig()
	failOnError("Error setting up confit", err)
	ctx := context.Background()

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError("Error creating store", err)

	fhirEndpointOrgNames, err := store.GetAllOrgNames(ctx)
	failOnError("Error getting endpoint org names", err)

	npiOrgNames, err := store.GetAllNormalizedOrgNames(ctx)
	failOnError("Error getting npi org names", err)

	exactMatchCount := 0
	nonExactMatchCount := 0
	unmatchable := []string{}
	// Iterate through fhir endpoints
	for _, endpoint := range fhirEndpointOrgNames {
		normalizedEndpointName := endpointlinker.NormalizeOrgName(endpoint.OrganizationName)
		exactPrimaryNameMatches := 0
		exactSecondaryNameMatches := 0
		nonExactPrimaryNameMatches := 0
		nonExactSecondaryNameMatches := 0
		println(normalizedEndpointName + " Matched To:")
		for _, npiOrg := range npiOrgNames {
			jacard1 := endpointlinker.CalculateJaccardIndex(normalizedEndpointName, npiOrg.NormalizedName)
			jacard2 := endpointlinker.CalculateJaccardIndex(normalizedEndpointName, npiOrg.NormalizedSecondaryName)
			if (jacard1 == 1){
				exactPrimaryNameMatches += 1
			}else if (jacard1 >= JACARD_THRESHOLD) {
				nonExactPrimaryNameMatches += 1
				println(normalizedEndpointName + "=>" + npiOrg.NormalizedName )
			}
			if (jacard2 == 1){
				exactSecondaryNameMatches += 1
			}else if (jacard2 >= JACARD_THRESHOLD) {
				nonExactSecondaryNameMatches += 1
				println(normalizedEndpointName + "=>" + npiOrg.NormalizedSecondaryName)
			}
		}
		if (exactPrimaryNameMatches > 0 || exactSecondaryNameMatches > 0 ){
			exactMatchCount += 1
		} else if (nonExactPrimaryNameMatches > 0 || nonExactSecondaryNameMatches > 0 ){
			nonExactMatchCount += 1
		}else{
			unmatchable = append(unmatchable, endpoint.OrganizationName )
		}
		println("NPI Orgs With Exact Primary Name: " + strconv.Itoa(exactPrimaryNameMatches))
		println("NPI Orgs With Non-Exact Primary Name: " + strconv.Itoa(nonExactPrimaryNameMatches))
		println("NPI Orgs With Exact Secondary Name: " + strconv.Itoa(exactSecondaryNameMatches))
		println("NPI Orgs With Non-Exact Secondary Name: " + strconv.Itoa(nonExactSecondaryNameMatches))
		println("===============================================")

	}

	println("Able To Find Exact Matches For: " + strconv.Itoa(exactMatchCount) + "/" + strconv.Itoa(len(fhirEndpointOrgNames)))
	println("Only Non-Exact Matches For: " + strconv.Itoa(nonExactMatchCount) + "/" + strconv.Itoa(len(fhirEndpointOrgNames)))
	println("Match Total: " + strconv.Itoa(nonExactMatchCount + exactMatchCount) + "/" + strconv.Itoa(len(fhirEndpointOrgNames)))

	println("UNMATCHABLE ENDPOINT ORG NAMES")
	for _, name := range unmatchable {
		println(name)
	}

}