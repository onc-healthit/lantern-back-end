package main

import (
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointwebscraper"
)

func main() {

	var vendor string
	var vendorURL string
	var fileToWriteTo string

	if len(os.Args) >= 1 {
		vendor = os.Args[1]
		vendorURL = os.Args[2]
		fileToWriteTo = os.Args[3]
	} else {
		log.Fatalf("ERROR: Missing command-line arguments")
	}

	endpointwebscraper.EndpointListWebscraper(vendorURL, vendor, fileToWriteTo)

}
