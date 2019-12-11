package testhelper

import (
	"database/sql"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
)

// HostAndPort holds the host and port information for a resource.
type HostAndPort struct {
	Host string
	Port string
}

// Assert checks that the boolean statement is true. If not, it fails the test with the given
// error value.
// Assert streamlines test checks.
func Assert(t *testing.T, boolStatement bool, errorValue interface{}) {
	if !boolStatement {
		t.Fatalf("%s: %v", t.Name(), errorValue)
	}
}

// CheckResources ensures that any resources needed for an integration test are available.
// If a resource is not available, it returns an error.
func CheckResources(haps ...HostAndPort) error {
	for _, hap := range haps {
		host := hap.Host
		port := hap.Port

		timeout := time.Second
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
		if err != nil {
			err := errors.Wrapf(err, "unable to connect to resource %s:%s", host, port)
			return err
		}
		if conn != nil {
			conn.Close()
		}
	}
	return nil
}

// IntegrationDBTestSetup ensures that the database is empty before running any tests and returns
// a teardown function that will delete all entries in the database.
func IntegrationDBTestSetup(t *testing.T, db *sql.DB) (func(t *testing.T, db *sql.DB), error) {

	tableNames, err := getTableNames(db)
	Assert(t, err == nil, err)

	areEmpty, err := tablesAreEmpty(tableNames, db)
	Assert(t, err == nil, err)
	Assert(t, areEmpty, "at least one database table has entries in it. database tables must be empty before running integration tests.")

	return integrationDBTestTeardown, nil
}

func integrationDBTestTeardown(t *testing.T, db *sql.DB) {
	tableNames, err := getTableNames(db)
	Assert(t, err == nil, err)

	err = deleteTableEntries(tableNames, db)
	Assert(t, err == nil, err)
}

func getTableNames(db *sql.DB) ([]string, error) {
	var err error
	var query string
	var tableNames []string

	schemanamesToIgnore := []string{
		"pg_catalog",
		"information_schema",
		"_timescaledb_catalog",
		"_timescaledb_config",
		"_timescaledb_internal",
		"_timescaledb_cache",
	}

	query = "SELECT tablename FROM pg_catalog.pg_tables"

	for i, schemaname := range schemanamesToIgnore {
		if i == 0 {
			query = query + " WHERE"
		} else {
			query = query + " AND"
		}

		query = fmt.Sprintf("%s schemaname != '%s'", query, schemaname)
	}

	rows, err := db.Query(query)
	if err != nil {
		print(err.Error())
		return nil, err
	}
	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			print(err.Error())
			return nil, err
		}
		// also ignore table names that start with 'metrics'
		if !strings.HasPrefix(tableName, "metrics") {
			tableNames = append(tableNames, tableName)
		}
	}

	return tableNames, err
}

func tablesAreEmpty(tableNames []string, db *sql.DB) (bool, error) {
	for _, tableName := range tableNames {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		row := db.QueryRow(query)
		err := row.Scan(&count)
		if err != nil {
			return false, err
		}

		if count != 0 {
			return false, nil
		}
	}
	return true, nil
}

func deleteTableEntries(tableNames []string, db *sql.DB) error {
	for _, tableName := range tableNames {
		query := fmt.Sprintf("DELETE FROM %s", tableName)
		result, err := db.Exec(query)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		println(affected)
	}
	return nil
}
