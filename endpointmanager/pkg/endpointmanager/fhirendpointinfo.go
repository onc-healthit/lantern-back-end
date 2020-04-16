package endpointmanager

import (
	"database/sql"
	"sort"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
)

// FHIREndpoint represents a fielded FHIR API endpoint hosted by a
// HealthITProduct and populated by a ProviderOrganization.
// Information about the FHIR API endpoint is populated by the FHIR
// capability statement found at that endpoint as well as information
// discovered about the IP address of the endpoint.
type FHIREndpointInfo struct {
	ID                  int
	FHIREndpointID      sql.NullInt64
	HealthITProductID   sql.NullInt64
	TLSVersion          string
	MIMETypes           []string
	HTTPResponse        int
	Errors              string
	Vendor              string
	CapabilityStatement capabilityparser.CapabilityStatement // the JSON representation of the FHIR capability statement
	Validation          map[string]interface{}
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Equal checks each field of the two FHIREndpointInfos except for the database ID, CreatedAt and UpdatedAt fields to see if they are equal.
func (e *FHIREndpointInfo) Equal(e2 *FHIREndpointInfo) bool {
	if e == nil && e2 == nil {
		return true
	} else if e == nil {
		return false
	} else if e2 == nil {
		return false
	}

	if e.FHIREndpointID != e2.FHIREndpointID {
		return false
	}
	if e.HealthITProductID != e2.HealthITProductID {
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
	if e.Vendor != e2.Vendor {
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
