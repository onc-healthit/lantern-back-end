package chplendpointquerier

import (
	"os"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

func BundleQuerierParser(CHPLURL string, fileToWriteTo string) {

	var endpointEntryList EndpointList

	respBody, err := helpers.QueryEndpointList(CHPLURL)
	if err != nil {

		// Log the underlying Go error type
		log.Errorf("NETWORK ERROR: Failed to fetch URL=%s | Error=%v", CHPLURL, err)

		// If it is a timeout error
		if os.IsTimeout(err) {
			log.Errorf("TIMEOUT: URL=%s did not respond in time", CHPLURL)
		}

		// Detect DNS (“no such host”)
		if strings.Contains(err.Error(), "no such host") {
			log.Errorf("DNS ERROR: Domain could not be resolved for URL=%s", CHPLURL)
		}

		// Detect connection refused / unreachable
		if strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "connection reset") ||
			strings.Contains(err.Error(), "connection timed out") {
			log.Errorf("CONNECTION ERROR: Could not reach URL=%s", CHPLURL)
		}

		log.Info("Error for the URL: ", CHPLURL)
		log.Fatal(err)
	}

	// convert bundle data to lantern format
	endpointEntryList.Endpoints = BundleToLanternFormat(respBody, CHPLURL)

	err = WriteCHPLFile(endpointEntryList, fileToWriteTo)
	if err != nil {
		log.Info("Error for the URL: ", CHPLURL)
		log.Fatal(err)
	}
}
