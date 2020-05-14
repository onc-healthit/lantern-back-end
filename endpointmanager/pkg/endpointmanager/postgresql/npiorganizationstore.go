package postgresql

import (
	"context"
	"encoding/json"

	"database/sql"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// prepared statements are left open to be used throughout the execution of the application
// TODO: figure out if there's a better way to manage this for bulk calls
var addNPIOrganizationStatement *sql.Stmt
var updateNPIOrganizationStatement *sql.Stmt
var updateNPIOrganizationByNPIIDStatement *sql.Stmt
var deleteNPIOrganizationStatement *sql.Stmt
var linkNPIOrganizationToFHIREndpointStatement *sql.Stmt
var getNPIOrganizationFHIREndpointLinkStatement *sql.Stmt
var updateNPIOrganizationFHIREndpointLinkLink *sql.Stmt

// GetNPIOrganizationByNPIID gets a NPIOrganization from the database using the NPI id as a key.
// If the NPIOrganization does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetNPIOrganizationByNPIID(ctx context.Context, npiID string) (*endpointmanager.NPIOrganization, error) {
	var org endpointmanager.NPIOrganization
	var locationJSON []byte

	sqlStatement := `
	SELECT
		id,
		npi_id,
		names,
		location,
		taxonomy,
		normalized_names,
		created_at,
		updated_at
	FROM npi_organizations WHERE npi_id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, npiID)

	err := row.Scan(
		&org.ID,
		&org.NPI_ID,
		pq.Array(&org.Names),
		&locationJSON,
		&org.Taxonomy,
		pq.Array(&org.NormalizedNames),
		&org.CreatedAt,
		&org.UpdatedAt)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &org.Location)

	if err != nil {
		return nil, err
	}

	return &org, err
}

// DeleteAllNPIOrganizations will remove all rows from the npi_organizations table
func (s *Store) DeleteAllNPIOrganizations(ctx context.Context) error {
	sqlStatement := `DELETE FROM npi_organizations`
	_, err := s.DB.ExecContext(ctx, sqlStatement)
	return err
}

// GetNPIOrganization gets a NPIOrganization from the database using the database id as a key.
// If the NPIOrganization does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetNPIOrganization(ctx context.Context, id int) (*endpointmanager.NPIOrganization, error) {
	var org endpointmanager.NPIOrganization
	var locationJSON []byte

	sqlStatement := `
	SELECT
		id,
		npi_id,
		names,
		location,
		taxonomy,
		normalized_names,
		created_at,
		updated_at
	FROM npi_organizations WHERE id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&org.ID,
		&org.NPI_ID,
		pq.Array(&org.Names),
		&locationJSON,
		&org.Taxonomy,
		pq.Array(&org.NormalizedNames),
		&org.CreatedAt,
		&org.UpdatedAt)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &org.Location)

	if err != nil {
		return nil, err
	}

	return &org, err
}

// AddNPIOrganization adds the NPIOrganization to the database or updates if there is an existsing entry with same NPI_ID
func (s *Store) AddNPIOrganization(ctx context.Context, org *endpointmanager.NPIOrganization) error {
	locationJSON, err := json.Marshal(org.Location)
	if err != nil {
		return err
	}

	row := addNPIOrganizationStatement.QueryRowContext(ctx,
		//sqlStatement,
		org.NPI_ID,
		pq.Array(org.Names),
		locationJSON,
		org.Taxonomy,
		pq.Array(org.NormalizedNames))

	err = row.Scan(&org.ID)

	return err
}

// UpdateNPIOrganization updates the NPIOrganization in the database using the NPIOrganization's database ID as the key.
func (s *Store) UpdateNPIOrganization(ctx context.Context, org *endpointmanager.NPIOrganization) error {
	locationJSON, err := json.Marshal(org.Location)
	if err != nil {
		return err
	}

	_, err = updateNPIOrganizationStatement.ExecContext(ctx,
		org.ID,
		org.NPI_ID,
		pq.Array(org.Names),
		locationJSON,
		org.Taxonomy,
		pq.Array(org.NormalizedNames))

	return err
}

// UpdateNPIOrganizationByNPIID updates the NPIOrganization in the database using the NPIOrganization's NPIID as the key.
func (s *Store) UpdateNPIOrganizationByNPIID(ctx context.Context, org *endpointmanager.NPIOrganization) error {
	locationJSON, err := json.Marshal(org.Location)
	if err != nil {
		return err
	}

	_, err = updateNPIOrganizationByNPIIDStatement.ExecContext(ctx,
		org.NPI_ID,
		pq.Array(org.Names),
		locationJSON,
		org.Taxonomy,
		pq.Array(org.NormalizedNames))

	return err
}

