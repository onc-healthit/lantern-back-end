package endpointmanager

import (
	"time"

	"github.com/pkg/errors"
)

// @TODO Update all comments
// HealthITProduct represents a health IT vendor product such as an
// EHR. This information is gathered from the Certified Health IT Products List
// (CHPL).
type CertificationCriteria struct {
	ID                     int
	CertificationID        int
	CertificationNumber    string
	Title                  string
	CertificationEditionID int
	CertificationEdition   string
	Description            string
	Removed                bool
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// Equal checks each field of the two HealthITProducts except for the database ID, CHPL ID, CreatedAt and UpdatedAt fields to see if they are equal.
func (certCri *CertificationCriteria) Equal(certCri2 *CertificationCriteria) bool {
	if certCri == nil && certCri2 == nil {
		return true
	} else if certCri == nil {
		return false
	} else if certCri2 == nil {
		return false
	}

	if certCri.ID != certCri2.ID {
		return false
	}
	if certCri.CertificationID != certCri2.CertificationID {
		return false
	}
	if certCri.CertificationNumber != certCri2.CertificationNumber {
		return false
	}
	if certCri.Title != certCri2.Title {
		return false
	}
	if certCri.CertificationEditionID != certCri2.CertificationEditionID {
		return false
	}
	if certCri.CertificationEdition != certCri2.CertificationEdition {
		return false
	}
	if certCri.Description != certCri2.Description {
		return false
	}
	if certCri.Removed != certCri2.Removed {
		return false
	}

	return true
}

// Update updates the receiver HealthITIProduct with entries from the provided HealthITProduct.
func (certCri *CertificationCriteria) Update(certCri2 *CertificationCriteria) error {
	if certCri == nil || certCri2 == nil {
		return errors.New("CertificationCriteria.Update: a given health IT certification criteria is nil")
	}

	certCri.CertificationID = certCri2.CertificationID
	certCri.CertificationNumber = certCri2.CertificationNumber
	certCri.Title = certCri2.Title
	certCri.CertificationEditionID = certCri2.CertificationEditionID
	certCri.CertificationEdition = certCri2.CertificationEdition
	certCri.Description = certCri2.Description
	certCri.Removed = certCri2.Removed

	return nil
}
