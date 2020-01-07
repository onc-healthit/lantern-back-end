package endpointmanager

import (
	"context"
	"time"
)

// FHIREndpoint represents a fielded FHIR API endpoint hosted by a
// HealthITProduct and populated by a ProviderOrganization.
// Information about the FHIR API endpoint is populated by the FHIR
// capability statement found at that endpoint as well as information
// discovered about the IP address of the endpoint.
type FHIREndpoint struct {
	ID                    int
	URL                   string
	FHIRVersion           string
	AuthorizationStandard string               // examples: OAuth 2.0, Basic, etc.
	Location              *Location            // location of the FHIR API endpoint's IP address from ipstack.com.
	CapabilityStatement   *CapabilityStatement // the JSON representation of the FHIR capability statement
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// FHIREndpointStore is the interface for interacting with the storage layer that holds
// FHIR endpoint objects.
type FHIREndpointStore interface {
	GetFHIREndpoint(context.Context, int) (*FHIREndpoint, error)
	GetFHIREndpointUsingURL(context.Context, string) (*FHIREndpoint, error)

	AddFHIREndpoint(context.Context, *FHIREndpoint) error
	UpdateFHIREndpoint(context.Context, *FHIREndpoint) error
	DeleteFHIREndpoint(context.Context, *FHIREndpoint) error

	Close()
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
	if e.FHIRVersion != e2.FHIRVersion {
		return false
	}
	if e.AuthorizationStandard != e2.AuthorizationStandard {
		return false
	}
	if !e.Location.Equal(e2.Location) {
		return false
	}
	if !e.CapabilityStatement.Equal(e2.CapabilityStatement) {
		return false
	}

	return true
}
