package main

import (
	"database/sql"
	"fmt"

	"testing"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func connectToDB(t *testing.T) *sql.DB {
	setupConfig()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		viper.GetString("dbhost"),
		viper.GetInt("dbport"),
		viper.GetString("dbuser"),
		viper.GetString("dbpass"),
		viper.GetString("dbname"),
		viper.GetString("dbsslmode"))

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		t.Fatalf("Error opening database: %s", err.Error())
	}

	// calling db.Ping to create a connection to the database.
	// db.Open only validates the arguments, it does not create the connection.
	err = db.Ping()
	if err != nil {
		t.Fatalf("Error creating connection to database: %s", err.Error())
	}

	return db
}

func Test_PersistFHIREndpoint(t *testing.T) {
	var err error

	var endpoint1 = &FHIREndpoint{
		URL:                   "example.com/FHIR/DSTU2",
		FHIRVersion:           "DSTU2",
		AuthorizationStandard: "OAuth 2.0",
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		CapabilityStatement: &CapabilityStatement{}}
	var endpoint2 = &FHIREndpoint{
		URL:                   "other.example.com/FHIR/DSTU2",
		FHIRVersion:           "DSTU2",
		AuthorizationStandard: "R4 2.0"}

	db = connectToDB(t)
	defer db.Close()

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

	e1, err := GetFHIREndpoint(endpoint1.GetID())
	if err != nil {
		t.Errorf("Error getting fhir endpoint: %s", err.Error())
	}
	if !e1.Equal(endpoint1) {
		t.Errorf("retrieved endpoint is not equal to saved endpoint.")
	}

	e2, err := GetFHIREndpointUsingURL(endpoint2.URL)
	if err != nil {
		t.Errorf("Error getting fhir endpoint: %s", err.Error())
	}
	if !e2.Equal(endpoint2) {
		t.Errorf("retrieved endpoint is not equal to saved endpoint.")
	}

	// update endpoint

	e1.FHIRVersion = "Unknown"

	err = e1.Update()
	if err != nil {
		t.Errorf("Error updating fhir endpoint: %s", err.Error())
	}

	e1, err = GetFHIREndpoint(endpoint1.GetID())
	if err != nil {
		t.Errorf("Error getting fhir endpoint: %s", err.Error())
	}
	if e1.Equal(endpoint1) {
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

	_, err = GetFHIREndpoint(endpoint1.GetID()) // ensure we deleted the entry
	if err == nil {
		t.Errorf("endpoint1 was not deleted: %s", err.Error())
	}

	_, err = GetFHIREndpoint(endpoint2.GetID()) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("error retrieving endpoint2 after deleting endpoint1: %s", err.Error())
	}

	err = endpoint2.Delete()
	if err != nil {
		t.Errorf("Error deleting fhir endpoint: %s", err.Error())
	}
}

func Test_FHIREndpointEqual(t *testing.T) {
	var endpoint1 = &FHIREndpoint{
		id:                    1,
		URL:                   "example.com/FHIR/DSTU2",
		FHIRVersion:           "DSTU2",
		AuthorizationStandard: "OAuth 2.0",
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		CapabilityStatement: &CapabilityStatement{}}
	var endpoint2 = &FHIREndpoint{
		id:                    1,
		URL:                   "example.com/FHIR/DSTU2",
		FHIRVersion:           "DSTU2",
		AuthorizationStandard: "OAuth 2.0",
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		CapabilityStatement: &CapabilityStatement{}}

	if !endpoint1.Equal(endpoint2) {
		t.Errorf("Expected endpoint1 to equal endpoint2. They are not equal.")
	}

	endpoint2.id = 2
	if !endpoint1.Equal(endpoint2) {
		t.Errorf("Expect endpoint 1 to equal endpoint 2. ids should be ignored. %d vs %d", endpoint1.id, endpoint2.id)
	}
	endpoint2.id = endpoint1.id

	endpoint2.URL = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. URL should be different. %s vs %s", endpoint1.URL, endpoint2.URL)
	}
	endpoint2.URL = endpoint1.URL

	endpoint2.FHIRVersion = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. FHIRVersion should be different. %s vs %s", endpoint1.FHIRVersion, endpoint2.FHIRVersion)
	}
	endpoint2.FHIRVersion = endpoint1.FHIRVersion

	endpoint2.AuthorizationStandard = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. AuthorizationStandard should be different. %s vs %s", endpoint1.AuthorizationStandard, endpoint2.AuthorizationStandard)
	}
	endpoint2.AuthorizationStandard = endpoint1.AuthorizationStandard

	endpoint2.Location.Address1 = "other"
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. Location should be different. %s vs %s", endpoint1.Location.Address1, endpoint2.Location.Address1)
	}
	endpoint2.Location.Address1 = endpoint1.Location.Address1

	endpoint2.CapabilityStatement = nil
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal endpoint 2. CapabilityStatement should be different. %s vs %s", endpoint1.CapabilityStatement, endpoint2.CapabilityStatement)
	}
	endpoint2.CapabilityStatement = endpoint1.CapabilityStatement

	endpoint2 = nil
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect endpoint1 to equal nil endpoint 2.")
	}
	endpoint2 = endpoint1

	endpoint1 = nil
	if endpoint1.Equal(endpoint2) {
		t.Errorf("Did not expect nil endpoint1 to equal endpoint 2.")
	}

	endpoint2 = nil
	if !endpoint1.Equal(endpoint2) {
		t.Errorf("Nil endpoint 1 should equal nil endpoint 2.")
	}
}
