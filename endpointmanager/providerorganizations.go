package main

import (
	"time"
)

// ProviderOrganization represents a hospital or group practice.
// Other organization types may be added in the future.
// From https://data.medicare.gov/Hospital-Compare/Hospital-General-Information/xubh-q36u
type ProviderOrganization struct {
	OrganizationID   int
	Name             string
	URL              string
	Location         Location
	OrganizationType string // "hospital" or "group practice"
	HospitalType     string // only applicable if the OrganizationType is "hospital". Otherwise, this should be "". Examples: "Acute Care", "Critical Access", "Psychiatric", etc.
	Ownership        string // The organization type that owns the hospital. Only applicable if the OrganizationType is "hospital". Otherwise, this should be nil. Examples: "Volunary non-profit", "Government - State", "Proprietary", etc.
	Beds             int    // the number of beds that the hospital has. This is only applicable if OrganizationType is "hospital". Otherwise, this should be -1. This is an indicator of relative size of the hospital compared to other hospitals.
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// GetProviderOrganization gets a ProviderOrganization from the database using the given url as a key.
func GetProviderOrganization(organizationID string) (*ProviderOrganization, error) {
	// TODO: location
	sqlStatement := `SELECT organization_id,
							name,
							url,
							organization_type,
							hospital_type,
							ownership,
							beds,
							created_at,
							updated_at
					FROM provider_organizations WHERE name=$1 AND value=$2`
	row := db.QueryRow(sqlStatement, organizationID)
	var po ProviderOrganization

	err := row.Scan(
		&po.OrganizationID,
		&po.Name,
		&po.URL,
		&po.OrganizationType,
		&po.HospitalType,
		&po.Ownership,
		&po.Beds,
		&po.CreatedAt,
		&po.UpdatedAt)

	return &po, err
}

// Add adds the ProviderOrganization to the database.
func (po *ProviderOrganization) Add() error {
	// TODO: location
	sqlStatement := `
	INSERT INTO provider_organizations (
		organization_id,
		name,
		url,
		organization_type,
		hospital_type,
		ownership,
		beds)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := db.Exec(sqlStatement,
		po.OrganizationID,
		po.Name,
		po.URL,
		po.OrganizationType,
		po.HospitalType,
		po.Ownership,
		po.Beds)

	return err
}

// Update updates the ProviderOrganization in the database using the ProviderOrganization's URL as the key.
func (po *ProviderOrganization) Update() error {
	// TODO: location
	sqlStatement := `
	UPDATE provider_organizations
	SET organization_id = $1,
		name = $2,
		url = $3,
		organization_type = $4,
		hospital_type = $5,
		ownership = $6,
		beds = $7
	WHERE organization_id = $1`

	_, err := db.Exec(sqlStatement,
		po.OrganizationID,
		po.Name,
		po.URL,
		po.OrganizationType,
		po.HospitalType,
		po.Ownership,
		po.Beds)

	return err
}

// Delete deletes the ProviderOrganization from the databse using the ProviderOrganization's URL as the key.
func (po *ProviderOrganization) Delete() error {
	sqlStatement := `
	DELETE FROM provider_organizations
	WHERE organization_id=$1`

	_, err := db.Exec(sqlStatement, po.OrganizationID)

	return err
}

// Equal checks each field of the two ProviderOrganizations except for the CreatedAt and UpdatedAt fields to see if they are equal.
func (po *ProviderOrganization) Equal(po2 *ProviderOrganization) bool {
	if po2 == nil {
		return false
	}

	if po.OrganizationID != po2.OrganizationID {
		return false
	}
	if po.Name != po2.Name {
		return false
	}
	if po.URL != po2.URL {
		return false
	}
	if po.Location != po2.Location {
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
