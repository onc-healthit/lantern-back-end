package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
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

var db *sql.DB

func main() {
	//var endpoint models.FHIREndpoint
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// calling db.Ping to create a connection to the database.
	// db.Open only validates the arguments, it does not create the connection.
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}
