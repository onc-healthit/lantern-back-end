package fetcher

import "testing"

func Test_GetListOfEndpoints(t *testing.T) {
	var EXPECTED_ENDPOINTS = 354
	var endpoints, _ = GetListOfEndpoints("../../resources/EndpointSources.json")
	var endpointsCount = len(endpoints.Entries)
	if endpointsCount != EXPECTED_ENDPOINTS {
		t.Errorf("Number of endpoints read from resource file incorrect, got: %d, want: %d.", endpointsCount, EXPECTED_ENDPOINTS)
	}
}
