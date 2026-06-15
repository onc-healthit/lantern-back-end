package datacleanup

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	log "github.com/sirupsen/logrus"
)

// CleanupStaleData removes data for CHPL list sources that weren't updated in the current run.
func CleanupStaleData(ctx context.Context, store *postgresql.Store, populationStartTime time.Time) error {
	log.Info("Starting cleanup of stale CHPL data...")

	staleSources, err := getStaleListSources(ctx, store, populationStartTime)
	if err != nil {
		return fmt.Errorf("failed to get stale list sources: %w", err)
	}

	if len(staleSources) == 0 {
		log.Info("No stale CHPL list sources found.")
		return nil
	}

	log.Infof("Found %d stale CHPL list sources to cleanup: %v", len(staleSources), staleSources)

	// Start transaction for all cleanup operations
	tx, err := store.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			rbErr := tx.Rollback()
			if rbErr != nil {
				log.Warnf("failed to rollback transaction: %v", rbErr)
			}
		}
	}()

	// Process all stale sources in batches for better performance
	err = deleteEndpointDataBatch(ctx, tx, staleSources)
	if err != nil {
		return fmt.Errorf("failed to delete endpoint data: %w", err)
	}

	err = deleteListSourcesBatch(ctx, tx, staleSources)
	if err != nil {
		return fmt.Errorf("failed to delete stale list sources: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit cleanup transaction: %w", err)
	}

	log.Infof("Cleanup completed successfully: processed %d stale CHPL list sources", len(staleSources))
	return nil
}

func getStaleListSources(ctx context.Context, store *postgresql.Store, since time.Time) ([]string, error) {
	query := `
		SELECT list_source
		FROM list_source_info
		WHERE is_chpl = 'CHPL'
		AND updated_at < $1
		ORDER BY list_source
	`

	rows, err := store.DB.QueryContext(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query stale list sources: %w", err)
	}
	defer rows.Close()

	var staleSources []string
	for rows.Next() {
		var listSource string
		if err := rows.Scan(&listSource); err != nil {
			return nil, fmt.Errorf("failed to scan stale list source: %w", err)
		}
		staleSources = append(staleSources, listSource)
	}

	return staleSources, nil
}

func deleteListSourcesBatch(ctx context.Context, tx *sql.Tx, listSources []string) error {
	if len(listSources) == 0 {
		return nil
	}

	query := `
		DELETE FROM list_source_info
		WHERE list_source = ANY($1) AND is_chpl = 'CHPL'
	`

	result, err := tx.ExecContext(ctx, query, pq.Array(listSources))
	if err != nil {
		return fmt.Errorf("failed to delete stale list sources: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Infof("Deleted %d stale list sources from list_source_info", rowsAffected)

	return nil
}

func deleteEndpointDataBatch(ctx context.Context, tx *sql.Tx, listSources []string) error {
	log.Infof("Deleting endpoint data for %d list sources", len(listSources))

	err := deleteBatchEndpointData(ctx, tx, listSources)
	if err != nil {
		return fmt.Errorf("failed to delete endpoint data: %w", err)
	}

	// Delete from fhir_endpoints last
	result, err := tx.ExecContext(ctx, `
		DELETE FROM fhir_endpoints
		WHERE list_source = ANY($1)
	`, pq.Array(listSources))
	if err != nil {
		return fmt.Errorf("failed to delete from fhir_endpoints: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Infof("Deleted %d records from fhir_endpoints", rowsAffected)

	return nil
}

func deleteBatchEndpointData(ctx context.Context, tx *sql.Tx, listSources []string) error {
	// Delete in order of foreign key dependencies
	// REMOVED: fhir_endpoints_info, fhir_endpoints_info_history, fhir_endpoints_availability, fhir_endpoints_metadata
	// These tables preserve history and metadata and should NOT be deleted

	// 1. Delete from fhir_endpoint_organizations first.
	// Scoped by list_source (via fhir_endpoints.id) so that org records shared with
	// a live list source that happens to contain the same URL are not affected.
	result, err := tx.ExecContext(ctx, `
		DELETE FROM fhir_endpoint_organizations
		WHERE id IN (
			SELECT m.org_database_id
			FROM fhir_endpoints e
			JOIN fhir_endpoint_organizations_map m ON e.id = m.id
			WHERE e.list_source = ANY($1)
		)
	`, pq.Array(listSources))
	if err != nil {
		log.Errorf("Failed to delete from fhir_endpoint_organizations: %v", err)
		return err
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		log.Infof("Deleted %d records from fhir_endpoint_organizations", rowsAffected)
	} else {
		log.Infof("No records to delete from fhir_endpoint_organizations")
	}

	// 2. Delete from fhir_endpoint_organizations_map.
	// Scoped by list_source so entries for the same URL in a live list source are preserved.
	result, err = tx.ExecContext(ctx, `
		DELETE FROM fhir_endpoint_organizations_map
		WHERE id IN (
			SELECT id FROM fhir_endpoints
			WHERE list_source = ANY($1)
		)
	`, pq.Array(listSources))
	if err != nil {
		log.Errorf("Failed to delete from fhir_endpoint_organizations_map: %v", err)
		return err
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		log.Infof("Deleted %d records from fhir_endpoint_organizations_map", rowsAffected)
	} else {
		log.Infof("No records to delete from fhir_endpoint_organizations_map")
	}

	// 3. Delete from endpoint_organization only for URLs that are no longer referenced
	// by any live list source. If a URL exists in another active list source, keep it.
	// NOT EXISTS is used instead of NOT IN to correctly handle NULL urls and improve
	// performance via index on (url, list_source).
	result, err = tx.ExecContext(ctx, `
		DELETE FROM endpoint_organization
		WHERE url IN (
			SELECT url FROM fhir_endpoints WHERE list_source = ANY($1)
		)
		AND NOT EXISTS (
			SELECT 1 FROM fhir_endpoints fe2
			WHERE fe2.url = endpoint_organization.url
			AND NOT (fe2.list_source = ANY($1))
		)
	`, pq.Array(listSources))
	if err != nil {
		log.Errorf("Failed to delete from endpoint_organization: %v", err)
		return err
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		log.Infof("Deleted %d records from endpoint_organization", rowsAffected)
	} else {
		log.Infof("No records to delete from endpoint_organization")
	}

	log.Info("Preserving fhir_endpoints_info, fhir_endpoints_info_history, fhir_endpoints_availability, and fhir_endpoints_metadata tables")

	return nil
}
