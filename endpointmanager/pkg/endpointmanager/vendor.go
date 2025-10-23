package endpointmanager

import (
	"time"
)

// Vendor represents a Health IT vendor. This information is gathered from the
// Certified Health IT Products List (CHPL).
type Vendor struct {
	ID                 int
	Name               string
	DeveloperCode      string
	URL                string
	Location           *Location // the address listed in CHPL for the Developer.
	Status             string
	LastModifiedInCHPL time.Time
	CHPLID             int // the product's unique ID within the CHPL system.
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Equal checks each field of the two Vendors except for the database ID, CreatedAt and UpdatedAt fields to see if they are equal.
func (v *Vendor) Equal(v2 *Vendor) bool {
	if v == nil && v2 == nil {
		return true
	} else if v == nil {
		return false
	} else if v2 == nil {
		return false
	}

	if v.Name != v2.Name {
		return false
	}
	if v.DeveloperCode != v2.DeveloperCode {
		return false
	}
	if v.URL != v2.URL {
		return false
	}
	if !v.Location.Equal(v2.Location) {
		return false
	}
	if v.Status != v2.Status {
		return false
	}
	if !v.LastModifiedInCHPL.Equal(v2.LastModifiedInCHPL) {
		return false
	}
	if v.CHPLID != v2.CHPLID {
		return false
	}

	return true
}
