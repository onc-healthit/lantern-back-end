package postgresql

import (
	"context"
	"database/sql"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

var addFHIREndpointAvailabilityStatement *sql.Stmt
var updateFHIREndpointAvailabilityStatement *sql.Stmt
var deleteFHIREndpointAvailabilityStatement *sql.Stmt

// GetFHIREndpointAvailabilityUsingURL gets the FHIREndpointAvailability object that corresponds to the FHIREndpoint with the given URL.
func (s *Store) GetFHIREndpointAvailabilityUsingURL(ctx context.Context, url string) (*endpointmanager.FHIREndpointAvailability, error) {
	var endpointAvailability endpointmanager.FHIREndpointAvailability

	sqlStatement := `
	SELECT
		url,
		http_200_count,
		http_all_count
	FROM fhir_endpoint_availability WHERE fhir_endpoint_availability.url = $1`

	row := s.DB.QueryRowContext(ctx, sqlStatement, url)

	err := row.Scan(
		&endpointAvailability.URL,
		&endpointAvailability.HTTP_200_COUNT,
		&endpointAvailability.HTTP_ALL_COUNT)
	if err != nil {
		return nil, err
	}

	return &endpointAvailability, err
}

// AddFHIREndpointAvailability adds the FHIREndpointAvailability to the database.
func (s *Store) AddFHIREndpointAvailability(ctx context.Context, e *endpointmanager.FHIREndpointAvailability) error {

	_, err := addFHIREndpointAvailabilityStatement.ExecContext(ctx,
		e.URL,
		e.HTTP_200_COUNT,
		e.HTTP_ALL_COUNT)

	return err
}

// UpdateFHIREndpointAvailability updates the FHIREndpointAvailability in the database using the url as the key.
func (s *Store) UpdateFHIREndpointAvailability(ctx context.Context, e *endpointmanager.FHIREndpointAvailability) error {
	var err error

	_, err = updateFHIREndpointAvailabilityStatement.ExecContext(ctx,
		e.URL,
		e.HTTP_200_COUNT,
		e.HTTP_ALL_COUNT)

	return err
}

// DeleteFHIREndpointAvailability deletes the FHIREndpointAvailability from the database using the url as the key.
func (s *Store) DeleteFHIREndpointAvailability(ctx context.Context, e *endpointmanager.FHIREndpointAvailability) error {
	_, err := deleteFHIREndpointAvailabilityStatement.ExecContext(ctx, e.URL)

	return err
}

func prepareFHIREndpointAvailabilityStatements(s *Store) error {
	var err error
	addFHIREndpointAvailabilityStatement, err = s.DB.Prepare(`
		INSERT INTO fhir_endpoint_availability (
			url,
			http_200_count,
			http_all_count)
		VALUES ($1, $2, $3)`)
	if err != nil {
		return err
	}
	updateFHIREndpointAvailabilityStatement, err = s.DB.Prepare(`
		UPDATE fhir_endpoint_availability
		SET 
		    http_200_count = $2,
			http_all_count = $3
		WHERE url = $1`)
	if err != nil {
		return err
	}
	deleteFHIREndpointAvailabilityStatement, err = s.DB.Prepare(`
        DELETE FROM fhir_endpoint_availability
        WHERE url = $1`)
	if err != nil {
		return err
	}
	return nil
}