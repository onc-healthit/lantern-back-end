package postgresql

import (
	"context"
	"database/sql"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// prepared statements are left open to be used throughout the execution of the application
var addFHIREndpointMetadataStatement *sql.Stmt
var updateFHIREndpointInfoMetadataStatement *sql.Stmt

// GetFHIREndpointMetadata gets a FHIREndpointMetadata from the database using the metadata id as a key.
// If the FHIREndpointInfo does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetFHIREndpointMetadata(ctx context.Context, metadataID int) (*endpointmanager.FHIREndpointMetadata, error) {
	var endpointMetadata endpointmanager.FHIREndpointMetadata
	endpointMetadata.ID = metadataID

	sqlStatementMetadata := `
	SELECT
		url,
		http_response,
		availability,
		errors,
		response_time_seconds,
		smart_http_response,
		updated_at,
		created_at 
	FROM fhir_endpoints_metadata WHERE id=$1;`

	row := s.DB.QueryRowContext(ctx, sqlStatementMetadata, metadataID)

	err := row.Scan(
		&endpointMetadata.URL,
		&endpointMetadata.HTTPResponse,
		&endpointMetadata.Availability,
		&endpointMetadata.Errors,
		&endpointMetadata.ResponseTime,
		&endpointMetadata.SMARTHTTPResponse,
		&endpointMetadata.UpdatedAt,
		&endpointMetadata.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &endpointMetadata, err
}

// AddFHIREndpointMetadata adds the FHIREndpointMetadata in the database
func (s *Store) AddFHIREndpointMetadata(ctx context.Context, e *endpointmanager.FHIREndpointMetadata) (int, error) {
	var err error
	var metadataID int

	row := addFHIREndpointMetadataStatement.QueryRowContext(ctx,
		e.URL,
		e.HTTPResponse,
		e.Availability,
		e.Errors,
		e.ResponseTime,
		e.SMARTHTTPResponse)

	err = row.Scan(&metadataID)

	return metadataID, err
}

// UpdateMetadataIDInfo only updates the metadata_id in the info table without affecting the info history table
func (s *Store) UpdateMetadataIDInfo(ctx context.Context, metadataID int, url string) error {
	infoHistoryTriggerDisable := `
	ALTER TABLE fhir_endpoints_info
	DISABLE TRIGGER add_fhir_endpoint_info_history_trigger;`

	infoHistoryTriggerEnable := `
	ALTER TABLE fhir_endpoints_info
	ENABLE TRIGGER add_fhir_endpoint_info_history_trigger;`

	timestampTriggerDisable := `
	ALTER TABLE fhir_endpoints_info
	DISABLE TRIGGER set_timestamp_fhir_endpoints_info;`

	timestampTriggerEnable := `
	ALTER TABLE fhir_endpoints_info
	ENABLE TRIGGER set_timestamp_fhir_endpoints_info;`

	_, err := s.DB.ExecContext(ctx, infoHistoryTriggerDisable)
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx, timestampTriggerDisable)
	if err != nil {
		return err
	}

	_, err = updateFHIREndpointInfoMetadataStatement.ExecContext(ctx, metadataID, url)
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx, infoHistoryTriggerEnable)
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx, timestampTriggerEnable)

	return err
}

func prepareFHIREndpointMetadataStatements(s *Store) error {
	var err error
	addFHIREndpointMetadataStatement, err = s.DB.Prepare(`
		INSERT INTO fhir_endpoints_metadata (
			url,
			http_response,
			availability,
			errors,
			response_time_seconds,
			smart_http_response)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`)
	if err != nil {
		return err
	}
	updateFHIREndpointInfoMetadataStatement, err = s.DB.Prepare(`
		UPDATE fhir_endpoints_info
		SET 
			metadata_id = $1		
		WHERE url = $2`)
	if err != nil {
		return err
	}
	return nil
}
