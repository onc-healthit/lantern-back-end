package main

// import (
// 	"database/sql"
// 	"fmt"

// 	_ "github.com/lib/pq" // specified to do this for accessing postgres db
// )

// func query(db *sql.DB, done chan struct{}) {
// 	rows, err := db.Query("SELECT 1")
// 	done <- struct{}{}
// 	if err == nil {
// 		err = rows.Close()
// 	}
// 	if err != nil {
// 		panic(err)
// 	}
// }

func main() {
	// psqlInfo := "host=localhost port=5432 user=lantern " +
	// 	"password=postgrespassword dbname=lantern sslmode=disable"
	// db, err := sql.Open("postgres", psqlInfo)

	// if err != nil {
	// 	panic(err)
	// }

	// // calling db.Ping to create a connection to the database.
	// // db.Open only validates the arguments, it does not create the connection.
	// err = db.Ping()
	// if err != nil {
	// 	err = fmt.Errorf("Error creating connection to database: %s", err.Error())
	// 	panic(err.Error())
	// }

	// println("Successfully connected to DB!")

	// //db.SetMaxOpenConns(1)
	// defer db.Close()
	// //done := make(chan struct{}, 1)
	// //go query(db, done)
	// //<-done
	// //time.Sleep(1200 * time.Millisecond)
	// //go query(db, done)
	// //<-done
	// //close(done)

	// stmt, err := db.Prepare("SELECT * FROM fhir_endpoints where $1;")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("%v\n", stmt)

	// stmt.Exec("http_response=0; DROP TABLE healthit_products")
}
