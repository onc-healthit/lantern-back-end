package main

import (
	"os"
	log "github.com/sirupsen/logrus"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/chplpopulator"
)

func main() {

	var chplURL string
	var fileToWriteToCHPLList string
	fileToWriteToSoftwareInfo := "CHPLProductsInfo.json"

	if len(os.Args) >= 1 {
		chplURL = os.Args[1]
		fileToWriteToCHPLList = os.Args[2]
	} else {
		log.Fatalf("ERROR: Missing command-line arguments")
	}

	chplpopulator.FetchCHPLEndpointListProducts(chplURL, fileToWriteToCHPLList, fileToWriteToSoftwareInfo)

}
