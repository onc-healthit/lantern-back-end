package postgresql

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/spf13/viper"
)

var pruningStatementQueryInterval *sql.Stmt
var pruningStatementNoQueryInterval *sql.Stmt
var pruningDeleteStatement *sql.Stmt

// PruningGetInfoHistory gets info history entries for pruning
func (s *Store) PruningGetInfoHistory(ctx context.Context, queryInterval bool) (*sql.Rows, error) {

	var rows *sql.Rows
	var err error

	if queryInterval {
		rows, err = pruningStatementQueryInterval.QueryContext(ctx)
	} else {
		rows, err = pruningStatementNoQueryInterval.QueryContext(ctx)
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

	pruningThreshold := viper.GetInt("pruning_threshold")
	queryInterval := viper.GetInt("capquery_qryintvl")

	thresholdString := strconv.Itoa(pruningThreshold)
	queryIntString := strconv.Itoa(pruningThreshold + (3 * queryInterval))

	pruningStatementQueryInterval, err = s.DB.Prepare(`
		SELECT operation, url, capability_statement, entered_at, tls_version, mime_types, smart_response FROM fhir_endpoints_info_history 
		WHERE (operation='U' OR operation='I') 
			AND (date_trunc('minute', entered_at) <= date_trunc('minute', current_date - INTERVAL '` + thresholdString + ` minute'))
			AND (date_trunc('minute', entered_at) >= date_trunc('minute', current_date - INTERVAL '` + queryIntString + ` minute'))
		ORDER BY url, entered_at ASC;`)
	if err != nil {
		return err
	}
	pruningStatementNoQueryInterval, err = s.DB.Prepare(`
		SELECT operation, url, capability_statement, entered_at, tls_version, mime_types, smart_response FROM fhir_endpoints_info_history 
		WHERE (operation='U' OR operation='I') 
			AND (date_trunc('minute', entered_at) <= date_trunc('minute', current_date - INTERVAL '` + thresholdString + ` minute')) 
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
