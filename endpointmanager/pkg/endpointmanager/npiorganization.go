package endpointmanager

import (
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
)

// NPIOrganization represents a hospital Group, Corporation or Partnership
// From https://data.medicare.gov/Hospital-Compare/Hospital-General-Information/xubh-q36u
type NPIOrganization struct {
	ID              int
	NPI_ID          string
	Names           []string
	Location        *Location
	Taxonomy        string // Taxonomy code mapping: http://www.wpc-edi.com/reference/codelists/healthcare/health-care-provider-taxonomy-code-set/
	NormalizedNames []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
	if !helpers.StringArraysEqual(org.Names, org2.Names) {
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
