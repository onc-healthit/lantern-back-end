package chplendpointquerier

import (
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

// AthenaBundleQuerierParser is a custom bundle parser for Athena that filters out the generic URL
func AthenaBundleQuerierParser(CHPLURL string, fileToWriteTo string) {

	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(CHPLURL)
	if err != nil {
		log.Info("Error for the URL: ", CHPLURL)
		log.Fatal(err)
	}

	// convert bundle data to lantern format with Athena filtering
	endpointEntryList.Endpoints = AthenaBundleToLanternFormat(respBody, CHPLURL)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Info("Error for the URL: ", CHPLURL)
		log.Fatal(err)
	}
}

// AthenaBundleToLanternFormat converts bundle data to lantern format while filtering out generic Athena URL
func AthenaBundleToLanternFormat(bundle []byte, chplURL string) []LanternEntry {
	// First get all entries using the existing function
	allEntries := BundleToLanternFormat(bundle, chplURL)

	// Filter out the generic Athena URL
	var filteredEntries []LanternEntry
	const genericAthenaURL = "https://api.platform.athenahealth.com/fhir/r4"

	for _, entry := range allEntries {
		// Skip entries with the generic Athena URL (case-insensitive and trimmed comparison)
		if strings.TrimSpace(strings.ToLower(entry.URL)) != strings.ToLower(genericAthenaURL) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	log.Infof("Athena URL filtering: Original entries: %d, Filtered entries: %d, Removed: %d",
		len(allEntries), len(filteredEntries), len(allEntries)-len(filteredEntries))

	return filteredEntries
}
