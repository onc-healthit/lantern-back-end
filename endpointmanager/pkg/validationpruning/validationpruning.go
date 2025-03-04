package validationpruning

import (
	"context"
	"log"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

func PruneValidationInfo(ctx context.Context, store *postgresql.Store) {
	query := `
		DELETE FROM validations v
		WHERE v.validation_result_id IS NOT NULL
		AND NOT EXISTS (
			SELECT 1 FROM fhir_endpoints_info fei WHERE fei.validation_result_id = v.validation_result_id
		)
		AND NOT EXISTS (
			SELECT 1 FROM fhir_endpoints_info_history feih WHERE feih.validation_result_id = v.validation_result_id
		);
	`

	// Execute the DELETE query within the provided context
	result, err := store.DB.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Error pruning validation records: %v\n", err)
		return
	}

	// Fetch the number of rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error fetching affected rows: %v\n", err)
		return
	}

	log.Printf("PruneValidationInfo: Successfully deleted %d orphaned validation records.\n", rowsAffected)
}
