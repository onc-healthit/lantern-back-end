package datacleanup

import (
	"context"
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

	for _, source := range staleSources {
		err = deleteEndpointDataByListSource(ctx, store, source)
		if err != nil {
			log.Errorf("Failed to delete endpoint data for list source %s: %v", source, err)
		}
	}

	err = deleteListSources(ctx, store, staleSources)
	if err != nil {
		return fmt.Errorf("failed to delete stale list sources: %w", err)
	}

	log.Infof("Cleanup completed: processed %d stale CHPL list sources", len(staleSources))
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

func deleteListSources(ctx context.Context, store *postgresql.Store, listSources []string) error {
	if len(listSources) == 0 {
		return nil
	}

	query := `
		DELETE FROM list_source_info 
		WHERE list_source = ANY($1) AND is_chpl = true
	`

	result, err := store.DB.ExecContext(ctx, query, pq.Array(listSources))
	if err != nil {
		return fmt.Errorf("failed to delete stale list sources: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Infof("Deleted %d stale list sources from list_source_info", rowsAffected)

	return nil
}

func deleteEndpointDataByListSource(ctx context.Context, store *postgresql.Store, listSource string) error {
	log.Infof("Deleting endpoint data for list source: %s", listSource)

	findEndpointsQuery := `
		SELECT id, url FROM fhir_endpoints 
		WHERE list_source = $1
	`

	rows, err := store.DB.QueryContext(ctx, findEndpointsQuery, listSource)
	if err != nil {
		return fmt.Errorf("failed to find endpoints for list source %s: %w", listSource, err)
	}
	defer rows.Close()

	var endpointURLs []string
	for rows.Next() {
		var id int
		var url string
		if err := rows.Scan(&id, &url); err != nil {
			return fmt.Errorf("failed to scan endpoint: %w", err)
		}
		endpointURLs = append(endpointURLs, url)
	}

	if len(endpointURLs) == 0 {
		log.Infof("No endpoints found for list source: %s", listSource)
		return nil
	}

	log.Infof("Deleting %d endpoints for list source: %s", len(endpointURLs), listSource)

	_, err = store.DB.ExecContext(ctx, `
		DELETE FROM fhir_endpoints_info_history 
		WHERE url = ANY($1)
	`, pq.Array(endpointURLs))
	if err != nil {
		log.Errorf("Failed to delete from fhir_endpoints_info_history: %v", err)
	}

	_, err = store.DB.ExecContext(ctx, `
		DELETE FROM fhir_endpoints_info 
		WHERE url = ANY($1)
	`, pq.Array(endpointURLs))
	if err != nil {
		log.Errorf("Failed to delete from fhir_endpoints_info: %v", err)
	}

	_, err = store.DB.ExecContext(ctx, `
		DELETE FROM fhir_endpoints_metadata 
		WHERE url = ANY($1)
		AND id NOT IN (
			SELECT DISTINCT metadata_id FROM fhir_endpoints_info WHERE metadata_id IS NOT NULL
			UNION
			SELECT DISTINCT metadata_id FROM fhir_endpoints_info_history WHERE metadata_id IS NOT NULL
		)
	`, pq.Array(endpointURLs))
	if err != nil {
		log.Errorf("Failed to delete from fhir_endpoints_metadata: %v", err)
	}

	_, err = store.DB.ExecContext(ctx, `
		DELETE FROM endpoint_organization 
		WHERE url = ANY($1)
	`, pq.Array(endpointURLs))
	if err != nil {
		log.Errorf("Failed to delete from endpoint_organization: %v", err)
	}

	_, err = store.DB.ExecContext(ctx, `
		DELETE FROM fhir_endpoints_availability 
		WHERE url = ANY($1)
	`, pq.Array(endpointURLs))
	if err != nil {
		log.Errorf("Failed to delete from fhir_endpoints_availability: %v", err)
	}

	result, err := store.DB.ExecContext(ctx, `
		DELETE FROM fhir_endpoints 
		WHERE list_source = $1
	`, listSource)
	if err != nil {
		return fmt.Errorf("failed to delete from fhir_endpoints: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Infof("Deleted %d endpoints from fhir_endpoints for list source: %s", rowsAffected, listSource)

	return nil
}
