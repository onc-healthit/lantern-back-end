package postgresql

import (
	"context"
	"database/sql"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// prepared statements are left open to be used throughout the execution of the application
var addFHIREndpointMetadataStatement *sql.Stmt

// GetFHIREndpointMetadata gets a FHIREndpointMetadata from the database using the metadata id as a key.
// If the FHIREndpointMetadata does not exist in the database, sql.ErrNoRows will be returned.
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
		requested_fhir_version,
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
		&endpointMetadata.RequestedFhirVersion,
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
		e.SMARTHTTPResponse,
		e.RequestedFhirVersion)

	err = row.Scan(&metadataID)

	return metadataID, err
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
			smart_http_response,
			requested_fhir_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`)
	return err
}