// DeleteNPIOrganization deletes the NPIOrganization from the database using the NPIOrganization's database ID as the key.
func (s *Store) DeleteNPIOrganization(ctx context.Context, org *endpointmanager.NPIOrganization) error {
	_, err := deleteNPIOrganizationStatement.ExecContext(ctx, org.ID)

	return err
}

// GetAllNPIOrganizationNormalizedNames gets list of all primary and secondary names
func (s *Store) GetAllNPIOrganizationNormalizedNames(ctx context.Context) ([]*endpointmanager.NPIOrganization, error) {
	sqlStatement := `
	SELECT id, normalized_names, npi_id FROM npi_organizations`
	rows, err := s.DB.QueryContext(ctx, sqlStatement)
	if err != nil {
		return nil, err
	}
	var orgs []*endpointmanager.NPIOrganization
	defer rows.Close()
	for rows.Next() {
		var org endpointmanager.NPIOrganization
		err = rows.Scan(&org.ID, pq.Array(&org.NormalizedNames), &org.NPI_ID)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, &org)
	}
	return orgs, nil
}

// LinkNPIOrganizationToFHIREndpoint links an npi organization database id to a FHIR endpoint database id
func (s *Store) LinkNPIOrganizationToFHIREndpoint(ctx context.Context, orgID string, endpointURL string, confidence float64) error {
	_, err := linkNPIOrganizationToFHIREndpointStatement.ExecContext(ctx,
		orgID,
		endpointURL,
		confidence)
	return err
}

// GetNPIOrganizationFHIREndpointLink retrieves an existing link and its confidence level
func (s *Store) GetNPIOrganizationFHIREndpointLink(ctx context.Context, orgID string, endpointURL string) (int, string, float64, error) {
	var retOrgID int
	var retEndpointURL string
	var retConfidence float64

	row := getNPIOrganizationFHIREndpointLinkStatement.QueryRowContext(ctx,
		orgID,
		endpointURL)

	err := row.Scan(
		&retOrgID,
		&retEndpointURL,
		&retConfidence,
	)

	return retOrgID, retEndpointURL, retConfidence, err
}

// UpdateNPIOrganizationFHIREndpointLink updates an existing link with a new confidence level
func (s *Store) UpdateNPIOrganizationFHIREndpointLink(ctx context.Context, orgID string, endpointURL string, confidence float64) error {
	_, err := updateNPIOrganizationFHIREndpointLinkLink.ExecContext(ctx,
		orgID,
		endpointURL,
		confidence)
	return err
}

func prepareNPIOrganizationStatements(s *Store) error {
	var err error
	addNPIOrganizationStatement, err = s.DB.Prepare(`
		INSERT INTO npi_organizations (
			npi_id,
			names,
			location,
			taxonomy,
			normalized_names)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`)
	if err != nil {
		return err
	}
	updateNPIOrganizationStatement, err = s.DB.Prepare(`
		UPDATE npi_organizations
		SET npi_id = $2,
		        names = $3,
		        location = $4,
		        taxonomy = $5,
		        normalized_names = $6
		WHERE id=$1`)
	if err != nil {
		return err
	}
	updateNPIOrganizationByNPIIDStatement, err = s.DB.Prepare(`
		UPDATE npi_organizations
		SET names = $2,
			location = $3,
			taxonomy = $4,
			normalized_names = $5
		WHERE npi_id=$1`)
	if err != nil {
		return err
	}
	deleteNPIOrganizationStatement, err = s.DB.Prepare(`
		DELETE FROM npi_organizations
		WHERE id=$1`)
	if err != nil {
		return err
	}
	linkNPIOrganizationToFHIREndpointStatement, err = s.DB.Prepare(`
		INSERT INTO endpoint_organization (
			organization_npi_id,
			url,
			confidence)
		VALUES ($1, $2, $3)`)
	if err != nil {
		return err
	}
	getNPIOrganizationFHIREndpointLinkStatement, err = s.DB.Prepare(`
		SELECT
			organization_npi_id,
			url,
			confidence
		FROM endpoint_organization
		WHERE organization_npi_id=$1 AND url=$2
	`)
	if err != nil {
		return err
	}
	updateNPIOrganizationFHIREndpointLinkLink, err = s.DB.Prepare(`
		UPDATE endpoint_organization
		SET confidence = $3
		WHERE organization_npi_id = $1 AND url = $2`)
	if err != nil {
		return err
	}
	return nil
}
