package postgresql

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // specified to do this for accessing postgres db
)

// Store is the structure for working with the postgres database.
// Usage:
//
// store := postgresql.NewStore(host, port, user, password, dbname, sslmode)
// defer store.Close()
// po := store.GetProviderOrganization(poID)
// <etc.>
type Store struct {
	DB *sql.DB
}

// NewStore creates a connection to the postgresql database and adds a reference to the database
// in store.DB.
func NewStore(host string, port int, user string, password string, dbname string, sslmode string) (*Store, error) {
	var store Store
	var err error

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	store.DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		err = fmt.Errorf("Error opening database: %s", err.Error())
		panic(err.Error())
	}

	// calling db.Ping to create a connection to the database.
	// db.Open only validates the arguments, it does not create the connection.
	err = store.DB.Ping()
	if err != nil {
		err = fmt.Errorf("Error creating connection to database: %s", err.Error())
		panic(err.Error())
	}

	err = prepareFHIREndpointStatements(&store)
	if err != nil {
		return nil, err
	}
	err = prepareFHIREndpointInfoStatements(&store)
	if err != nil {
		return nil, err
	}
	err = prepareHealthITProductStatements(&store)
	if err != nil {
		return nil, err
	}
	err = prepareCriteriaStatements(&store)
	if err != nil {
		return nil, err
	}
	err = prepareNPIOrganizationStatements(&store)
	if err != nil {
		return nil, err
	}
	err = prepareVendorStatements(&store)
	if err != nil {
		return nil, err
	}
	err = prepareNPIContactStatements(&store)
	if err != nil {
		return nil, err
	}

	return &store, nil
}

// Close closes the postgresql database connection.
func (s *Store) Close() {
	s.DB.Close()
}

// converts foreign key ints to nullable ints so we don't have issues with non-existent foreign key references.
func getNullableInts(regularInts []int) []sql.NullInt64 {
	nullableInts := make([]sql.NullInt64, len(regularInts))

	for i, regInt := range regularInts {
		var nullInt sql.NullInt64
		if regInt < 1 {
			nullInt.Valid = false
		} else {
			nullInt.Valid = true
			nullInt.Int64 = int64(regInt)
		}
		nullableInts[i] = nullInt
	}
	return nullableInts
}

// converts nullable into to an integer. null values are made to be 0s. This should only be used for foreign key references. postgres does not use 0 as an index - starts at 1.
func getRegularInts(nullableInts []sql.NullInt64) []int {
	regularInts := make([]int, len(nullableInts))

	for i, nullInt := range nullableInts {
		var regInt int

		if !nullInt.Valid {
			regInt = 0
		} else {
			regInt = int(nullInt.Int64)
		}
		regularInts[i] = regInt
	}
	return regularInts
}
