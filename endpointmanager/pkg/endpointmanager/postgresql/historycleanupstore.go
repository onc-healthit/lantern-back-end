package postgresql

import (
	"context"
	"database/sql"

	log "github.com/sirupsen/logrus"
)

var duplicateInfoHistoryStatement *sql.Stmt
var distinctURLStatement *sql.Stmt

// GetDistinctURLs gets a list of ordered distinct URLs from the history table
func (s *Store) GetDistinctURLsFromHistory(ctx context.Context) (*sql.Rows, error) {

	log.Info("Inside GetDistinctURLsFromHistory")

	var err error

	distinctURLStatement, err = s.DB.Prepare(`
		select DISTINCT(url) FROM fhir_endpoints_info_history
		WHERE (operation='U' OR operation='I')
		ORDER BY url;`)

	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

	rows, err = distinctURLStatement.QueryContext(ctx)

	return rows, err
}

// PruningGetInfoHistoryUsingURL gets info history entries matching the given URL for pruning
func (s *Store) PruningGetInfoHistoryUsingURL(ctx context.Context, queryInterval bool, url string) (*sql.Rows, error) {

	log.Info("Inside PruningGetInfoHistoryUsingURL")

	var err error

	duplicateInfoHistoryStatement, err = s.DB.Prepare(`
		SELECT operation, url, capability_statement, entered_at, tls_version, mime_types, smart_response, validation_result_id, requested_fhir_version FROM fhir_endpoints_info_history 
		WHERE (operation='U' OR operation='I') AND url = $1
		ORDER BY entered_at ASC;`)

	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

	rows, err = duplicateInfoHistoryStatement.QueryContext(ctx, url)

	return rows, err
}
