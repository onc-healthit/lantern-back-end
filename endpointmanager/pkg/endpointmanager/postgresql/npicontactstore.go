package postgresql

import (
	"context"
	"encoding/json"

	"database/sql"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// prepared statements are left open to be used throughout the execution of the application
var addNPIContactStatement *sql.Stmt
var updateNPIContactByNPIIDStatement *sql.Stmt
var deleteNPIContactStatement *sql.Stmt

// GetNPIContactByNPIID gets a NPIContact from the database using the NPI id as a key.
// If the NPIContact does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetNPIContactByNPIID(ctx context.Context, npiID string) (*endpointmanager.NPIContact, error) {
	var contact endpointmanager.NPIContact
	var locationJSON []byte

	sqlStatement := `
	SELECT
	id,
	npi_id,
	endpoint_type,
	endpoint_type_description,
	endpoint,
	valid_url,
	affiliation,
	endpoint_description,
	affiliation_legal_business_name,
	use_code,
	use_description,
	other_use_description,
	content_type,
	content_description,
	other_content_description,
	location,
	created_at,
	updated_at
	FROM npi_contacts WHERE npi_id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, npiID)

	err := row.Scan(
		&contact.ID,
		&contact.NPI_ID,
		&contact.EndpointType,
		&contact.EndpointTypeDescription,
		&contact.Endpoint,
		&contact.ValidURL,
		&contact.Affiliation,
		&contact.EndpointDescription,
		&contact.AffiliationLegalBusinessName,
		&contact.UseCode,
		&contact.UseDescription,
		&contact.OtherUseDescription,
		&contact.ContentType,
		&contact.ContentDescription,
		&contact.OtherContentDescription,
		&locationJSON,
		&contact.CreatedAt,
		&contact.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &contact.Location)

	if err != nil {
		return nil, err
	}

	return &contact, err
}

// DeleteAllNPIContacts will remove all rows from the npi_Contacts table
func (s *Store) DeleteAllNPIContacts(ctx context.Context) error {
	sqlStatement := `DELETE FROM npi_contacts`
	_, err := s.DB.ExecContext(ctx, sqlStatement)
	return err
}

// AddNPIContact adds the NPIContact to the database or updates if there is an existsing entry with same NPI_ID
func (s *Store) AddNPIContact(ctx context.Context, contact *endpointmanager.NPIContact) error {
	locationJSON, err := json.Marshal(contact.Location)
	if err != nil {
		return err
	}
	row := addNPIContactStatement.QueryRowContext(ctx,
		contact.NPI_ID,
		contact.EndpointType,
		contact.EndpointTypeDescription,
		contact.Endpoint,
		contact.ValidURL,
		contact.Affiliation,
		contact.EndpointDescription,
		contact.AffiliationLegalBusinessName,
		contact.UseCode,
		contact.UseDescription,
		contact.OtherUseDescription,
		contact.ContentType,
		contact.ContentDescription,
		contact.OtherUseDescription,
		locationJSON,
	)
	err = row.Scan(&contact.ID)

	return err
}

// UpdateNPIContactByNPIID updates the NPIContact in the database using the NPIContact's NPIID as the key.
func (s *Store) UpdateNPIContactByNPIID(ctx context.Context, contact *endpointmanager.NPIContact) error {
	locationJSON, err := json.Marshal(contact.Location)
	if err != nil {
		return err
	}

	_, err = updateNPIContactByNPIIDStatement.ExecContext(ctx,
		contact.NPI_ID,
		contact.EndpointType,
		contact.EndpointTypeDescription,
		contact.Endpoint,
		contact.ValidURL,
		contact.Affiliation,
		contact.EndpointDescription,
		contact.AffiliationLegalBusinessName,
		contact.UseCode,
		contact.UseDescription,
		contact.OtherUseDescription,
		contact.ContentType,
		contact.ContentDescription,
		contact.OtherUseDescription,
		locationJSON)

	return err
}

// DeleteNPIContact deletes the NPIContact from the database using the NPIContact's database ID as the key.
func (s *Store) DeleteNPIContact(ctx context.Context, org *endpointmanager.NPIContact) error {
	_, err := deleteNPIContactStatement.ExecContext(ctx, org.ID)

	return err
}

func prepareNPIContactStatements(s *Store) error {
	var err error
	addNPIContactStatement, err = s.DB.Prepare(`
		INSERT INTO npi_contacts (
			npi_id,
			endpoint_type,
			endpoint_type_description,
			endpoint,
			valid_url,
			affiliation,
			endpoint_description,
			affiliation_legal_business_name,
			use_code,
			use_description,
			other_use_description,
			content_type,
			content_description,
			other_content_description,
			location)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id`)
	if err != nil {
		return err
	}
	updateNPIContactByNPIIDStatement, err = s.DB.Prepare(`
		UPDATE npi_contacts
		SET endpoint_type = $2,
		endpoint_type_description = $3,
		endpoint = $4,
		valid_url = $5,
		affiliation = $6,
		endpoint_description = $7,
		affiliation_legal_business_name = $8,
		use_code = $9,
		use_description = $10,
		other_use_description = $11,
		content_type = $12,
		content_description = $13,
		other_content_description = $14,
		location = $15
		WHERE npi_id=$1`)
	if err != nil {
		return err
	}
	deleteNPIContactStatement, err = s.DB.Prepare(`
		DELETE FROM npi_contacts
		WHERE id=$1`)
	if err != nil {
		return err
	}
	return nil
}
