package endpointmanager

// FHIREndpointAvailability contains the number of times an endpoint
// returned a http response of 200 and number of times an endpoint
// has been queried overall. This info is used to calculate the
// availability of an endpoint.
type FHIREndpointAvailability struct {
	URL            string
	HTTP_200_COUNT int
	HTTP_ALL_COUNT int
}
