package postgresql

import (
	"context"
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// GetNPIOrganizationByNPIID gets a NPIOrganization from the database using the NPI id as a key.
// If the NPIOrganization does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetNPIOrganizationByNPIID(ctx context.Context, npi_id string) (*endpointmanager.NPIOrganization, error) {
	var org endpointmanager.NPIOrganization
	var locationJSON []byte

	sqlStatement := `
	SELECT
		id,
		npi_id,
		name,
		secondary_name,
		location,
		taxonomy,
		normalized_name,
		normalized_secondary_name,
		created_at,
		updated_at
	FROM npi_organizations WHERE npi_id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, npi_id)

	err := row.Scan(
		&org.ID,
		&org.NPI_ID,
		&org.Name,
		&org.SecondaryName,
		&locationJSON,
		&org.Taxonomy,
		&org.NormalizedName,
		&org.NormalizedSecondaryName,
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
		name,
		secondary_name,
		location,
		taxonomy,
		normalized_name,
		normalized_secondary_name,
		created_at,
		updated_at
	FROM npi_organizations WHERE id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&org.ID,
		&org.NPI_ID,
		&org.Name,
		&org.SecondaryName,
		&locationJSON,
		&org.Taxonomy,
		&org.NormalizedName,
		&org.NormalizedSecondaryName,
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
	sqlStatement := `
	INSERT INTO npi_organizations (
		npi_id,
		name,
		secondary_name,
		location,
		taxonomy,
		normalized_name,
		normalized_secondary_name)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id`

	locationJSON, err := json.Marshal(org.Location)
	if err != nil {
		return err
	}

	row := s.DB.QueryRowContext(ctx,
		sqlStatement,
		org.NPI_ID,
		org.Name,
		org.SecondaryName,
		locationJSON,
		org.Taxonomy,
		org.NormalizedName,
		org.NormalizedSecondaryName)

	err = row.Scan(&org.ID)

	return err
}

// UpdateNPIOrganization updates the NPIOrganization in the database using the NPIOrganization's database ID as the key.
func (s *Store) UpdateNPIOrganization(ctx context.Context, org *endpointmanager.NPIOrganization) error {
	sqlStatement := `
	UPDATE npi_organizations
	SET npi_id = $2,
		name = $3,
		secondary_name = $4,
		location = $5,
		taxonomy = $6,
		normalized_name = $7,
		normalized_secondary_name = $8
	WHERE id=$1`

	locationJSON, err := json.Marshal(org.Location)
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx,
		sqlStatement,
		org.ID,
		org.NPI_ID,
		org.Name,
		org.SecondaryName,
		locationJSON,
		org.Taxonomy,
		org.NormalizedName,
		org.NormalizedSecondaryName)

	return err
}

// UpdateNPIOrganizationByNPIID updates the NPIOrganization in the database using the NPIOrganization's NPIID as the key.
func (s *Store) UpdateNPIOrganizationByNPIID(ctx context.Context, org *endpointmanager.NPIOrganization) error {
	sqlStatement := `
	UPDATE npi_organizations
	SET name = $2,
		secondary_name = $3,
		location = $4,
		taxonomy = $5,
		normalized_name = $6,
		normalized_secondary_name = $7
	WHERE npi_id=$1`

	locationJSON, err := json.Marshal(org.Location)
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx,
		sqlStatement,
		org.NPI_ID,
		org.Name,
		org.SecondaryName,
		locationJSON,
		org.Taxonomy,
		org.NormalizedName,
		org.NormalizedSecondaryName)

	return err
}

// DeleteNPIOrganization deletes the NPIOrganization from the database using the NPIOrganization's database ID as the key.
func (s *Store) DeleteNPIOrganization(ctx context.Context, org *endpointmanager.NPIOrganization) error {
	sqlStatement := `
	DELETE FROM npi_organizations
	WHERE id=$1`

	_, err := s.DB.ExecContext(ctx, sqlStatement, org.ID)

	return err
}

// GetAllNPIOrganizationNormalizedNames gets list of all primary and secondary names
func (s *Store) GetAllNPIOrganizationNormalizedNames(ctx context.Context) ([]endpointmanager.NPIOrganization, error) {
	sqlStatement := `
	SELECT id, normalized_name, normalized_secondary_name FROM npi_organizations`
	rows, err := s.DB.QueryContext(ctx, sqlStatement)
	if err != nil {
		return nil, err
	}
	var orgs []endpointmanager.NPIOrganization
	defer rows.Close()
	for rows.Next() {
		var org endpointmanager.NPIOrganization
		err = rows.Scan(&org.ID, &org.NormalizedName, &org.NormalizedSecondaryName)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

// LinkNPIOrganizationToFHIREndpoint links an npi organization database id to a FHIR endpoint database id
func (s *Store) LinkNPIOrganizationToFHIREndpoint(ctx context.Context, orgId int, endpointId int, confidence float64) error {
	sqlStatement := `
	INSERT INTO endpoint_organization (
		organization_id,
		endpoint_id,
		confidence)
	VALUES ($1, $2, $3)`

	_, err := s.DB.ExecContext(ctx,
		sqlStatement,
		orgId,
		endpointId,
		confidence)
	return err
}
