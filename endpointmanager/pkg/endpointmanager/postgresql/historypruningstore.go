package postgresql

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/spf13/viper"
)

var pruningStatementQueryInterval *sql.Stmt
var pruningStatementNoQueryInterval *sql.Stmt
var pruningStatementCustomQueryInterval *sql.Stmt
var pruningDeleteStatement *sql.Stmt
var pruningDeleteValStatement *sql.Stmt
var pruningDeleteValResStatement *sql.Stmt
var distinctURLStatementQueryInterval *sql.Stmt
var distinctURLStatementNoQueryInterval *sql.Stmt
var distinctURLStatementCustomQueryInterval *sql.Stmt
var addPruningMetadataStatementQueryInterval *sql.Stmt
var addPruningMetadataStatementNoQueryInterval *sql.Stmt
var addPruningMetadataStatementCustomQueryInterval *sql.Stmt
var updatePruningMetadataStatement *sql.Stmt
var pruningMetadataCountStatement *sql.Stmt
var lastPruneStatement *sql.Stmt

// GetDistinctURLsFromHistory gets a list of ordered distinct URLs from the history table
func (s *Store) GetDistinctURLs(ctx context.Context, queryInterval bool, lastPruneSuccessful bool, lastPruneQueryIntStartDate string, lastPruneQueryIntEndDate string) (*sql.Rows, error) {

	var err error
	var rows *sql.Rows

	if queryInterval {
		if lastPruneQueryIntStartDate != "" && lastPruneQueryIntEndDate != "" {
			if lastPruneSuccessful {
				rows, err = distinctURLStatementCustomQueryInterval.QueryContext(ctx, lastPruneQueryIntEndDate)
			} else {
				rows, err = distinctURLStatementCustomQueryInterval.QueryContext(ctx, lastPruneQueryIntStartDate)
			}
		} else {
			rows, err = distinctURLStatementQueryInterval.QueryContext(ctx)
		}
	} else {
		rows, err = distinctURLStatementNoQueryInterval.QueryContext(ctx)
	}

	return rows, err
}

func (s *Store) GetPruningMetadataCount(ctx context.Context) (*sql.Rows, error) {

	var err error
	var rows *sql.Rows

	rows, err = pruningMetadataCountStatement.QueryContext(ctx)

	return rows, err
}

func (s *Store) GetLastPruneEntryDate(ctx context.Context) (*sql.Rows, error) {

	var err error
	var rows *sql.Rows

	rows, err = lastPruneStatement.QueryContext(ctx)

	return rows, err
}

func (s *Store) AddPruningMetadata(ctx context.Context, queryInterval bool, lastPruneSuccessful bool, lastPruneQueryIntStartDate string, lastPruneQueryIntEndDate string) (int, error) {

	var err error
	var row *sql.Row
	var id int

	if queryInterval {
		if lastPruneQueryIntStartDate != "" && lastPruneQueryIntEndDate != "" {
			if lastPruneSuccessful {
				row = addPruningMetadataStatementCustomQueryInterval.QueryRowContext(ctx, lastPruneQueryIntEndDate)
			} else {
				row = addPruningMetadataStatementCustomQueryInterval.QueryRowContext(ctx, lastPruneQueryIntStartDate)
			}
		} else {
			row = addPruningMetadataStatementQueryInterval.QueryRowContext(ctx)
		}
	} else {
		row = addPruningMetadataStatementNoQueryInterval.QueryRowContext(ctx)
	}

	err = row.Scan(&id)

	return id, err
}

// GetInfoHistoryCountBeforeThreshold gets the count of rows in the info history table during the pruning query interval and threshold time frame
func (s *Store) UpdatePruningMetadata(ctx context.Context, pruningMetadataId int, successful bool, numRowsProcessed int, numRowsPruned int) error {

	var err error

	_, err = updatePruningMetadataStatement.ExecContext(ctx, pruningMetadataId, successful, numRowsProcessed, numRowsPruned)

	return err
}

// PruningGetInfoHistory gets info history entries for pruning
func (s *Store) PruningGetInfoHistory(ctx context.Context, queryInterval bool, url string, lastPruneSuccessful bool, lastPruneQueryIntStartDate string, lastPruneQueryIntEndDate string) (*sql.Rows, error) {

	var rows *sql.Rows
	var err error

	if queryInterval {
		if lastPruneQueryIntStartDate != "" && lastPruneQueryIntEndDate != "" {
			if lastPruneSuccessful {
				rows, err = pruningStatementCustomQueryInterval.QueryContext(ctx, url, lastPruneQueryIntEndDate)
			} else {
				rows, err = pruningStatementCustomQueryInterval.QueryContext(ctx, url, lastPruneQueryIntStartDate)
			}
		} else {
			rows, err = pruningStatementQueryInterval.QueryContext(ctx, url)
		}
	} else {
		rows, err = pruningStatementNoQueryInterval.QueryContext(ctx, url)
	}

	return rows, err
}

