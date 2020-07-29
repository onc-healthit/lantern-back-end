package endpointmanager

import (
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_CertificationCriteriaEqual(t *testing.T) {
	crit1 := &CertificationCriteria{
		ID:                     1,
		CertificationID:        44,
		CertificationNumber:    "170.315 (f)(2)",
		Title:                  "Transmission to Public Health Agencies - Syndromic Surveillance",
		CertificationEditionID: 3,
		CertificationEdition:   "2015",
		Description:            "Syndromic Surveillance",
		Removed:                false,
	}
	crit2 := &CertificationCriteria{
		ID:                     1,
		CertificationID:        44,
		CertificationNumber:    "170.315 (f)(2)",
		Title:                  "Transmission to Public Health Agencies - Syndromic Surveillance",
		CertificationEditionID: 3,
		CertificationEdition:   "2015",
		Description:            "Syndromic Surveillance",
		Removed:                false,
	}

	th.Assert(t, crit1.Equal(crit2), "Expected crit1 to equal crit2. They are not equal.")

	crit2.ID = 2
	th.Assert(t, crit1.Equal(crit2), fmt.Errorf("Expect criteria 1 to equal criteria 2. ids should be ignored. %d vs %d", crit1.ID, crit2.ID))
	crit2.ID = crit1.ID

	crit2.CertificationID = 45
	th.Assert(t, !(crit1.Equal(crit2)), fmt.Errorf("Did not expect criteria 1 to equal criteria 2. CertificationID should be different. %d vs %d", crit1.CertificationID, crit2.CertificationID))
	crit2.CertificationID = crit1.CertificationID

	crit2.CertificationNumber = "other"
	th.Assert(t, !(crit1.Equal(crit2)), fmt.Errorf("Did not expect criteria 1 to equal criteria 2. CertificationNumber should be different. %s vs %s", crit1.CertificationNumber, crit2.CertificationNumber))
	crit2.CertificationNumber = crit1.CertificationNumber

	crit2.Title = "other"
	th.Assert(t, !(crit1.Equal(crit2)), fmt.Errorf("Did not expect criteria 1 to equal criteria 2. Title should be different. %s vs %s", crit1.Title, crit2.Title))
	crit2.Title = crit1.Title

	crit2.CertificationEditionID = 4
	th.Assert(t, !(crit1.Equal(crit2)), fmt.Errorf("Did not expect criteria 1 to equal criteria 2. CertificationEditionID should be different. %d vs %d", crit1.CertificationEditionID, crit2.CertificationEditionID))
	crit2.CertificationEditionID = crit1.CertificationEditionID

	crit2.CertificationEdition = "other"
	th.Assert(t, !(crit1.Equal(crit2)), fmt.Errorf("Did not expect criteria 1 to equal criteria 2. CertificationEdition should be different. %s vs %s", crit1.CertificationEdition, crit2.CertificationEdition))
	crit2.CertificationEdition = crit1.CertificationEdition

	crit2.Description = "other"
	th.Assert(t, !(crit1.Equal(crit2)), fmt.Errorf("Did not expect criteria 1 to equal criteria 2. Description should be different. %s vs %s", crit1.Description, crit2.Description))
	crit2.Description = crit1.Description

	crit2.Removed = true
	th.Assert(t, !(crit1.Equal(crit2)), fmt.Errorf("Did not expect criteria 1 to equal criteria 2. Removed should be different. %v vs %v", crit1.Removed, crit2.Removed))
	crit2.Removed = crit1.Removed

	crit2 = nil
	th.Assert(t, !(crit1.Equal(crit2)), "Did not expect crit1 to equal nil crit2.")
	crit2 = crit1

	crit1 = nil
	th.Assert(t, !(crit1.Equal(crit2)), "Did not expect nil crit1 to equal crit2.")

	crit2 = nil
	th.Assert(t, crit1.Equal(crit2), "Nil crit1 should equal nil crit2.")
}
