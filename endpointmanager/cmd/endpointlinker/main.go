package main

import (
	"strconv"
	"context"
	"os"
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

func verbosePrint(message string, verbose bool) {
	if verbose == true {
		println(message)
	}
}

func main() {
	var verbose = false
	if len(os.Args) > 1 && os.Args[1] == "--verbose" {
		verbose = true
	}

	err := config.SetupConfig()
	failOnError("Error setting up confit", err)
	ctx := context.Background()

	store, err := postgresql.NewStore(viper.GetString("dbhost"), viper.GetInt("dbport"), viper.GetString("dbuser"), viper.GetString("dbpassword"), viper.GetString("dbname"), viper.GetString("dbsslmode"))
	failOnError("Error creating store", err)

	fhirEndpointOrgNames, err := store.GetAllOrgNames(ctx)
	failOnError("Error getting endpoint org names", err)

	matchCount := 0
	unmatchable := []string{}
	// Iterate through fhir endpoints
	for _, endpoint := range fhirEndpointOrgNames {
		normalizedEndpointName := endpointlinker.NormalizeOrgName(endpoint.OrganizationName)
		matches := []int{}
		matches, err = endpointlinker.GetIdsOfMatchingNPIOrgs(store, ctx, normalizedEndpointName, verbose)
		if (len(matches) > 0){
			matchCount += 1
			// Iterate over matches and add to linking table
			for _, match := range matches {
				store.LinkOrganizationToEndpoint(ctx, match, endpoint.ID)
			}
		}else{
			unmatchable = append(unmatchable, endpoint.OrganizationName )
		}

	}

	verbosePrint("Match Total: " + strconv.Itoa(matchCount) + "/" + strconv.Itoa(len(fhirEndpointOrgNames)), verbose)

	verbosePrint("UNMATCHABLE ENDPOINT ORG NAMES", verbose)
	for _, name := range unmatchable {
		verbosePrint(name, verbose)
	}

}