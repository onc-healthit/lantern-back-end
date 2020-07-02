package endpoints

import (
	"errors"
	"os"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fetcher"
)

// GetEndpoints gets the endpoints from a resource file.
// TODO: this is temporary!!! These endpoints will be retrieved from a queue eventually.
func GetEndpoints(endpointsFile string) (*fetcher.ListOfEndpoints, error) {
	if len(endpointsFile) == 0 {
		if len(os.Args) != 1 {
			endpointsFile = os.Args[1]
		} else {
			return nil, errors.New("no endpoints file given")
		}
	}
	var listOfEndpoints, err = fetcher.GetEndpointsFromFilepath(endpointsFile, "CareEvolution")
	if err != nil {
		return nil, err
	}
	return &listOfEndpoints, nil
}
