package main

import (
	"os"

	querier "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/medicaidendpointquerier"

	log "github.com/sirupsen/logrus"
)

func main() {

	var fileToWriteTo string

	if len(os.Args) >= 1 {
		fileToWriteTo = os.Args[1]
	} else {
		log.Fatalf("ERROR: Missing command-line arguments")
	}

	querier.QueryMedicaidEndpointList(fileToWriteTo)
}
