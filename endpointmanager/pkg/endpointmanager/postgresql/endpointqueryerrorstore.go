package postgresql

import (
	"context"
	"database/sql"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

var addEndpointQueryErrorStatement *sql.Stmt

func prepareEndpointQueryErrorStatements(s *Store) error {
	var err error
	addEndpointQueryErrorStatement, err = s.DB.Prepare(`
		INSERT INTO endpoint_query_errors (list_source, error_message, queried_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`)
	return err
}

func (s *Store) AddEndpointQueryError(ctx context.Context, qe *endpointmanager.EndpointQueryError) error {
	row := addEndpointQueryErrorStatement.QueryRowContext(ctx,
		qe.ListSource,
		qe.ErrorMessage,
		qe.QueriedAt)
	return row.Scan(&qe.ID, &qe.CreatedAt)
}
