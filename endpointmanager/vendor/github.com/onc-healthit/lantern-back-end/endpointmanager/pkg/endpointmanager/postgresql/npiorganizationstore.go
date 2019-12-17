package postgresql

import (
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// GetNPIOrganizationByNPIID gets a NPIOrganization from the database using the NPI id as a key.
// If the NPIOrganization does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetNPIOrganizationByNPIID(npi_id string) (*endpointmanager.NPIOrganization, error) {
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
		created_at,
		updated_at
	FROM npi_organizations WHERE npi_id=$1`
	row := s.DB.QueryRow(sqlStatement, npi_id)

	err := row.Scan(
		&org.ID,
		&org.NPI_ID,
		&org.Name,
		&org.SecondaryName,
		&locationJSON,
		&org.Taxonomy,
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

// DeleteAllNPIOrganixations will remove all rows from the npi_organizations table
func (s *Store) DeleteAllNPIOrganizations() error {
	sqlStatement := `DELETE FROM npi_organizations`
	_, err := s.DB.Exec(sqlStatement)
	return err
}

// GetNPIOrganization gets a NPIOrganization from the database using the database id as a key.
// If the NPIOrganization does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetNPIOrganization(id int) (*endpointmanager.NPIOrganization, error) {
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
		created_at,
		updated_at
	FROM npi_organizations WHERE id=$1`
	row := s.DB.QueryRow(sqlStatement, id)

	err := row.Scan(
		&org.ID,
		&org.NPI_ID,
		&org.Name,
		&org.SecondaryName,
		&locationJSON,
		&org.Taxonomy,
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
func (s *Store) AddNPIOrganization(org *endpointmanager.NPIOrganization) error {
	sqlStatement := `
	INSERT INTO npi_organizations (
		npi_id,
		name,
		secondary_name,
		location,
		taxonomy)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id`

	locationJSON, err := json.Marshal(org.Location)
	if err != nil {
		return err
	}

	row := s.DB.QueryRow(sqlStatement,
		org.NPI_ID,
		org.Name,
		org.SecondaryName,
		locationJSON,
		org.Taxonomy)

	err = row.Scan(&org.ID)

	return err
}

// UpdateNPIOrganization updates the NPIOrganization in the database using the NPIOrganization's database ID as the key.
func (s *Store) UpdateNPIOrganization(org *endpointmanager.NPIOrganization) error {
	sqlStatement := `
	UPDATE npi_organizations
	SET npi_id = $2,
		name = $3,
		secondary_name = $4,
		location = $5,
		taxonomy = $6
	WHERE id=$1`

	locationJSON, err := json.Marshal(org.Location)
	if err != nil {
		return err
	}

	_, err = s.DB.Exec(sqlStatement,
		org.ID,
		org.NPI_ID,
		org.Name,
		org.SecondaryName,
		locationJSON,
		org.Taxonomy)

	return err
}

// UpdateNPIOrganizationByNPIID updates the NPIOrganization in the database using the NPIOrganization's NPIID as the key.
func (s *Store) UpdateNPIOrganizationByNPIID(org *endpointmanager.NPIOrganization) error {
	sqlStatement := `
	UPDATE npi_organizations
	SET name = $2,
		secondary_name = $3,
		location = $4,
		taxonomy = $5
	WHERE npi_id=$1`

	locationJSON, err := json.Marshal(org.Location)
	if err != nil {
		return err
	}

	_, err = s.DB.Exec(sqlStatement,
		org.NPI_ID,
		org.Name,
		org.SecondaryName,
		locationJSON,
		org.Taxonomy)

	return err
}

// DeleteNPIOrganization deletes the NPIOrganization from the database using the NPIOrganization's database ID as the key.
func (s *Store) DeleteNPIOrganization(org *endpointmanager.NPIOrganization) error {
	sqlStatement := `
	DELETE FROM npi_organizations
	WHERE id=$1`

	_, err := s.DB.Exec(sqlStatement, org.ID)

	return err
}