// PruningDeleteInfoHistory deletes info history entry due to pruning
func (s *Store) PruningDeleteInfoHistory(ctx context.Context, url string, entryDate string, requested_fhir_version string) error {
	_, err := pruningDeleteStatement.ExecContext(ctx, url, requested_fhir_version, entryDate)
	return err
}

// CheckIfValidationResultIDExists returns true if the fhir_endpoints_info table contains validation_result_id, else false
func (s *Store) CheckIfValidationResultIDExists(ctx context.Context, valResID int) (bool, error) {
	var count int

	// Ensure the current entry in fhir_endpoints_info table does not this validation result id
	row := s.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoints_info WHERE validation_result_id=$1;", valResID)

	err := row.Scan(&count)
	if err != nil {
		return true, err
	}

	// If there is an entry in the fhir endpoints info table that has this id, do nothing
	if count > 0 {
		return true, nil
	}

	return false, nil
}

// PruningDeleteValidationTable deletes validation table entries based on the given ID
func (s *Store) PruningDeleteValidationTable(ctx context.Context, valResID int) error {
	_, err := pruningDeleteValStatement.ExecContext(ctx, valResID)
	return err
}

// PruningDeleteValidationResultEntry deletes an entry from the validation_results table based
// on the given ID
func (s *Store) PruningDeleteValidationResultEntry(ctx context.Context, valResID int) error {
	_, err := pruningDeleteValResStatement.ExecContext(ctx, valResID)
	return err
}

