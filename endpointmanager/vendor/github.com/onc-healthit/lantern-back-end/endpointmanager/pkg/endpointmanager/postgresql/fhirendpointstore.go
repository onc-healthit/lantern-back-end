package postgresql

import (
	"context"
	"encoding/json"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"

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
		tls_version,
		mime_types,
		http_response,
		errors,
		organization_name,
		fhir_version,
		authorization_standard,
		vendor,
		location,
		capability_statement,
		created_at,
		updated_at
	FROM fhir_endpoints WHERE id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&endpoint.ID,
		&endpoint.URL,
		&endpoint.TLSVersion,
		pq.Array(&endpoint.MIMETypes),
		&endpoint.HTTPResponse,
		&endpoint.Errors,
		&endpoint.OrganizationName,
		&endpoint.FHIRVersion,
		&endpoint.AuthorizationStandard,
		&endpoint.Vendor,
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
	if capabilityStatementJSON != nil {
		endpoint.CapabilityStatement, err = capabilityparser.NewCapabilityStatement(capabilityStatementJSON)
	}

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
		tls_version,
		mime_types,
		http_response,
		errors,
		organization_name,
		fhir_version,
		authorization_standard,
		vendor,
		location,
		capability_statement,
		created_at,
		updated_at
	FROM fhir_endpoints WHERE url=$1`

	row := s.DB.QueryRowContext(ctx, sqlStatement, url)

	err := row.Scan(
		&endpoint.ID,
		&endpoint.URL,
		&endpoint.TLSVersion,
		pq.Array(&endpoint.MIMETypes),
		&endpoint.HTTPResponse,
		&endpoint.Errors,
		&endpoint.OrganizationName,
		&endpoint.FHIRVersion,
		&endpoint.AuthorizationStandard,
		&endpoint.Vendor,
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
	if capabilityStatementJSON != nil {
		endpoint.CapabilityStatement, err = capabilityparser.NewCapabilityStatement(capabilityStatementJSON)
	}

	return &endpoint, err
}

// AddFHIREndpoint adds the FHIREndpoint to the database.
func (s *Store) AddFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	sqlStatement := `
	INSERT INTO fhir_endpoints (url,
		tls_version,
		mime_types,
		http_response,
		errors,
		organization_name,
		fhir_version,
		authorization_standard,
		vendor,
		location,
		capability_statement)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	RETURNING id`

	locationJSON, err := json.Marshal(e.Location)
	if err != nil {
		return err
	}
	var capabilityStatementJSON []byte
	if e.CapabilityStatement != nil {
		capabilityStatementJSON, err = e.CapabilityStatement.GetJSON()
		if err != nil {
			return err
		}
	} else {
		capabilityStatementJSON = []byte("null")
	}

	row := s.DB.QueryRowContext(ctx,
		sqlStatement,
		e.URL,
		e.TLSVersion,
		pq.Array(e.MIMETypes),
		e.HTTPResponse,
		e.Errors,
		e.OrganizationName,
		e.FHIRVersion,
		e.AuthorizationStandard,
		e.Vendor,
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
		tls_version = $2,
		mime_types = $3,
		http_response = $4,
		errors = $5,
		organization_name = $6,
		fhir_version = $7,
		authorization_standard = $8,
		vendor = $9,
		location = $10,
		capability_statement = $11
	WHERE id = $12`

	locationJSON, err := json.Marshal(e.Location)
	if err != nil {
		return err
	}
	var capabilityStatementJSON []byte
	if e.CapabilityStatement != nil {
		capabilityStatementJSON, err = e.CapabilityStatement.GetJSON()
		if err != nil {
			return err
		}
	} else {
		capabilityStatementJSON = []byte("null")
	}

	_, err = s.DB.ExecContext(ctx,
		sqlStatement,
		e.URL,
		e.TLSVersion,
		pq.Array(e.MIMETypes),
		e.HTTPResponse,
		e.Errors,
		e.OrganizationName,
		e.FHIRVersion,
		e.AuthorizationStandard,
		e.Vendor,
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
