package endpointmanager

import (
	"time"
)

// NPIOrganization represents a hospital Group, Corporation or Partnership
// From https://data.medicare.gov/Hospital-Compare/Hospital-General-Information/xubh-q36u
type NPIOrganization struct {
	ID                      int
	NPI_ID                  string
	Name                    string
	SecondaryName           string
	Location                *Location
	Taxonomy                string // Taxonomy code mapping: http://www.wpc-edi.com/reference/codelists/healthcare/health-care-provider-taxonomy-code-set/
	NormalizedName          string
	NormalizedSecondaryName string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// Equal checks each field of the two NPIOrganizations except for the database ID, CreatedAt and UpdatedAt fields to see if they are equal.
func (org *NPIOrganization) Equal(org2 *NPIOrganization) bool {
	if org == nil && org2 == nil {
		return true
	} else if org == nil {
		return false
	} else if org2 == nil {
		return false
	}
	if org.NPI_ID != org2.NPI_ID {
		return false
	}
	if org.Name != org2.Name {
		return false
	}
	if org.SecondaryName != org2.SecondaryName {
		return false
	}
	if !org.Location.Equal(org2.Location) {
		return false
	}
	if org.Taxonomy != org2.Taxonomy {
		return false
	}

	return true
}