func prepareHistoryPruningStatements(s *Store) error {
	var err error

	pruningThreshold := viper.GetInt("pruning_threshold")

	thresholdString := strconv.Itoa(pruningThreshold)
	queryIntString := strconv.Itoa(pruningThreshold + 7200)

	distinctURLStatementQueryInterval, err = s.DB.Prepare(`
		select DISTINCT(url) FROM fhir_endpoints_info_history
		WHERE (operation='U' OR operation='I')
			AND (date_trunc('minute', entered_at) <= date_trunc('minute', current_date - INTERVAL '` + thresholdString + ` minute'))
			AND (date_trunc('minute', entered_at) >= date_trunc('minute', current_date - INTERVAL '` + queryIntString + ` minute'))
		ORDER BY url;`)
	if err != nil {
		return err
	}
	distinctURLStatementCustomQueryInterval, err = s.DB.Prepare(`
		select DISTINCT(url) FROM fhir_endpoints_info_history
		WHERE (operation='U' OR operation='I')
			AND (date_trunc('minute', entered_at) <= date_trunc('minute', current_date - INTERVAL '` + thresholdString + ` minute'))
			AND (date_trunc('minute', entered_at) >= LEAST(date_trunc('minute', date($1)),
														date_trunc('minute', current_date - INTERVAL '` + queryIntString + ` minute')))
		ORDER BY url;`)
	if err != nil {
		return err
	}
	distinctURLStatementNoQueryInterval, err = s.DB.Prepare(`
		select DISTINCT(url) FROM fhir_endpoints_info_history
		WHERE (operation='U' OR operation='I')
			AND (date_trunc('minute', entered_at) <= date_trunc('minute', current_date - INTERVAL '` + thresholdString + ` minute'))
		ORDER BY url;`)
	if err != nil {
		return err
	}
	pruningMetadataCountStatement, err = s.DB.Prepare(`
		SELECT COUNT(*) FROM info_history_pruning_metadata;`)
	if err != nil {
		return err
	}
	lastPruneStatement, err = s.DB.Prepare(`
		SELECT successful, query_int_start_date, query_int_end_date FROM info_history_pruning_metadata
		ORDER BY started_on DESC LIMIT 1;`)
	if err != nil {
		return err
	}
	addPruningMetadataStatementQueryInterval, err = s.DB.Prepare(`
		INSERT INTO info_history_pruning_metadata (query_int_start_date, query_int_end_date)
		VALUES (date_trunc('minute', current_date - INTERVAL '` + queryIntString + ` minute'),
				date_trunc('minute', current_date - INTERVAL '` + thresholdString + ` minute'))
		RETURNING id`)
	if err != nil {
		return err
	}
	addPruningMetadataStatementCustomQueryInterval, err = s.DB.Prepare(`
		INSERT INTO info_history_pruning_metadata (query_int_start_date, query_int_end_date)
		VALUES (LEAST(date_trunc('minute', date($1)),
					date_trunc('minute', current_date - INTERVAL '` + queryIntString + ` minute')),
				date_trunc('minute', current_date - INTERVAL '` + thresholdString + ` minute'))
		RETURNING id`)
	if err != nil {
		return err
	}
	addPruningMetadataStatementNoQueryInterval, err = s.DB.Prepare(`
		INSERT INTO info_history_pruning_metadata (query_int_start_date, query_int_end_date)
		SELECT date_trunc('minute', entered_at),
				date_trunc('minute', current_date - INTERVAL '` + thresholdString + ` minute')
		FROM fhir_endpoints_info_history
		ORDER BY entered_at ASC
		LIMIT 1
		RETURNING id`)
	if err != nil {
		return err
	}
	updatePruningMetadataStatement, err = s.DB.Prepare(`
		UPDATE info_history_pruning_metadata
		SET successful = $2,
			num_rows_processed = $3,
			num_rows_pruned = $4,
			ended_on = now()
		WHERE id = $1`)
	if err != nil {
		return err
	}
	pruningStatementQueryInterval, err = s.DB.Prepare(`
		SELECT operation, url, capability_statement, entered_at, tls_version, mime_types, smart_response, validation_result_id, requested_fhir_version FROM fhir_endpoints_info_history 
		WHERE (operation='U' OR operation='I')
			AND url = $1
			AND (date_trunc('minute', entered_at) <= date_trunc('minute', current_date - INTERVAL '` + thresholdString + ` minute'))
			AND (date_trunc('minute', entered_at) >= COALESCE((SELECT date_trunc('minute', entered_at) 
														FROM fhir_endpoints_info_history 
														WHERE (operation='U' OR operation='I') 
														AND (date_trunc('minute', entered_at) < date_trunc('minute', current_date - INTERVAL '` + queryIntString + ` minute'))
														AND url = $1
													ORDER BY entered_at DESC LIMIT 1),
													date_trunc('minute', current_date - INTERVAL '` + queryIntString + ` minute')))
		ORDER BY entered_at ASC;`)
	if err != nil {
		return err
	}
	pruningStatementCustomQueryInterval, err = s.DB.Prepare(`
		SELECT operation, url, capability_statement, entered_at, tls_version, mime_types, smart_response, validation_result_id, requested_fhir_version FROM fhir_endpoints_info_history 
		WHERE (operation='U' OR operation='I')
			AND url = $1
			AND (date_trunc('minute', entered_at) <= date_trunc('minute', current_date - INTERVAL '` + thresholdString + ` minute'))
			AND (date_trunc('minute', entered_at) >= COALESCE((SELECT date_trunc('minute', entered_at) 
																FROM fhir_endpoints_info_history 
																WHERE (operation='U' OR operation='I') 
																AND (date_trunc('minute', entered_at) < LEAST(date_trunc('minute', date($2)),
																											date_trunc('minute', current_date - INTERVAL '` + queryIntString + ` minute')))
																AND url = $1
															ORDER BY entered_at DESC LIMIT 1), 
															LEAST(date_trunc('minute', date($2)),
																date_trunc('minute', current_date - INTERVAL '` + queryIntString + ` minute'))))
		ORDER BY entered_at ASC;`)
	if err != nil {
		return err
	}
	pruningStatementNoQueryInterval, err = s.DB.Prepare(`
		SELECT operation, url, capability_statement, entered_at, tls_version, mime_types, smart_response, validation_result_id, requested_fhir_version FROM fhir_endpoints_info_history 
		WHERE (operation='U' OR operation='I') 
			AND (date_trunc('minute', entered_at) <= date_trunc('minute', current_date - INTERVAL '` + thresholdString + ` minute'))
			AND url = $1
		ORDER BY entered_at ASC;`)
	if err != nil {
		return err
	}
	pruningDeleteStatement, err = s.DB.Prepare(`
		DELETE FROM fhir_endpoints_info_history WHERE url=$1 AND operation='U' AND requested_fhir_version=$2 AND entered_at = $3;`)
	if err != nil {
		return err
	}
	pruningDeleteValStatement, err = s.DB.Prepare(`
		DELETE FROM validations WHERE validation_result_id = $1;`)
	if err != nil {
		return err
	}
	pruningDeleteValResStatement, err = s.DB.Prepare(`
		DELETE FROM validation_results WHERE id = $1;`)
	if err != nil {
		return err
	}
	return nil
}
