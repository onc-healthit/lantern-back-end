package endpointmanager

import (
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
)

// FHIREndpointInfo represents a fielded FHIR API endpoint hosted by a
// HealthITProduct and populated by a ProviderOrganization.
// Information about the FHIR API endpoint is populated by the FHIR
// capability statement found at that endpoint.
type FHIREndpointInfo struct {
	ID                  int
	HealthITProductID   int
	URL                 string
	TLSVersion          string
	MIMETypes           []string
	HTTPResponse        int
	Errors              string
	VendorID            int
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

	if e.URL != e2.URL {
		return false
	}
	if e.HealthITProductID != e2.HealthITProductID {
		return false
	}

	if e.TLSVersion != e2.TLSVersion {
		return false
	}

	if !helpers.StringArraysEqual(e.MIMETypes, e2.MIMETypes) {
		return false
	}

	if e.HTTPResponse != e2.HTTPResponse {
		return false
	}
	if e.Errors != e2.Errors {
		return false
	}
	if e.VendorID != e2.VendorID {
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
