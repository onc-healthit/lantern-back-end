package postgresql

import (
	"context"
	"database/sql"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// prepared statements are left open to be used throughout the execution of the application
// TODO: figure out if there's a better way to manage this for bulk calls
var addCriteriaStatement *sql.Stmt
var updateCriteriaStatement *sql.Stmt
var deleteCriteriaStatement *sql.Stmt

// GetCriteria gets a CertificationCriteria from the database using the database ID as a key.
// If the CertificationCriteria does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetCriteria(ctx context.Context, id int) (*endpointmanager.CertificationCriteria, error) {
	var criteria endpointmanager.CertificationCriteria

	sqlStatement := `
	SELECT
		id,
		certification_id,
		cerification_number,
		title,
		certification_edition_id,
		certification_edition,
		description,
		removed,
		created_at,
		updated_at
	FROM certification_criteria WHERE id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&criteria.ID,
		&criteria.CertificationID,
		&criteria.CertificationNumber,
		&criteria.Title,
		&criteria.CertificationEditionID,
		&criteria.CertificationEdition,
		&criteria.Description,
		&criteria.Removed,
		&criteria.CreatedAt,
		&criteria.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &criteria, err
}

// GetCriteriaByCertificationID gets a CertificationCriteria from the database using the certification's
// id value as key. If the CertificationCriteria does not exist in the database, sql.ErrNoRows
// will be returned.
func (s *Store) GetCriteriaByCertificationID(ctx context.Context, certID int) (*endpointmanager.CertificationCriteria, error) {
	var criteria endpointmanager.CertificationCriteria

	sqlStatement := `
	SELECT
		id,
		certification_id,
		cerification_number,
		title,
		certification_edition_id,
		certification_edition,
		description,
		removed,
		created_at,
		updated_at
	FROM certification_criteria WHERE certification_id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, certID)

	err := row.Scan(
		&criteria.ID,
		&criteria.CertificationID,
		&criteria.CertificationNumber,
		&criteria.Title,
		&criteria.CertificationEditionID,
		&criteria.CertificationEdition,
		&criteria.Description,
		&criteria.Removed,
		&criteria.CreatedAt,
		&criteria.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &criteria, err
}

// AddCriteria adds the CertificationCriteria to the database.
func (s *Store) AddCriteria(ctx context.Context, criteria *endpointmanager.CertificationCriteria) error {
	row := addCriteriaStatement.QueryRowContext(ctx,
		criteria.CertificationID,
		criteria.CertificationNumber,
		criteria.Title,
		criteria.CertificationEditionID,
		criteria.CertificationEdition,
		criteria.Description,
		criteria.Removed)

	err := row.Scan(&criteria.ID)

	return err
}

// UpdateCriteria updates the CertificationCriteria in the database using the CertificationCriteria's database ID as the key.
func (s *Store) UpdateCriteria(ctx context.Context, criteria *endpointmanager.CertificationCriteria) error {

	_, err := updateCriteriaStatement.ExecContext(ctx,
		criteria.CertificationID,
		criteria.CertificationNumber,
		criteria.Title,
		criteria.CertificationEditionID,
		criteria.CertificationEdition,
		criteria.Description,
		criteria.Removed,
		criteria.ID)

	return err
}

// DeleteCriteria deletes the CertificationCriteria from the database using the CertificationCriteria's database ID as the key.
func (s *Store) DeleteCriteria(ctx context.Context, criteria *endpointmanager.CertificationCriteria) error {
	_, err := deleteCriteriaStatement.ExecContext(ctx, criteria.ID)

	return err
}

func prepareCriteriaStatements(s *Store) error {
	var err error
	addCriteriaStatement, err = s.DB.Prepare(`
		INSERT INTO certification_criteria (
			certification_id,
			cerification_number,
			title,
			certification_edition_id,
			certification_edition,
			description,
			removed)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`)
	if err != nil {
		return err
	}
	updateCriteriaStatement, err = s.DB.Prepare(`
		UPDATE certification_criteria
		SET certification_id = $1,
			cerification_number = $2,
			title = $3,
			certification_edition_id = $4,
			certification_edition = $5,
			description = $6,
			removed = $7
		WHERE id=$8`)
	if err != nil {
		return err
	}
	deleteCriteriaStatement, err = s.DB.Prepare(`
		DELETE FROM certification_criteria
		WHERE id=$1`)
	if err != nil {
		return err
	}
	return nil
}
