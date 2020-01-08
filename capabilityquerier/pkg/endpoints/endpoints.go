package endpoints

import (
	"log"
	"os"

	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"
)

// GetEndpoints gets the endpoints from a resource file.
// TODO: this is temporary!!! These endpoints will be retrieved from a queue eventually.
func GetEndpoints() (*fetcher.ListOfEndpoints, error) {
	var endpointsFile string
	if len(os.Args) != 1 {
		endpointsFile = os.Args[1]
	} else {
		endpointsFile = "../../../endpointnetworkquerier/resources/EndpointSources.json"
	}
	var listOfEndpoints, err = fetcher.GetListOfEndpoints(endpointsFile)
	if err != nil {
		log.Fatal("Endpoint List Parsing Error: ", err.Error())
	}
	return &listOfEndpoints, nil
}
