package postgresql

import (
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// GetNPIOrganization gets a NPIOrganization from the database using the database id as a key.
// If the NPIOrganization does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetNPIOrganization(id int) (*endpointmanager.NPIOrganization, error) {
	var po endpointmanager.NPIOrganization
	var locationJSON []byte
	var fhirEndpointJSON []byte

	sqlStatement := `
	SELECT
		id,
		npi_id,
		name,
		secondary_name,
		fhir_endpoint,
		location,
		taxonomy,
		created_at,
		updated_at
	FROM npi_organizations WHERE id=$1`
	row := s.DB.QueryRow(sqlStatement, id)

	err := row.Scan(
		&po.ID,
		&po.NPI_ID,
		&po.Name,
		&po.SecondaryName,
		&fhirEndpointJSON,
		&locationJSON,
		&po.Taxonomy,
		&po.CreatedAt,
		&po.UpdatedAt)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &po.Location)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(fhirEndpointJSON, &po.FHIREndpoint)

	if err != nil {
		return nil, err
	}

	return &po, err
}

// AddNPIOrganization adds the NPIOrganization to the database.
func (s *Store) AddNPIOrganization(po *endpointmanager.NPIOrganization) error {
	sqlStatement := `
	INSERT INTO npi_organizations (
		npi_id,
		name,
		secondary_name,
		fhir_endpoint,
		location,
		taxonomy)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id`

	locationJSON, err := json.Marshal(po.Location)
	if err != nil {
		return err
	}
	fhirEndpointJSON, err := json.Marshal(po.FHIREndpoint)
	if err != nil {
		return err
	}

	row := s.DB.QueryRow(sqlStatement,
		po.NPI_ID,
		po.Name,
		po.SecondaryName,
		fhirEndpointJSON,
		locationJSON,
		po.Taxonomy)

	err = row.Scan(&po.ID)

	return err
}

// UpdateNPIOrganization updates the NPIOrganization in the database using the NPIOrganization's database ID as the key.
func (s *Store) UpdateNPIOrganization(po *endpointmanager.NPIOrganization) error {
	sqlStatement := `
	UPDATE npi_organizations
	SET npi_id = $2,
		name = $3,
		secondary_name = $4,
		fhir_endpoint = $5,
		location = $6,
		taxonomy = $7,
	WHERE id = $1`

	locationJSON, err := json.Marshal(po.Location)
	if err != nil {
		return err
	}
	fhirEndpointJSON, err := json.Marshal(po.FHIREndpoint)
	if err != nil {
		return err
	}

	_, err = s.DB.Exec(sqlStatement,
		po.NPI_ID,
		po.Name,
		po.SecondaryName,
		fhirEndpointJSON,
		locationJSON,
		po.Taxonomy)

	return err
}

// DeleteNPIOrganization deletes the NPIOrganization from the database using the NPIOrganization's database ID as the key.
func (s *Store) DeleteNPIOrganization(po *endpointmanager.NPIOrganization) error {
	sqlStatement := `
	DELETE FROM npi_organizations
	WHERE id=$1`

	_, err := s.DB.Exec(sqlStatement, po.ID)

	return err
}
