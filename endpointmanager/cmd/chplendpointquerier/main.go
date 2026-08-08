package main

import (
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/chplendpointquerier"
	log "github.com/sirupsen/logrus"
)

func main() {

	var chplURL string
	var fileToWriteTo string

	if len(os.Args) >= 1 {
		chplURL = os.Args[1]
		fileToWriteTo = os.Args[2]
	} else {
		log.Fatalf("ERROR: Missing command-line arguments")
	}

	errorFilePath := os.Getenv("CHPL_ERROR_FILE")
	if errorFilePath == "" {
		errorFilePath = "/etc/lantern/logs/chpl_query_errors.csv"
	}

	chplendpointquerier.QueryCHPLEndpointList(chplURL, fileToWriteTo, errorFilePath)
}
