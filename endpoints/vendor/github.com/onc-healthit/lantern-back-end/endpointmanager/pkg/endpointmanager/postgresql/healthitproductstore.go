package postgresql

import (
	"context"
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// GetHealthITProduct gets a HealthITProduct from the database using the database ID as a key.
// If the HealthITProduct does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetHealthITProduct(ctx context.Context, id int) (*endpointmanager.HealthITProduct, error) {
	var hitp endpointmanager.HealthITProduct
	var locationJSON []byte
	var certificationCriteriaJSON []byte

	sqlStatement := `
	SELECT
		id,
		name,
		version,
		developer,
		location,
		authorization_standard,
		api_syntax,
		api_url,
		certification_criteria,
		certification_status,
		certification_date,
		certification_edition,
		last_modified_in_chpl,
		chpl_id,
		created_at,
		updated_at
	FROM healthit_products WHERE id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&hitp.ID,
		&hitp.Name,
		&hitp.Version,
		&hitp.Developer,
		&locationJSON,
		&hitp.AuthorizationStandard,
		&hitp.APISyntax,
		&hitp.APIURL,
		&certificationCriteriaJSON,
		&hitp.CertificationStatus,
		&hitp.CertificationDate,
		&hitp.CertificationEdition,
		&hitp.LastModifiedInCHPL,
		&hitp.CHPLID,
		&hitp.CreatedAt,
		&hitp.UpdatedAt)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &hitp.Location)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(certificationCriteriaJSON, &hitp.CertificationCriteria)

	return &hitp, err
}

// GetHealthITProductUsingNameAndVersion gets a HealthITProduct from the database using the healthit product's name and version as a key.
// If the HealthITProduct does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetHealthITProductUsingNameAndVersion(ctx context.Context, name string, version string) (*endpointmanager.HealthITProduct, error) {
	var hitp endpointmanager.HealthITProduct
	var locationJSON []byte
	var certificationCriteriaJSON []byte

	sqlStatement := `
	SELECT
		id,
		name,
		version,
		developer,
		location,
		authorization_standard,
		api_syntax,
		api_url,
		certification_criteria,
		certification_status,
		certification_date,
		certification_edition,
		last_modified_in_chpl,
		chpl_id,
		created_at,
		updated_at
	FROM healthit_products WHERE name=$1 AND version=$2`
	row := s.DB.QueryRowContext(ctx, sqlStatement, name, version)

	err := row.Scan(
		&hitp.ID,
		&hitp.Name,
		&hitp.Version,
		&hitp.Developer,
		&locationJSON,
		&hitp.AuthorizationStandard,
		&hitp.APISyntax,
		&hitp.APIURL,
		&certificationCriteriaJSON,
		&hitp.CertificationStatus,
		&hitp.CertificationDate,
		&hitp.CertificationEdition,
		&hitp.LastModifiedInCHPL,
		&hitp.CHPLID,
		&hitp.CreatedAt,
		&hitp.UpdatedAt)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &hitp.Location)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(certificationCriteriaJSON, &hitp.CertificationCriteria)

	return &hitp, err
}

// AddHealthITProduct adds the HealthITProduct to the database.
func (s *Store) AddHealthITProduct(ctx context.Context, hitp *endpointmanager.HealthITProduct) error {
	sqlStatement := `
	INSERT INTO healthit_products (
		name,
		version,
		developer,
		location,
		authorization_standard,
		api_syntax,
		api_url,
		certification_criteria,
		certification_status,
		certification_date,
		certification_edition,
		last_modified_in_chpl,
		chpl_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	RETURNING id`

	locationJSON, err := json.Marshal(hitp.Location)
	if err != nil {
		return err
	}

	certificationCriteriaJSON, err := json.Marshal(hitp.CertificationCriteria)
	if err != nil {
		return err
	}

	row := s.DB.QueryRowContext(ctx,
		sqlStatement,
		hitp.Name,
		hitp.Version,
		hitp.Developer,
		locationJSON,
		hitp.AuthorizationStandard,
		hitp.APISyntax,
		hitp.APIURL,
		certificationCriteriaJSON,
		hitp.CertificationStatus,
		hitp.CertificationDate,
		hitp.CertificationEdition,
		hitp.LastModifiedInCHPL,
		hitp.CHPLID)

	err = row.Scan(&hitp.ID)

	return err
}

// UpdateHealthITProduct updates the HealthITProduct in the database using the HealthITProduct's database ID as the key.
func (s *Store) UpdateHealthITProduct(ctx context.Context, hitp *endpointmanager.HealthITProduct) error {
	sqlStatement := `
	UPDATE healthit_products
	SET name = $1,
		version = $2,
		developer = $3,
		authorization_standard = $4,
		api_syntax = $5,
		api_url = $6,
		certification_status = $7,
		certification_date = $8,
		certification_edition = $9,
		last_modified_in_chpl = $10,
		chpl_id = $11,
		location = $12,
		certification_criteria = $13
	WHERE id=$14`

	locationJSON, err := json.Marshal(hitp.Location)
	if err != nil {
		return err
	}

	certificationCriteriaJSON, err := json.Marshal(hitp.CertificationCriteria)
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx,
		sqlStatement,
		hitp.Name,
		hitp.Version,
		hitp.Developer,
		hitp.AuthorizationStandard,
		hitp.APISyntax,
		hitp.APIURL,
		hitp.CertificationStatus,
		hitp.CertificationDate,
		hitp.CertificationEdition,
		hitp.LastModifiedInCHPL,
		hitp.CHPLID,
		locationJSON,
		certificationCriteriaJSON,
		hitp.ID)

	return err
}

// DeleteHealthITProduct deletes the HealthITProduct from the database using the HealthITProduct's database ID as the key.
func (s *Store) DeleteHealthITProduct(ctx context.Context, hitp *endpointmanager.HealthITProduct) error {
	sqlStatement := `
	DELETE FROM healthit_products
	WHERE id=$1`

	_, err := s.DB.ExecContext(ctx, sqlStatement, hitp.ID)

	return err
}
