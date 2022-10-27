package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointwebscraper"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
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

	err := config.SetupConfig()
	if err != nil {
		log.Fatalf("Error setting up config")
	}

	endpointwebscraper.EndpointListWebscraper(vendorURL, vendor, fileToWriteTo)

}
