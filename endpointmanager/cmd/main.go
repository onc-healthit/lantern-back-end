package main

import (
	"fmt"

	_ "github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
)

// TODO: configuration file or commandline arguments
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "" // NOTE: this needs to be replaced with the appropriate password
	dbname   = "postgres"
	sslmode  = "disable"
)

func main() {
	//var endpoint models.FHIREndpoint
	var err error

	store, err := postgresql.NewStore(host, port, user, password, dbname, sslmode)
	if err != nil {
		panic(err.Error())
	}
	defer store.Close()

	fmt.Println("Successfully connected!")
}
