package endpointmanager

import (
	"time"
)

// FHIREndpoint represents a fielded FHIR API endpoint hosted by a
// HealthITProduct and populated by a ProviderOrganization.
// Information about the FHIR API endpoint is populated by the FHIR
// capability statement found at that endpoint as well as information
// discovered about the IP address of the endpoint.
type FHIREndpoint struct {
	ID               int
	URL              string
	OrganizationName string
	ListSource       string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Equal checks each field of the two FHIREndpoints except for the database ID, CreatedAt and UpdatedAt fields to see if they are equal.
func (e *FHIREndpoint) Equal(e2 *FHIREndpoint) bool {
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
	if e.OrganizationName != e2.OrganizationName {
		return false
	}
	if e.ListSource != e2.ListSource {
		return false
	}

	return true
}
