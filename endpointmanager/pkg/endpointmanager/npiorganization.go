package endpointmanager

import (
	"time"
)

// NPIOrganization represents a hospital Group, Corporation or Partnership
// From https://data.medicare.gov/Hospital-Compare/Hospital-General-Information/xubh-q36u
type NPIOrganization struct {
	ID               int
	NPI_ID			 string
	Name             string
	SecondaryName    string
	FHIREndpoint     *FHIREndpoint
	Location         *Location
	Taxonomy 		 string // Taxonomy code mapping: http://www.wpc-edi.com/reference/codelists/healthcare/health-care-provider-taxonomy-code-set/
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NPIOrganizationStore is the interface for interacting with the storage layer that holds
// NPIOrganization objects.
type NPIOrganizationStore interface {
	GetNPIOrganization(int) (*NPIOrganization, error)

	AddNPIOrganization(*NPIOrganization) error
	UpdateNPIOrganization(*NPIOrganization) error
	DeleteNPIOrganization(*NPIOrganization) error

	Close()
}

// Equal checks each field of the two NPIOrganizations except for the database ID, CreatedAt and UpdatedAt fields to see if they are equal.
func (po *NPIOrganization) Equal(po2 *NPIOrganization) bool {
	if po == nil && po2 == nil {
		return true
	} else if po == nil {
		return false
	} else if po2 == nil {
		return false
	}
	if po.Name != po2.Name {
		return false
	}
	if po.SecondaryName != po2.SecondaryName {
		return false
	}
	if !po.Location.Equal(po2.Location) {
		return false
	}
	if !po.FHIREndpoint.Equal(po2.FHIREndpoint) {
		return false
	}
	if po.Taxonomy != po2.Taxonomy {
		return false
	}

	return true
}
