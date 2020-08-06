package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// prepared statements are left open to be used throughout the execution of the application
// TODO: figure out if there's a better way to manage this for bulk calls
var addHealthITProductStatement *sql.Stmt
var updateHealthITProductStatement *sql.Stmt
var deleteHealthITProductStatement *sql.Stmt
var getProductCriteriaLinkStatement *sql.Stmt
var linkProductToCriteriaStatement *sql.Stmt

// GetHealthITProduct gets a HealthITProduct from the database using the database ID as a key.
// If the HealthITProduct does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetHealthITProduct(ctx context.Context, id int) (*endpointmanager.HealthITProduct, error) {
	var hitp endpointmanager.HealthITProduct
	var locationJSON []byte
	var certificationCriteriaJSON []byte
	var vendorIDNullable sql.NullInt64

	sqlStatement := `
	SELECT
		id,
		name,
		version,
		vendor_id,
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
		&vendorIDNullable,
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

	ints := getRegularInts([]sql.NullInt64{vendorIDNullable})
	hitp.VendorID = ints[0]

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
	var vendorIDNullable sql.NullInt64

	sqlStatement := `
	SELECT
		id,
		name,
		version,
		vendor_id,
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
		&vendorIDNullable,
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

	ints := getRegularInts([]sql.NullInt64{vendorIDNullable})
	hitp.VendorID = ints[0]

	err = json.Unmarshal(locationJSON, &hitp.Location)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(certificationCriteriaJSON, &hitp.CertificationCriteria)

	return &hitp, err
}

// GetHealthITProductsUsingVendor returns a slice of HealthITProducts that were created by the given vendor_id
func (s *Store) GetHealthITProductsUsingVendor(ctx context.Context, vendorID int) ([]*endpointmanager.HealthITProduct, error) {
	var hitps []*endpointmanager.HealthITProduct
	var hitp endpointmanager.HealthITProduct
	var locationJSON []byte
	var certificationCriteriaJSON []byte
	var vendorIDNullable sql.NullInt64

	sqlStatement := `
	SELECT
		id,
		name,
		version,
		vendor_id,
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
	FROM healthit_products WHERE vendor_id=$1`
	rows, err := s.DB.QueryContext(ctx, sqlStatement, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&hitp.ID,
			&hitp.Name,
			&hitp.Version,
			&vendorIDNullable,
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

		ints := getRegularInts([]sql.NullInt64{vendorIDNullable})
		hitp.VendorID = ints[0]

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

// AddHealthITProduct adds the HealthITProduct to the database.
func (s *Store) AddHealthITProduct(ctx context.Context, hitp *endpointmanager.HealthITProduct) error {
	locationJSON, err := json.Marshal(hitp.Location)
	if err != nil {
		return err
	}

	certificationCriteriaJSON, err := json.Marshal(hitp.CertificationCriteria)
	if err != nil {
		return err
	}

	nullableInts := getNullableInts([]int{hitp.VendorID})

	row := addHealthITProductStatement.QueryRowContext(ctx,
		hitp.Name,
		hitp.Version,
		nullableInts[0],
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
	locationJSON, err := json.Marshal(hitp.Location)
	if err != nil {
		return err
	}

	certificationCriteriaJSON, err := json.Marshal(hitp.CertificationCriteria)
	if err != nil {
		return err
	}

	nullableInts := getNullableInts([]int{hitp.VendorID})

	_, err = updateHealthITProductStatement.ExecContext(ctx,
		hitp.Name,
		hitp.Version,
		nullableInts[0],
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
	_, err := deleteHealthITProductStatement.ExecContext(ctx, hitp.ID)

	return err
}

// GetProductCriteriaLink retrieves the product database id, criteria id, and criteria number for the requested
// product db id and criteria id. If the link doesn't exist, returns a SQL no rows error.
func (s *Store) GetProductCriteriaLink(ctx context.Context, criteriaID int, productID int) (int, int, string, error) {
	var retProductID int
	var retCriteriaID int
	var retCriteriaNumber string

	row := getProductCriteriaLinkStatement.QueryRowContext(ctx,
		productID,
		criteriaID)

	err := row.Scan(
		&retProductID,
		&retCriteriaID,
		&retCriteriaNumber,
	)

	return retProductID, retCriteriaID, retCriteriaNumber, err
}

// LinkProductToCriteria links a product database id to a certification criteria id
func (s *Store) LinkProductToCriteria(ctx context.Context, criteriaID int, productID int, productNumber string) error {
	_, err := linkProductToCriteriaStatement.ExecContext(ctx,
		productID,
		criteriaID,
		productNumber)
	return err
}

// DeleteLinksByProduct deletes all of the links in product_criteria with the given health it product database id
func (s *Store) DeleteLinksByProduct(ctx context.Context, productID int) error {
	sqlStatement := `DELETE FROM product_criteria WHERE healthit_product_id=$1`
	_, err := s.DB.ExecContext(ctx, sqlStatement, productID)
	return err
}

func prepareHealthITProductStatements(s *Store) error {
	var err error
	addHealthITProductStatement, err = s.DB.Prepare(`
		INSERT INTO healthit_products (
			name,
			version,
			vendor_id,
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
			vendor_id = $3,
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
	getProductCriteriaLinkStatement, err = s.DB.Prepare(`
		SELECT
			healthit_product_id,
			certification_id,
			certification_number
		FROM product_criteria
		WHERE healthit_product_id=$1 AND certification_id=$2
	`)
	if err != nil {
		return err
	}
	linkProductToCriteriaStatement, err = s.DB.Prepare(`
		INSERT INTO product_criteria (
			healthit_product_id,
			certification_id,
			certification_number)
		VALUES ($1, $2, $3)`)
	if err != nil {
		return err
	}
	return nil
}
