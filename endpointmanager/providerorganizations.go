package main

import "time"

// ProviderOrganization represents a hospital or group practice.
// Other organization types may be added in the future.
type ProviderOrganization struct {
	OrganizationID int
	Name           string
	Location       Location
	// OrganizationType is either "hospital" or "group practice"
	OrganizationType string
	// HospitalType is the type of hospital and is only applicable if the
	// OrganizationType is "hospital". Otherwise, this should be nil.
	// Examples of HospitalType include "Acute Care", "Critical Access", "Psychiatric", etc.
	// This information is gathered from https://data.medicare.gov/Hospital-Compare/Hospital-General-Information/xubh-q36u.
	HospitalType string
	// Ownership is the type of organization that owns the hospital and is only
	// applicable if OrganizationType is "hospital". Otherwise, this should be nil.
	// Examples of Orwnership include "Volunary non-profit", "Government - State", "Proprietary", etc.
	// This information is gathered from https://data.medicare.gov/Hospital-Compare/Hospital-General-Information/xubh-q36u.
	Ownership string
	// Beds is the number of beds that the hospital has. This is only applicable if
	// OrganizationType is "hospital". Otherwise, this should be -1.
	// This is an indicator of relative size of the hospital compared to other hospitals.
	Beds      int
	CreatedAt time.Time
	UpdatedAt time.Time
}
