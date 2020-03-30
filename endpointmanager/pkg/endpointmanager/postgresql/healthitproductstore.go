package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// prepared statements are left open to be used throughout the execution of the application
// TODO: figure out if there's a better way to manage this for bulk calls
var areHealthITProductStatementsPrepared = false
var addHealthITProductStatement *sql.Stmt
var updateHealthITProductStatement *sql.Stmt
var updateHealthITProductByNPIIDStatement *sql.Stmt
var deleteHealthITProductStatement *sql.Stmt
var linkHealthITProductToFHIREndpointStatement *sql.Stmt

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

// GetHealthITProductsUsingVendor returns a slice of HealthITProducts that were created by the given developer
func (s *Store) GetHealthITProductsUsingVendor(ctx context.Context, developer string) ([]*endpointmanager.HealthITProduct, error) {
	var hitps []*endpointmanager.HealthITProduct
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
	FROM healthit_products WHERE developer=$1`
	rows, err := s.DB.QueryContext(ctx, sqlStatement, developer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
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
		if err != nil {
			return nil, err
		}

		hitps = append(hitps, &hitp)
	}

	return hitps, nil
}

// GetHealthITProductDevelopers returns a list of all of the developers associated with the health IT products.
func (s *Store) GetHealthITProductDevelopers(ctx context.Context) ([]string, error) {
	var developers []string
	var developer string
	sqlStatement := "SELECT DISTINCT developer FROM healthit_products"
	rows, err := s.DB.QueryContext(ctx, sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&developer)
		if err != nil {
			return nil, err
		}
		developers = append(developers, developer)
	}

	return developers, nil
}

// AddHealthITProduct adds the HealthITProduct to the database.
func (s *Store) AddHealthITProduct(ctx context.Context, hitp *endpointmanager.HealthITProduct) error {
	err := prepareHealthITProductStatements(s)
	if err != nil {
		return err
	}

	locationJSON, err := json.Marshal(hitp.Location)
	if err != nil {
		return err
	}

	certificationCriteriaJSON, err := json.Marshal(hitp.CertificationCriteria)
	if err != nil {
		return err
	}

	row := addHealthITProductStatement.QueryRowContext(ctx,
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
	err := prepareHealthITProductStatements(s)
	if err != nil {
		return err
	}

	locationJSON, err := json.Marshal(hitp.Location)
	if err != nil {
		return err
	}

	certificationCriteriaJSON, err := json.Marshal(hitp.CertificationCriteria)
	if err != nil {
		return err
	}

	_, err = updateHealthITProductStatement.ExecContext(ctx,
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
	err := prepareHealthITProductStatements(s)
	if err != nil {
		return err
	}

	_, err = deleteHealthITProductStatement.ExecContext(ctx, hitp.ID)

	return err
}

func prepareHealthITProductStatements(s *Store) error {
	var err error
	if !areHealthITProductStatementsPrepared {
		areHealthITProductStatementsPrepared = true
		addHealthITProductStatement, err = s.DB.Prepare(`
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
		RETURNING id`)
		if err != nil {
			return err
		}
		updateHealthITProductStatement, err = s.DB.Prepare(`
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
		WHERE id=$14`)
		if err != nil {
			return err
		}
		deleteHealthITProductStatement, err = s.DB.Prepare(`
		DELETE FROM healthit_products
		WHERE id=$1`)
		if err != nil {
			return err
		}
	}
	return nil
}
