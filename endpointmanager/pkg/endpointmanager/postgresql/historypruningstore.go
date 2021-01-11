package postgresql

import (
	"context"
	"database/sql"
	"strconv"
)

var pruningStatementQueryInterval *sql.Stmt
var pruningStatementNoQueryInterval *sql.Stmt
var pruningDeleteStatement *sql.Stmt

// PruningGetInfoHistory gets info history entries for pruning
func (s *Store) PruningGetInfoHistory(ctx context.Context, threshold int, queryInterval int) (*sql.Rows, error) {

	var rows *sql.Rows
	var err error

	thresholdString := strconv.Itoa(threshold)

	if queryInterval >= 0 {
		queryIntString := strconv.Itoa(threshold + (3 * queryInterval))
		rows, err = pruningStatementQueryInterval.QueryContext(ctx, thresholdString, queryIntString)
	} else {
		rows, err = pruningStatementNoQueryInterval.QueryContext(ctx, thresholdString)
	}

	return rows, err
}

// PruningDeleteInfoHistory deletes info history entry due to pruning
func (s *Store) PruningDeleteInfoHistory(ctx context.Context, url string, entryDate string) error {
	_, err := pruningDeleteStatement.ExecContext(ctx, url, entryDate)
	return err
}

func prepareHistoryPruningStatements(s *Store) error {
	var err error

	pruningStatementQueryInterval, err = s.DB.Prepare(`
		SELECT operation, url, capability_statement, entered_at, tls_version, mime_types, smart_response FROM fhir_endpoints_info_history 
		WHERE (operation='U' OR operation='I') 
			AND (date_trunc('minute', entered_at) <= date_trunc('minute', current_date - interval '$1 minute'))
			AND (date_trunc('minute', entered_at) >= date_trunc('minute', current_date - interval '$2 minute'))
		ORDER BY url, entered_at ASC;`)
	if err != nil {
		return err
	}
	pruningStatementNoQueryInterval, err = s.DB.Prepare(`
		SELECT operation, url, capability_statement, entered_at, tls_version, mime_types, smart_response FROM fhir_endpoints_info_history 
		WHERE (operation='U' OR operation='I') 
			AND (date_trunc('minute', entered_at) <= date_trunc('minute', current_date - interval '$1 minute')) 
		ORDER BY url, entered_at ASC;`)
	if err != nil {
		return err
	}
	pruningDeleteStatement, err = s.DB.Prepare(`
		DELETE FROM fhir_endpoints_info_history WHERE url=$1 AND operation='U' AND entered_at = $2;`)
	if err != nil {
		return err
	}
	return nil
}
