package main

import (
	"database/sql"
	"fmt"

	"testing"

	_ "github.com/lib/pq"
)

var endpoint1 = &FHIREndpoint{URL: "example.com/FHIR/DSTU2",
	FHIRVersion:           "DSTU2",
	AuthorizationStandard: "OAuth 2.0"}

var endpoint2 = &FHIREndpoint{URL: "other.example.com/FHIR/DSTU2",
	FHIRVersion:           "DSTU2",
	AuthorizationStandard: "R4 2.0"}

func connectToDB(t *testing.T) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		t.Errorf("Error opening database: %s", err.Error())
	}

	// calling db.Ping to create a connection to the database.
	// db.Open only validates the arguments, it does not create the connection.
	err = db.Ping()
	if err != nil {
		t.Errorf("Error creating connection to database: %s", err.Error())
	}

	return db
}

func Test_PersistFHIREndpoint(t *testing.T) {
	var err error
	db = connectToDB(t)

	// add endpoints

	err = endpoint1.Add()
	if err != nil {
		t.Errorf("Error adding fhir endpoint: %s", err.Error())
	}

	err = endpoint2.Add()
	if err != nil {
		t.Errorf("Error adding fhir endpoint: %s", err.Error())
	}

	// retrieve endpoints

	e1, err := GetFHIREndpoint(endpoint1.URL)
	if err != nil {
		t.Errorf("Error getting fhir endpoint: %s", err.Error())
	}
	if !e1.Equals(endpoint1) {
		t.Errorf("retrieved endpoint is not equal to saved endpoint.")
	}

	e2, err := GetFHIREndpoint(endpoint2.URL)
	if err != nil {
		t.Errorf("Error getting fhir endpoint: %s", err.Error())
	}
	if !e2.Equals(endpoint2) {
		t.Errorf("retrieved endpoint is not equal to saved endpoint.")
	}

	// update endpoint
	e1.FHIRVersion = "Unknown"

	err = e1.Update()
	if err != nil {
		t.Errorf("Error updating fhir endpoint: %s", err.Error())
	}
	err = e2.Update()
	if err != nil {
		t.Errorf("Error updating fhir endpoint: %s", err.Error())
	}

	e1, err = GetFHIREndpoint(endpoint1.URL)
	if err != nil {
		t.Errorf("Error getting fhir endpoint: %s", err.Error())
	}
	if e1.Equals(endpoint1) {
		t.Errorf("retrieved UPDATED endpoint is equal to original endpoint.")
	}
	if e1.UpdatedAt.Equal(e1.CreatedAt) {
		t.Errorf("UpdatedAt is not being properly set on update.")
	}

	// delete endpoints

	err = endpoint1.Delete()
	if err != nil {
		t.Errorf("Error deleting fhir endpoint: %s", err.Error())
	}

	e2, err = GetFHIREndpoint(endpoint2.URL) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("endpoint2 no longer exists in DB after deleting endpoint1: %s", err.Error())
	}

	err = endpoint2.Delete()
	if err != nil {
		t.Errorf("Error deleting fhir endpoint: %s", err.Error())
	}

	db.Close()
}
