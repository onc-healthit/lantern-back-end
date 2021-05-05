// +build integration

package postgresql

import (
	"context"
	"fmt"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_PersistCriteria(t *testing.T) {
	SetupStore()
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

	var crit1 = &endpointmanager.CertificationCriteria{
		CertificationID:        44,
		CertificationNumber:    "170.315 (f)(2)",
		Title:                  "Transmission to Public Health Agencies - Syndromic Surveillance",
		CertificationEditionID: 3,
		CertificationEdition:   "2015",
		Description:            "Syndromic Surveillance",
		Removed:                false,
	}
	var crit2 = &endpointmanager.CertificationCriteria{
		CertificationID:        64,
		CertificationNumber:    "170.314 (a)(4)",
		Title:                  "Vital signs, body mass index, and growth Charts",
		CertificationEditionID: 2,
		CertificationEdition:   "2014",
		Description:            "Vital signs",
		Removed:                false,
	}

	// add criteria

	err = store.AddCriteria(ctx, crit1)
	th.Assert(t, err == nil, fmt.Errorf("Error adding criteria: %s", err))

	err = store.AddCriteria(ctx, crit2)
	th.Assert(t, err == nil, fmt.Errorf("Error adding criteria: %s", err))

	// retrieve criteria

	c1, err := store.GetCriteria(ctx, crit1.ID)
	t.Logf("HERE 1 %d", crit1.ID)
	th.Assert(t, err == nil, fmt.Errorf("Error getting criteria: %s", err))
	th.Assert(t, c1.Equal(crit1), fmt.Errorf("Error adding criteria: %s", err))

	c2, err := store.GetCriteriaByCertificationID(ctx, crit2.CertificationID)
	t.Logf("HERE 2 %d", crit2.ID)
	th.Assert(t, err == nil, fmt.Errorf("Error getting criteria: %s", err))
	th.Assert(t, c2.Equal(crit2), "retrieved criteria is not equal to saved criteria.")

	// update criteria

	c1.CertificationEdition = "2020"

	err = store.UpdateCriteria(ctx, c1)
	t.Logf("HERE C1 %d", c1.ID)
	th.Assert(t, err == nil, fmt.Errorf("Error updating criteria: %s", err))

	c1, err = store.GetCriteria(ctx, crit1.ID)
	t.Logf("HERE 3 %d", crit1.ID)
	t.Logf("HERE 2 C1 %d", c1.ID)
	th.Assert(t, err == nil, fmt.Errorf("Error getting criteria: %s", err))
	th.Assert(t, !(c1.Equal(crit1)), "retrieved updated criteria is equal to original criteria.")
	th.Assert(t, !(c1.UpdatedAt.Equal(c1.CreatedAt)), "UpdatedAt is not being properly set on update.")

	// delete criteria

	err = store.DeleteCriteria(ctx, crit1)
	th.Assert(t, err == nil, fmt.Errorf("Error deleting criteria: %s", err))

	_, err = store.GetCriteria(ctx, crit1.ID) // ensure we deleted the entry
	t.Logf("HERE 4 %d", crit1.ID)
	th.Assert(t, err != nil, fmt.Errorf("crit1 was not deleted: %s", err))

	_, err = store.GetCriteria(ctx, crit2.ID) // ensure we haven't deleted all entries
	th.Assert(t, err == nil, fmt.Errorf("error retrieving crit2 after deleting crit1: %s", err))

	err = store.DeleteCriteria(ctx, crit2)
	th.Assert(t, err == nil, fmt.Errorf("Error deleting criteria: %s", err))
}
