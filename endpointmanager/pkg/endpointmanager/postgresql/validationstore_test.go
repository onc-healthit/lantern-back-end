// +build integration

package postgresql

import (
	"context"
	"fmt"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

func Test_PersistValidation(t *testing.T) {
	teardown, _ := th.IntegrationDBTestSetup(t, store.DB)
	defer teardown(t, store.DB)

	var err error
	ctx := context.Background()

	// validation objects
	testValidation1 := endpointmanager.Validation{
		Results: []endpointmanager.Rule{
			{
				RuleName: endpointmanager.CapStatExistRule,
				Valid:    true,
				Expected: "true",
				Actual:   "true",
				Comment:  "The Conformance Resource exists. Servers SHALL provide a Conformance Resource that specifies which interactions and resources are supported.",
			},
		},
	}
	testValidation2 := endpointmanager.Validation{
		Results: []endpointmanager.Rule{
			{
				RuleName: endpointmanager.CapStatExistRule,
				Valid:    true,
				Expected: "true",
				Actual:   "true",
				Comment:  "The Conformance Resource exists. Servers SHALL provide a Conformance Resource that specifies which interactions and resources are supported.",
			},
			{
				RuleName:  endpointmanager.OtherResourceExists,
				Valid:     true,
				Actual:    "true",
				Expected:  "true",
				Comment:   "The US Core Server SHALL support at least one additional resource profile (besides Patient) from the list of US Core Profiles.",
				ImplGuide: "USCore 3.1",
				Reference: "https://www.hl7.org/fhir/us/core/CapabilityStatement-us-core-server.html",
			},
		},
	}

	// add validation result

	valResID1, err := store.AddValidationResult(ctx)
	th.Assert(t, err == nil, fmt.Sprintf("Error adding validation result ID: %s", err))

	valResID2, err := store.AddValidationResult(ctx)
	th.Assert(t, err == nil, fmt.Sprintf("Error adding validation result ID: %s", err))

	// check that there are two ids in the validation_result table

	var count int
	valResRow := store.DB.QueryRow("SELECT COUNT(*) FROM validation_results;")
	err = valResRow.Scan(&count)
	th.Assert(t, err == nil, fmt.Sprintf("Error getting validation result table count: %s", err))
	th.Assert(t, count == 2, fmt.Sprintf("Should only be two entries in validation result table, is instead %d", count))

	// add validations

	err = store.AddValidation(ctx, &testValidation1, valResID1)
	th.Assert(t, err == nil, fmt.Sprintf("Error adding validation object 1 to table: %s", err))

	err = store.AddValidation(ctx, &testValidation2, valResID2)
	th.Assert(t, err == nil, fmt.Sprintf("Error adding validation object 2 to table: %s", err))

	// retrieve validations

	validationRows, err := store.GetValidationByID(ctx, valResID1)
	th.Assert(t, err == nil, fmt.Sprintf("Error getting validation from ID %d, error: %s", valResID1, err))
	th.Assert(t, len(*validationRows) == 1, fmt.Sprintf("ID %d should have length 1, is instead %d", valResID1, len(*validationRows)))

	validationRows, err = store.GetValidationByID(ctx, valResID2)
	th.Assert(t, err == nil, fmt.Sprintf("Error getting validation from ID %d, error: %s", valResID2, err))
	th.Assert(t, len(*validationRows) == 2, fmt.Sprintf("ID %d should have length 2, is instead %d", valResID2, len(*validationRows)))
}
