package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// prepared statements are left open to be used throughout the execution of the application
var addVendorStatement *sql.Stmt
var updateVendorStatement *sql.Stmt
var deleteVendorStatement *sql.Stmt

// GetVendor gets a Vendor from the database using the database id as a key.
// If the Vendor does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetVendor(ctx context.Context, id int) (*endpointmanager.Vendor, error) {
	var vendor endpointmanager.Vendor
	var locationJSON []byte

	sqlStatement := `
	SELECT
		id,
		name,
		developer_code,
		url,
		location,
		status,
		last_modified_in_chpl,
		chpl_id,
		created_at,
		updated_at
	FROM vendors WHERE id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&vendor.ID,
		&vendor.Name,
		&vendor.DeveloperCode,
		&vendor.URL,
		&locationJSON,
		&vendor.Status,
		&vendor.LastModifiedInCHPL,
		&vendor.CHPLID,
		&vendor.CreatedAt,
		&vendor.UpdatedAt)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &vendor.Location)
	if err != nil {
		return nil, err
	}

	return &vendor, err
}

// GetVendorUsingCHPLID gets a Vendor from the database using the given chpl id as a key.
// If the Vendor does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetVendorUsingCHPLID(ctx context.Context, id int) (*endpointmanager.Vendor, error) {
	var vendor endpointmanager.Vendor
	var locationJSON []byte

	sqlStatement := `
	SELECT
		id,
		name,
		developer_code,
		url,
		location,
		status,
		last_modified_in_chpl,
		chpl_id,
		created_at,
		updated_at
	FROM vendors WHERE chpl_id=$1`

	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&vendor.ID,
		&vendor.Name,
		&vendor.DeveloperCode,
		&vendor.URL,
		&locationJSON,
		&vendor.Status,
		&vendor.LastModifiedInCHPL,
		&vendor.CHPLID,
		&vendor.CreatedAt,
		&vendor.UpdatedAt)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &vendor.Location)
	if err != nil {
		return nil, err
	}

	return &vendor, err
}

// GetVendorUsingName gets a Vemdpr from the database using the given name as a key.
// If the Vendor does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetVendorUsingName(ctx context.Context, name string) (*endpointmanager.Vendor, error) {
	var vendor endpointmanager.Vendor
	var locationJSON []byte

	sqlStatement := `
	SELECT
		id,
		name,
		developer_code,
		url,
		location,
		status,
		last_modified_in_chpl,
		chpl_id,
		created_at,
		updated_at
	FROM vendors WHERE name=$1`

	row := s.DB.QueryRowContext(ctx, sqlStatement, name)

	err := row.Scan(
		&vendor.ID,
		&vendor.Name,
		&vendor.DeveloperCode,
		&vendor.URL,
		&locationJSON,
		&vendor.Status,
		&vendor.LastModifiedInCHPL,
		&vendor.CHPLID,
		&vendor.CreatedAt,
		&vendor.UpdatedAt)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &vendor.Location)
	if err != nil {
		return nil, err
	}

	return &vendor, err
}

// GetVendorNames returns a list of all of the vendor names
func (s *Store) GetVendorNames(ctx context.Context) ([]string, error) {
	var developers []string
	var developer string
	sqlStatement := "SELECT name FROM vendors"
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

// AddVendor adds the Vendor to the database.
func (s *Store) AddVendor(ctx context.Context, v *endpointmanager.Vendor) error {
	var err error

	locationJSON, err := json.Marshal(v.Location)
	if err != nil {
		return err
	}

	row := addVendorStatement.QueryRowContext(ctx,
		v.Name,
		v.DeveloperCode,
		v.URL,
		locationJSON,
		v.Status,
		v.LastModifiedInCHPL,
		v.CHPLID)

	err = row.Scan(&v.ID)

	return err
}

// UpdateVendor updates the Vendor in the database using the Vendor's database id as the key.
func (s *Store) UpdateVendor(ctx context.Context, v *endpointmanager.Vendor) error {
	var err error

	locationJSON, err := json.Marshal(v.Location)
	if err != nil {
		return err
	}

	_, err = updateVendorStatement.ExecContext(ctx,
		v.Name,
		v.DeveloperCode,
		v.URL,
		locationJSON,
		v.Status,
		v.LastModifiedInCHPL,
		v.CHPLID,
		v.ID)

	return err
}

// DeleteVendor deletes the Vendor from the database using the Vendor's database id  as the key.
func (s *Store) DeleteVendor(ctx context.Context, v *endpointmanager.Vendor) error {
	_, err := deleteVendorStatement.ExecContext(ctx, v.ID)

	return err
}

func prepareVendorStatements(s *Store) error {
	var err error
	addVendorStatement, err = s.DB.Prepare(`
		INSERT INTO vendors (
			name,
			developer_code,
			url,
			location,
			status,
			last_modified_in_chpl,
			chpl_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`)
	if err != nil {
		return err
	}
	updateVendorStatement, err = s.DB.Prepare(`
		UPDATE vendors
		SET 
			name = $1,
			developer_code = $2,
			url = $3,
			location = $4,
			status = $5,
			last_modified_in_chpl = $6,
			chpl_id = $7
		WHERE id = $8`)
	if err != nil {
		return err
	}
	deleteVendorStatement, err = s.DB.Prepare(`
        DELETE FROM vendors
        WHERE id = $1`)
	if err != nil {
		return err
	}
	return nil
}
