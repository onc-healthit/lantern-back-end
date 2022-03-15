package main

import (
	"os"

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

	chplendpointquerier.queryCHPLEndpointList(chplURL, fileToWriteTo)
}
