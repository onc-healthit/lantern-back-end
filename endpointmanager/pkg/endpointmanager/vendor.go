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

// // Equal checks each field of the two HealthITProducts except for the database ID, CHPL ID, CreatedAt and UpdatedAt fields to see if they are equal.
// func (hitp *HealthITProduct) Equal(hitp2 *HealthITProduct) bool {
// 	if hitp == nil && hitp2 == nil {
// 		return true
// 	} else if hitp == nil {
// 		return false
// 	} else if hitp2 == nil {
// 		return false
// 	}

// 	if hitp.Name != hitp2.Name {
// 		return false
// 	}
// 	if hitp.Version != hitp2.Version {
// 		return false
// 	}
// 	if hitp.Developer != hitp2.Developer {
// 		return false
// 	}
// 	if !hitp.Location.Equal(hitp2.Location) {
// 		return false
// 	}
// 	if hitp.AuthorizationStandard != hitp2.AuthorizationStandard {
// 		return false
// 	}
// 	if hitp.APISyntax != hitp2.APISyntax {
// 		return false
// 	}
// 	if hitp.APIURL != hitp2.APIURL {
// 		return false
// 	}
// 	if !cmp.Equal(hitp.CertificationCriteria, hitp2.CertificationCriteria) {
// 		return false
// 	}
// 	if hitp.CertificationStatus != hitp2.CertificationStatus {
// 		return false
// 	}
// 	if !hitp.CertificationDate.Equal(hitp2.CertificationDate) {
// 		return false
// 	}
// 	if hitp.CertificationEdition != hitp2.CertificationEdition {
// 		return false
// 	}
// 	if !hitp.LastModifiedInCHPL.Equal(hitp2.LastModifiedInCHPL) {
// 		return false
// 	}

// 	return true
// }

// // Update updates the receiver HealthITIProduct with entries from the provided HealthITProduct.
// func (hitp *HealthITProduct) Update(hitp2 *HealthITProduct) error {
// 	if hitp == nil || hitp2 == nil {
// 		return errors.New("HealthITPrdouct.Update: a given health IT product is nil")
// 	}

// 	hitp.Name = hitp2.Name
// 	hitp.Version = hitp2.Version
// 	hitp.Developer = hitp2.Developer
// 	hitp.Location = hitp2.Location
// 	hitp.AuthorizationStandard = hitp2.AuthorizationStandard
// 	hitp.APISyntax = hitp2.APISyntax
// 	hitp.APIURL = hitp2.APIURL
// 	hitp.CertificationCriteria = hitp2.CertificationCriteria
// 	hitp.CertificationStatus = hitp2.CertificationStatus
// 	hitp.CertificationDate = hitp2.CertificationDate
// 	hitp.CertificationEdition = hitp2.CertificationEdition
// 	hitp.LastModifiedInCHPL = hitp2.LastModifiedInCHPL
// 	hitp.CHPLID = hitp2.CHPLID

// 	return nil
// }
