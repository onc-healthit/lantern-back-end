package postgresql

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// prepared statements are left open to be used throughout the execution of the application
// TODO: figure out if there's a better way to manage this for bulk calls
var addFHIREndpointStatement *sql.Stmt
var updateFHIREndpointStatement *sql.Stmt
var deleteFHIREndpointStatement *sql.Stmt

// GetAllFHIREndpoints gets the id and url from every row in the fhir_endpoints table
func (s *Store) GetAllFHIREndpoints(ctx context.Context) ([]endpointmanager.FHIREndpoint, error) {
	sqlStatement := `
	SELECT
		id,
		url
	FROM fhir_endpoints`
	rows, err := s.DB.QueryContext(ctx, sqlStatement)
	if err != nil {
		return nil, err
	}

	var endpoints []endpointmanager.FHIREndpoint
	defer rows.Close()
	for rows.Next() {
		var endpoint endpointmanager.FHIREndpoint
		err = rows.Scan(
			&endpoint.ID,
			&endpoint.URL)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

// GetFHIREndpoint gets a FHIREndpoint from the database using the database id as a key.
// If the FHIREndpoint does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetFHIREndpoint(ctx context.Context, id int) (*endpointmanager.FHIREndpoint, error) {
	var endpoint endpointmanager.FHIREndpoint

	sqlStatement := `
	SELECT
		id,
		url,
		organization_names,
		npi_ids,
		list_source,
		created_at,
		updated_at
	FROM fhir_endpoints WHERE id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&endpoint.ID,
		&endpoint.URL,
		pq.Array(&endpoint.OrganizationNames),
		pq.Array(&endpoint.NPIIDs),
		&endpoint.ListSource,
		&endpoint.CreatedAt,
		&endpoint.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &endpoint, err
}

// GetFHIREndpointUsingURLAndListSource gets a FHIREndpoint from the database using the given url as a key.
// If the FHIREndpoint does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetFHIREndpointUsingURLAndListSource(ctx context.Context, url string, listSource string) (*endpointmanager.FHIREndpoint, error) {
	var endpoint endpointmanager.FHIREndpoint

	sqlStatement := `
	SELECT
		id,
		url,
		organization_names,
		npi_ids,
		list_source,
		created_at,
		updated_at
	FROM fhir_endpoints WHERE url=$1 AND list_source=$2`

	row := s.DB.QueryRowContext(ctx, sqlStatement, url, listSource)

	err := row.Scan(
		&endpoint.ID,
		&endpoint.URL,
		pq.Array(&endpoint.OrganizationNames),
		pq.Array(&endpoint.NPIIDs),
		&endpoint.ListSource,
		&endpoint.CreatedAt,
		&endpoint.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &endpoint, err
}

// AddFHIREndpoint adds the FHIREndpoint to the database.
func (s *Store) AddFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	var err error

	row := addFHIREndpointStatement.QueryRowContext(ctx,
		e.URL,
		pq.Array(e.OrganizationNames),
		pq.Array(e.NPIIDs),
		e.ListSource)

	err = row.Scan(&e.ID)

	return err
}

// UpdateFHIREndpoint updates the FHIREndpoint in the database using the FHIREndpoint's database id as the key.
func (s *Store) UpdateFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	var err error

	_, err = updateFHIREndpointStatement.ExecContext(ctx,
		e.URL,
		pq.Array(e.OrganizationNames),
		pq.Array(e.NPIIDs),
		e.ListSource,
		e.ID)

	return err
}

// DeleteFHIREndpoint deletes the FHIREndpoint from the database using the FHIREndpoint's database id  as the key.
func (s *Store) DeleteFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	_, err := deleteFHIREndpointStatement.ExecContext(ctx, e.ID)

	return err
}

// GetAllFHIREndpointOrgNames returns a sql.Rows of all of the orgNames
func (s *Store) GetAllFHIREndpointOrgNames(ctx context.Context) ([]endpointmanager.FHIREndpoint, error) {
	sqlStatement := `
        SELECT id, organization_names FROM fhir_endpoints`
	rows, err := s.DB.QueryContext(ctx, sqlStatement)

	if err != nil {
		return nil, err
	}
	var endpoints []endpointmanager.FHIREndpoint
	defer rows.Close()
	for rows.Next() {
		var endpoint endpointmanager.FHIREndpoint
		err = rows.Scan(&endpoint.ID, pq.Array(&endpoint.OrganizationNames))
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func prepareFHIREndpointStatements(s *Store) error {
	var err error
	addFHIREndpointStatement, err = s.DB.Prepare(`
		INSERT INTO fhir_endpoints (url,
			organization_names,
			npi_ids,
			list_source)
		VALUES ($1, $2, $3, $4)
		RETURNING id`)
	if err != nil {
		return err
	}
	updateFHIREndpointStatement, err = s.DB.Prepare(`
		UPDATE fhir_endpoints
		SET url = $1,
			organization_names = $2,
			npi_ids = $3,
			list_source = $4
		WHERE id = $5`)
	if err != nil {
		return err
	}
	deleteFHIREndpointStatement, err = s.DB.Prepare(`
        DELETE FROM fhir_endpoints
        WHERE id = $1`)
	if err != nil {
		return err
	}
	return nil
}
