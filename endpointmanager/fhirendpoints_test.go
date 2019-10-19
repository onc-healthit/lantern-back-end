package main

import (
	"database/sql"
	"fmt"

	"testing"

	_ "github.com/lib/pq"
)

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

	var endpoint1 = &FHIREndpoint{URL: "example.com/FHIR/DSTU2",
		FHIRVersion:           "DSTU2",
		AuthorizationStandard: "OAuth 2.0"}
	var endpoint2 = &FHIREndpoint{URL: "other.example.com/FHIR/DSTU2",
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

	e1, err := GetFHIREndpoint(endpoint1.URL)
	if err != nil {
		t.Errorf("Error getting fhir endpoint: %s", err.Error())
	}
	if !e1.Equal(endpoint1) {
		t.Errorf("retrieved endpoint is not equal to saved endpoint.")
	}

	e2, err := GetFHIREndpoint(endpoint2.URL)
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

	e1, err = GetFHIREndpoint(endpoint1.URL)
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

	e2, err = GetFHIREndpoint(endpoint2.URL) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("endpoint2 no longer exists in DB after deleting endpoint1: %s", err.Error())
	}

	err = endpoint2.Delete()
	if err != nil {
		t.Errorf("Error deleting fhir endpoint: %s", err.Error())
	}
}

func Test_PersistHealthITProduct(t *testing.T) {
	var err error

	var hitp1 = &HealthITProduct{
		Name:                 "Health IT System 1",
		Version:              "1.0",
		Developer:            "Epic",
		APISyntax:            "FHIR R4",
		CertificationEdition: "2015"}
	var hitp2 = &HealthITProduct{
		Name:                 "Health IT System 2",
		Version:              "2.0",
		Developer:            "Cerner",
		APISyntax:            "FHIR DSTU2",
		CertificationEdition: "2014"}

	db = connectToDB(t)
	defer db.Close()

	// add products

	err = hitp1.Add()
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}

	err = hitp2.Add()
	if err != nil {
		t.Errorf("Error adding health it product: %s", err.Error())
	}

	// retrieve products

	h1, err := GetHealthITProduct(hitp1.Name, hitp1.Version)
	if err != nil {
		t.Errorf("Error getting health it product: %s", err.Error())
	}
	if !h1.Equal(hitp1) {
		t.Errorf("retrieved product is not equal to saved product.")
	}

	h2, err := GetHealthITProduct(hitp2.Name, hitp2.Version)
	if err != nil {
		t.Errorf("Error getting health it product: %s", err.Error())
	}
	if !h2.Equal(hitp2) {
		t.Errorf("retrieved product is not equal to saved product.")
	}

	// update product

	h1.APISyntax = "FHIR R5"

	err = h1.Update()
	if err != nil {
		t.Errorf("Error updating health it product: %s", err.Error())
	}

	h1, err = GetHealthITProduct(hitp1.Name, hitp1.Version)
	if err != nil {
		t.Errorf("Error getting health it product: %s", err.Error())
	}
	if h1.Equal(hitp1) {
		t.Errorf("retrieved UPDATED product is equal to original product.")
	}
	if h1.UpdatedAt.Equal(h1.CreatedAt) {
		t.Errorf("UpdatedAt is not being properly set on update.")
	}

	// delete products

	err = hitp1.Delete()
	if err != nil {
		t.Errorf("Error deleting health it product: %s", err.Error())
	}

	h2, err = GetHealthITProduct(hitp2.Name, hitp2.Version) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("hitp2 no longer exists in DB after deleting hitp1: %s", err.Error())
	}

	err = hitp2.Delete()
	if err != nil {
		t.Errorf("Error deleting health it product: %s", err.Error())
	}
}

func Test_PersistProviderOrganization(t *testing.T) {
	var err error

	var po1 = &ProviderOrganization{
		Name:             "Hospital #1 of America",
		URL:              "hospital.example.com",
		OrganizationType: "hospital",
		HospitalType:     "Acute Care",
		Ownership:        "Volunary non-profit",
		Beds:             250}
	var po2 = &ProviderOrganization{
		Name:             "Group Practice #1 of America",
		URL:              "grouppractice.example.com",
		OrganizationType: "group practice",
		HospitalType:     "",
		Ownership:        "",
		Beds:             -1}

	db = connectToDB(t)
	defer db.Close()

	// add organizations

	err = po1.Add()
	if err != nil {
		t.Errorf("Error adding provider organization: %s", err.Error())
	}

	err = po2.Add()
	if err != nil {
		t.Errorf("Error adding provider organization: %s", err.Error())
	}

	// retrieve organizations

	p1, err := GetProviderOrganization(po1.OrganizationID)
	if err != nil {
		t.Errorf("Error getting provider organization: %s", err.Error())
	}
	if !p1.Equal(po1) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	p2, err := GetProviderOrganization(po2.OrganizationID)
	if err != nil {
		t.Errorf("Error getting provider organization: %s", err.Error())
	}
	if !p2.Equal(po2) {
		t.Errorf("retrieved organization is not equal to saved organization.")
	}

	// update organization

	p1.HospitalType = "Critical Access"

	err = p1.Update()
	if err != nil {
		t.Errorf("Error updating provider organization: %s", err.Error())
	}

	p1, err = GetProviderOrganization(po1.OrganizationID)
	if err != nil {
		t.Errorf("Error getting provider organization: %s", err.Error())
	}
	if p1.Equal(po1) {
		t.Errorf("retrieved UPDATED organization is equal to original organization.")
	}
	if p1.UpdatedAt.Equal(p1.CreatedAt) {
		t.Errorf("UpdatedAt is not being properly set on update.")
	}

	// delete organizations

	err = po1.Delete()
	if err != nil {
		t.Errorf("Error deleting provider organization: %s", err.Error())
	}

	p2, err = GetProviderOrganization(po2.OrganizationID) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("po2 no longer exists in DB after deleting po1: %s", err.Error())
	}

	err = po2.Delete()
	if err != nil {
		t.Errorf("Error deleting provider organization: %s", err.Error())
	}
}
