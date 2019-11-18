package postgresql

import (
	"context"
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// GetProviderOrganization gets a ProviderOrganization from the database using the database id as a key.
// If the ProviderOrganization does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetProviderOrganization(ctx context.Context, id int) (*endpointmanager.ProviderOrganization, error) {
	var po endpointmanager.ProviderOrganization
	var locationJSON []byte

	sqlStatement := `
	SELECT
		id,
		name,
		url,
		location,
		organization_type,
		hospital_type,
		ownership,
		beds,
		created_at,
		updated_at
	FROM provider_organizations WHERE id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&po.ID,
		&po.Name,
		&po.URL,
		&locationJSON,
		&po.OrganizationType,
		&po.HospitalType,
		&po.Ownership,
		&po.Beds,
		&po.CreatedAt,
		&po.UpdatedAt)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &po.Location)

	return &po, err
}

// AddProviderOrganization adds the ProviderOrganization to the database.
func (s *Store) AddProviderOrganization(ctx context.Context, po *endpointmanager.ProviderOrganization) error {
	sqlStatement := `
	INSERT INTO provider_organizations (
		name,
		url,
		location,
		organization_type,
		hospital_type,
		ownership,
		beds)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id`

	locationJSON, err := json.Marshal(po.Location)
	if err != nil {
		return err
	}

	row := s.DB.QueryRowContext(ctx,
		sqlStatement,
		po.Name,
		po.URL,
		locationJSON,
		po.OrganizationType,
		po.HospitalType,
		po.Ownership,
		po.Beds)

	err = row.Scan(&po.ID)

	return err
}

// UpdateProviderOrganization updates the ProviderOrganization in the database using the ProviderOrganization's database ID as the key.
func (s *Store) UpdateProviderOrganization(ctx context.Context, po *endpointmanager.ProviderOrganization) error {
	sqlStatement := `
	UPDATE provider_organizations
	SET name = $2,
		url = $3,
		organization_type = $4,
		hospital_type = $5,
		ownership = $6,
		beds = $7,
		location = $8
	WHERE id = $1`

	locationJSON, err := json.Marshal(po.Location)
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx,
		sqlStatement,
		po.ID,
		po.Name,
		po.URL,
		po.OrganizationType,
		po.HospitalType,
		po.Ownership,
		po.Beds,
		locationJSON)

	return err
}

// DeleteProviderOrganization deletes the ProviderOrganization from the database using the ProviderOrganization's database ID as the key.
func (s *Store) DeleteProviderOrganization(ctx context.Context, po *endpointmanager.ProviderOrganization) error {
	sqlStatement := `
	DELETE FROM provider_organizations
	WHERE id=$1`

	_, err := s.DB.ExecContext(ctx, sqlStatement, po.ID)

	return err
}
