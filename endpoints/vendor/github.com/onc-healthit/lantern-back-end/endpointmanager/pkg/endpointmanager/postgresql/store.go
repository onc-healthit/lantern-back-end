package postgresql

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// Store is the structure for working with the postgres database. It implements the following interfaces:
// FHIREndpointStore
// HealthITProductStore
// NPIOrganizationStore
//
// Usage:
//
// store := postgresql.NewStore(host, port, user, password, dbname, sslmode)
// defer store.Close()
// po := store.GetProviderOrganization(poID)
// <etc.>
type Store struct {
	DB *sql.DB
}

// Ensure Store implements endpointmanager.FHIREndpointStore.
var _ endpointmanager.FHIREndpointStore = &Store{}

// Ensure Store implements endpointmanager.HealthITProductStore.
var _ endpointmanager.HealthITProductStore = &Store{}

// Ensure Store implements endpointmanager.NPIOrganizationStore.
var _ endpointmanager.NPIOrganizationStore = &Store{}

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

	return &store, nil
}

// Close closes the postgresql database connection.
func (s *Store) Close() {
	s.DB.Close()
}
