package postgresql

import (
	"context"
	"database/sql"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// prepared statements are left open to be used throughout the execution of the application
var addValidationStatement *sql.Stmt
var addValidationResultStatement *sql.Stmt

// GetValidationByID gets the rows of the validation table that have the given validation_result_id
func (s *Store) GetValidationByID(ctx context.Context, id int) (*[]endpointmanager.Rule, error) {
	var validationRows []endpointmanager.Rule

	sqlStatementInfo := `
	SELECT
		rule_name,
		valid,
		expected,
		actual,
		comment,
		reference,
		implementation_guide
	FROM validations WHERE validation_result_id=$1`

	rows, err := s.DB.QueryContext(ctx, sqlStatementInfo, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var ruleInfo endpointmanager.Rule

		err := rows.Scan(
			&ruleInfo.RuleName,
			&ruleInfo.Valid,
			&ruleInfo.Expected,
			&ruleInfo.Actual,
			&ruleInfo.Comment,
			&ruleInfo.Reference,
			&ruleInfo.ImplGuide)
		if err != nil {
			return nil, err
		}
		validationRows = append(validationRows, ruleInfo)
	}
	return &validationRows, nil
}

// AddValidationResult creates a new ID for the validation data and returns it
func (s *Store) AddValidationResult(ctx context.Context) (int, error) {
	var err error

	valResRow := addValidationResultStatement.QueryRowContext(ctx)
	valResID := 0
	err = valResRow.Scan(&valResID)

	return valResID, err
}

// AddValidation adds the Validation data to the database
func (s *Store) AddValidation(ctx context.Context, v *endpointmanager.Validation, valResID int) error {
	var err error

	for _, ruleInfo := range v.Results {
		_, err = addValidationStatement.ExecContext(ctx,
			ruleInfo.RuleName,
			ruleInfo.Valid,
			ruleInfo.Expected,
			ruleInfo.Actual,
			ruleInfo.Comment,
			ruleInfo.Reference,
			ruleInfo.ImplGuide,
			valResID)
		if err != nil {
			return err
		}
	}

	return err
}

func prepareValidationStatements(s *Store) error {
	var err error
	addValidationResultStatement, err = s.DB.Prepare(`
		INSERT INTO validation_results (id)
		VALUES (DEFAULT)
		RETURNING id;`)
	if err != nil {
		return err
	}
	addValidationStatement, err = s.DB.Prepare(`
	INSERT INTO validations (
		rule_name,
		valid,
		expected,
		actual,
		comment,
		reference,
		implementation_guide,
		validation_result_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)
	if err != nil {
		return err
	}
	return nil
}
