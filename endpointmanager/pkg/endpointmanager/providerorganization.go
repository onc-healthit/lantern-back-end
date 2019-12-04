package endpointmanager

import (
	"context"
	"time"
)

// ProviderOrganization represents a hospital or group practice.
// Other organization types may be added in the future.
// From https://data.medicare.gov/Hospital-Compare/Hospital-General-Information/xubh-q36u
type ProviderOrganization struct {
	ID               int
	Name             string
	URL              string
	Location         *Location
	OrganizationType string // "hospital" or "group practice"
	HospitalType     string // only applicable if the OrganizationType is "hospital". Otherwise, this should be "". Examples: "Acute Care", "Critical Access", "Psychiatric", etc.
	Ownership        string // The organization type that owns the hospital. Only applicable if the OrganizationType is "hospital". Otherwise, this should be nil. Examples: "Volunary non-profit", "Government - State", "Proprietary", etc.
	Beds             int    // the number of beds that the hospital has. This is only applicable if OrganizationType is "hospital". Otherwise, this should be -1. This is an indicator of relative size of the hospital compared to other hospitals.
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ProviderOrganizationStore is the interface for interacting with the storage layer that holds
// provider organization objects.
type ProviderOrganizationStore interface {
	GetProviderOrganization(context.Context, int) (*ProviderOrganization, error)

	AddProviderOrganization(context.Context, *ProviderOrganization) error
	UpdateProviderOrganization(context.Context, *ProviderOrganization) error
	DeleteProviderOrganization(context.Context, *ProviderOrganization) error

	Close()
}

// Equal checks each field of the two ProviderOrganizations except for the database ID, CreatedAt and UpdatedAt fields to see if they are equal.
func (po *ProviderOrganization) Equal(po2 *ProviderOrganization) bool {
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
	if po.URL != po2.URL {
		return false
	}
	if !po.Location.Equal(po2.Location) {
		return false
	}
	if po.OrganizationType != po2.OrganizationType {
		return false
	}
	if po.HospitalType != po2.HospitalType {
		return false
	}
	if po.Ownership != po2.Ownership {
		return false
	}
	if po.Beds != po2.Beds {
		return false
	}

	return true
}
