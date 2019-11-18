package postgresql

import (
	"context"
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// GetFHIREndpoint gets a FHIREndpoint from the database using the database id as a key.
// If the FHIREndpoint does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetFHIREndpoint(ctx context.Context, id int) (*endpointmanager.FHIREndpoint, error) {
	var endpoint endpointmanager.FHIREndpoint
	var locationJSON []byte
	var capabilityStatementJSON []byte

	sqlStatement := `
	SELECT
		id,
		url,
		fhir_version,
		authorization_standard,
		location,
		capability_statement,
		created_at,
		updated_at
	FROM fhir_endpoints WHERE id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&endpoint.ID,
		&endpoint.URL,
		&endpoint.FHIRVersion,
		&endpoint.AuthorizationStandard,
		&locationJSON,
		&capabilityStatementJSON,
		&endpoint.CreatedAt,
		&endpoint.UpdatedAt)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &endpoint.Location)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(capabilityStatementJSON, &endpoint.CapabilityStatement)

	return &endpoint, err
}

// GetFHIREndpointUsingURL gets a FHIREndpoint from the database using the given url as a key.
// If the FHIREndpoint does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetFHIREndpointUsingURL(ctx context.Context, url string) (*endpointmanager.FHIREndpoint, error) {
	var endpoint endpointmanager.FHIREndpoint
	var locationJSON []byte
	var capabilityStatementJSON []byte

	sqlStatement := `
	SELECT
		id,
		url,
		fhir_version,
		authorization_standard,
		location,
		capability_statement,
		created_at,
		updated_at
	FROM fhir_endpoints WHERE url=$1`

	row := s.DB.QueryRowContext(ctx, sqlStatement, url)

	err := row.Scan(
		&endpoint.ID,
		&endpoint.URL,
		&endpoint.FHIRVersion,
		&endpoint.AuthorizationStandard,
		&locationJSON,
		&capabilityStatementJSON,
		&endpoint.CreatedAt,
		&endpoint.UpdatedAt)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &endpoint.Location)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(capabilityStatementJSON, &endpoint.CapabilityStatement)

	return &endpoint, err
}

// AddFHIREndpoint adds the FHIREndpoint to the database.
func (s *Store) AddFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	sqlStatement := `
	INSERT INTO fhir_endpoints (url,
		fhir_version,
		authorization_standard,
		location,
		capability_statement)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id`

	locationJSON, err := json.Marshal(e.Location)
	if err != nil {
		return err
	}
	capabilityStatementJSON, err := json.Marshal(e.CapabilityStatement)
	if err != nil {
		return err
	}

	row := s.DB.QueryRowContext(ctx,
		sqlStatement,
		e.URL,
		e.FHIRVersion,
		e.AuthorizationStandard,
		locationJSON,
		capabilityStatementJSON)

	err = row.Scan(&e.ID)

	return err
}

// UpdateFHIREndpoint updates the FHIREndpoint in the database using the FHIREndpoint's database id as the key.
func (s *Store) UpdateFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	sqlStatement := `
	UPDATE fhir_endpoints
	SET url = $1,
		fhir_version = $2,
		authorization_standard = $3,
		location = $4,
		capability_statement = $5
	WHERE id = $6`

	locationJSON, err := json.Marshal(e.Location)
	if err != nil {
		return err
	}
	capabilityStatementJSON, err := json.Marshal(e.CapabilityStatement)
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx,
		sqlStatement,
		e.URL,
		e.FHIRVersion,
		e.AuthorizationStandard,
		locationJSON,
		capabilityStatementJSON,
		e.ID)

	return err
}

// DeleteFHIREndpoint deletes the FHIREndpoint from the database using the FHIREndpoint's database id  as the key.
func (s *Store) DeleteFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	sqlStatement := `
	DELETE FROM fhir_endpoints
	WHERE id = $1`

	_, err := s.DB.ExecContext(ctx, sqlStatement, e.ID)

	return err
}
