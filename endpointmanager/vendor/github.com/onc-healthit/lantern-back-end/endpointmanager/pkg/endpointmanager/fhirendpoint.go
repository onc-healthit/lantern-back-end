package endpointmanager

import (
	"context"
	"sort"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
)

// FHIREndpoint represents a fielded FHIR API endpoint hosted by a
// HealthITProduct and populated by a ProviderOrganization.
// Information about the FHIR API endpoint is populated by the FHIR
// capability statement found at that endpoint as well as information
// discovered about the IP address of the endpoint.
type FHIREndpoint struct {
	ID                    int
	URL                   string
	TLSVersion            string
	MIMETypes             []string
	HTTPResponse          int
	Errors                string
	OrganizationName      string
	FHIRVersion           string
	AuthorizationStandard string // examples: OAuth 2.0, Basic, etc.
	Vendor                string
	Location              *Location                            // location of the FHIR API endpoint's IP address from ipstack.com.
	CapabilityStatement   capabilityparser.CapabilityStatement // the JSON representation of the FHIR capability statement
	Validation            map[string]interface{}
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
	if e.TLSVersion != e2.TLSVersion {
		return false
	}

	// check MIMETypes equal
	if len(e.MIMETypes) != len(e2.MIMETypes) {
		return false
	}
	// don't care about order
	a := make([]string, len(e.MIMETypes))
	b := make([]string, len(e2.MIMETypes))
	sort.Strings(a)
	sort.Strings(b)
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	if e.HTTPResponse != e2.HTTPResponse {
		return false
	}
	if e.Errors != e2.Errors {
		return false
	}
	if e.OrganizationName != e2.OrganizationName {
		return false
	}
	if e.FHIRVersion != e2.FHIRVersion {
		return false
	}
	if e.AuthorizationStandard != e2.AuthorizationStandard {
		return false
	}
	if e.Vendor != e2.Vendor {
		return false
	}
	if !e.Location.Equal(e2.Location) {
		return false
	}
	// because CapabilityStatement is an interface, we need to confirm it's not nil before using the Equal
	// method.
	if e.CapabilityStatement != nil && !e.CapabilityStatement.Equal(e2.CapabilityStatement) {
		return false
	}
	if e.CapabilityStatement == nil && e2.CapabilityStatement != nil {
		return false
	}

	return true
}
