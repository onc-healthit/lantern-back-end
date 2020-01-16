package endpoints

import (
	"fmt"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_GetEndpoints(t *testing.T) {
	endpointsLocation := "../../../networkstatsquerier/resources/EndpointSources.json"
	expectedCount := 354

	// basic test
	eps, err := GetEndpoints(endpointsLocation)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(eps.Entries) == expectedCount, fmt.Sprintf("expected %d and received %d endpoint entries", expectedCount, len(eps.Entries)))

	// can't test command line arguments or lack of because the test itself passes in a command line argument
}
