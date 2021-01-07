package endpointmanager

import (
	"time"

	"github.com/google/go-cmp/cmp"
)

// FHIREndpointMetadata represents a fielded FHIR API endpoint hosted by a
// HealthITProduct and populated by a ProviderOrganization.
// Information about the FHIR API endpoint is populated by the FHIR
// capability statement found at that endpoint.
type FHIREndpointMetadata struct {
	ID                int
	URL               string
	HTTPResponse      int
	Errors            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	SMARTHTTPResponse int
	ResponseTime      float64
	Availability      float64
}

// Equal checks each field of the two FHIREndpointInfos except for the database ID, CreatedAt and UpdatedAt fields to see if they are equal.
func (e *FHIREndpointMetadata) Equal(e2 *FHIREndpointMetadata) bool {
	if e == nil && e2 == nil {
		return true
	} else if e == nil {
		return false
	} else if e2 == nil {
		return false
	}

	if e.URL != e2.URL {
		return false
	}
	if e.HTTPResponse != e2.HTTPResponse {
		return false
	}
	if e.Availability != e2.Availability {
		return false
	}
	if e.Errors != e2.Errors {
		return false
	}
	if e.SMARTHTTPResponse != e2.SMARTHTTPResponse {
		return false
	}
	if !cmp.Equal(e.ResponseTime, e2.ResponseTime) {
		return false
	}

	return true
}
