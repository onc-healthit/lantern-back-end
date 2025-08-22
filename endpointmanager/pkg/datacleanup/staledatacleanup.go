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
			tx.Rollback()
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
		WHERE is_chpl = true 
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
		WHERE list_source = ANY($1) AND is_chpl = true
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

	// Step 1: Get all URLs from endpoints to be deleted
	findEndpointsQuery := `
		SELECT DISTINCT url FROM fhir_endpoints 
		WHERE list_source = ANY($1)
	`

	rows, err := tx.QueryContext(ctx, findEndpointsQuery, pq.Array(listSources))
	if err != nil {
		return fmt.Errorf("failed to find endpoints for list sources: %w", err)
	}
	defer rows.Close()

	var endpointURLs []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return fmt.Errorf("failed to scan endpoint URL: %w", err)
		}
		endpointURLs = append(endpointURLs, url)
	}

	if len(endpointURLs) == 0 {
		log.Info("No endpoints found for the stale list sources")
		return nil
	}

	log.Infof("Found %d unique endpoints to delete", len(endpointURLs))

	// Process URLs in batches to avoid parameter limits
	batchSize := 1000
	for i := 0; i < len(endpointURLs); i += batchSize {
		end := i + batchSize
		if end > len(endpointURLs) {
			end = len(endpointURLs)
		}

		batch := endpointURLs[i:end]
		log.Infof("Processing batch %d-%d (%d URLs)", i+1, end, len(batch))

		err = deleteBatchEndpointData(ctx, tx, batch)
		if err != nil {
			return fmt.Errorf("failed to delete batch %d-%d: %w", i+1, end, err)
		}
	}

	// Step 2: Delete from fhir_endpoints (this should be last)
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

func deleteBatchEndpointData(ctx context.Context, tx *sql.Tx, urls []string) error {
	// Delete in order of foreign key dependencies
	// REMOVED: fhir_endpoints_info, fhir_endpoints_info_history, fhir_endpoints_availability, fhir_endpoints_metadata
	// These tables preserve history and metadata and should NOT be deleted

	// 1. Delete from fhir_endpoint_organizations first
	result, err := tx.ExecContext(ctx, `
		DELETE FROM fhir_endpoint_organizations 
		WHERE id IN (
			SELECT m.org_database_id 
			FROM fhir_endpoints e, fhir_endpoint_organizations_map m 
			WHERE e.id = m.id 
			AND e.url = ANY($1)
		)
	`, pq.Array(urls))
	if err != nil {
		log.Errorf("Failed to delete from fhir_endpoint_organizations: %v", err)
		return err
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		log.Infof("Deleted %d records from fhir_endpoint_organizations", rowsAffected)
	} else {
		log.Infof("No records to delete from fhir_endpoint_organizations")
	}

	// 2. Delete from fhir_endpoint_organizations_map
	result, err = tx.ExecContext(ctx, `
		DELETE FROM fhir_endpoint_organizations_map 
		WHERE id IN (
			SELECT id FROM fhir_endpoints 
			WHERE url = ANY($1)
		)
	`, pq.Array(urls))
	if err != nil {
		log.Errorf("Failed to delete from fhir_endpoint_organizations_map: %v", err)
		return err
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		log.Infof("Deleted %d records from fhir_endpoint_organizations_map", rowsAffected)
	} else {
		log.Infof("No records to delete from fhir_endpoint_organizations_map")
	}

	// 3. Delete from endpoint_organization
	result, err = tx.ExecContext(ctx, `
		DELETE FROM endpoint_organization 
		WHERE url = ANY($1)
	`, pq.Array(urls))
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
