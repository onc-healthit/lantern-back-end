package main

import (
	"time"

	"testing"

	_ "github.com/lib/pq"
)

func Test_PersistHealthITProduct(t *testing.T) {
	var err error

	var hitp1 = &HealthITProduct{
		Name:      "Health IT System 1",
		Version:   "1.0",
		Developer: "Epic",
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		AuthorizationStandard: "OAuth 2.0",
		APISyntax:             "FHIR R4",
		APIURL:                "example.com",
		CertificationCriteria: []string{"criteria1", "criteria2"},
		CertificationStatus:   "Active",
		CertificationDate:     time.Date(2019, 10, 19, 0, 0, 0, 0, time.UTC),
		CertificationEdition:  "2015",
		LastModifiedInCHPL:    time.Date(2019, 10, 19, 0, 0, 0, 0, time.UTC),
		CHPLID:                "ID"}
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

	h1, err := GetHealthITProduct(hitp1.GetID())
	if err != nil {
		t.Errorf("Error getting health it product: %s", err.Error())
	}
	if !h1.Equal(hitp1) {
		t.Errorf("retrieved product is not equal to saved product.")
	}

	h2, err := GetHealthITProductUsingNameAndVersion(hitp2.Name, hitp2.Version)
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

	h1, err = GetHealthITProduct(hitp1.GetID())
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

	_, err = GetHealthITProduct(hitp2.GetID()) // ensure we haven't deleted all entries
	if err != nil {
		t.Errorf("hitp2 no longer exists in DB after deleting hitp1: %s", err.Error())
	}

	err = hitp2.Delete()
	if err != nil {
		t.Errorf("Error deleting health it product: %s", err.Error())
	}
}

func Test_HealthITProductEqual(t *testing.T) {
	now := time.Now()
	var hitp1 = &HealthITProduct{
		id:        1,
		Name:      "Health IT System 1",
		Version:   "1.0",
		Developer: "Epic",
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		AuthorizationStandard: "OAuth 2.0",
		APISyntax:             "FHIR R4",
		APIURL:                "example.com",
		CertificationCriteria: []string{"criteria1", "criteria2"},
		CertificationStatus:   "Active",
		CertificationDate:     now,
		CertificationEdition:  "2015",
		LastModifiedInCHPL:    now,
		CHPLID:                "ID"}
	var hitp2 = &HealthITProduct{
		id:        1,
		Name:      "Health IT System 1",
		Version:   "1.0",
		Developer: "Epic",
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		AuthorizationStandard: "OAuth 2.0",
		APISyntax:             "FHIR R4",
		APIURL:                "example.com",
		CertificationCriteria: []string{"criteria1", "criteria2"},
		CertificationStatus:   "Active",
		CertificationDate:     now,
		CertificationEdition:  "2015",
		LastModifiedInCHPL:    now,
		CHPLID:                "ID"}

	if !hitp1.Equal(hitp2) {
		t.Errorf("Expected hitp1 to equal hitp2. They are not equal.")
	}

	hitp2.id = 2
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. id should be different. %d vs %d", hitp1.id, hitp2.id)
	}
	hitp2.id = hitp1.id

	hitp2.Name = "other"
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. Name should be different. %s vs %s", hitp1.Name, hitp2.Name)
	}
	hitp2.Name = hitp1.Name

	hitp2.Version = "other"
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. Version should be different. %s vs %s", hitp1.Version, hitp2.Version)
	}
	hitp2.Version = hitp1.Version

	hitp2.Developer = "other"
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. Developer should be different. %s vs %s", hitp1.Developer, hitp2.Developer)
	}
	hitp2.Developer = hitp1.Developer

	hitp2.Location.Address1 = "other"
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. Location.Address1 should be different. %s vs %s", hitp1.Location.Address1, hitp2.Location.Address1)
	}
	hitp2.Location.Address1 = hitp1.Location.Address1

	hitp2.AuthorizationStandard = "other"
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. AuthorizationStandard should be different. %s vs %s", hitp1.AuthorizationStandard, hitp2.AuthorizationStandard)
	}
	hitp2.AuthorizationStandard = hitp1.AuthorizationStandard

	hitp2.APISyntax = "other"
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. APISyntax should be different. %s vs %s", hitp1.APISyntax, hitp2.APISyntax)
	}
	hitp2.APISyntax = hitp1.APISyntax

	hitp2.APIURL = "other"
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. APIURL should be different. %s vs %s", hitp1.APIURL, hitp2.APIURL)
	}
	hitp2.APIURL = hitp1.APIURL

	hitp2.CertificationCriteria[0] = "other"
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. CertificationCriteria should be different. %s vs %s", hitp1.CertificationCriteria[0], hitp2.CertificationCriteria[0])
	}
	hitp2.CertificationCriteria[0] = hitp1.CertificationCriteria[0]

	hitp2.CertificationStatus = "other"
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. CertificationStatus should be different. %s vs %s", hitp1.CertificationStatus, hitp2.CertificationStatus)
	}
	hitp2.CertificationStatus = hitp1.CertificationStatus

	hitp2.CertificationDate = hitp2.CertificationDate.Add(500)
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. CertificationDate should be different.")
	}
	hitp2.CertificationDate = hitp1.CertificationDate

	hitp2.CertificationEdition = "other"
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. CertificationEdition should be different. %s vs %s", hitp1.CertificationEdition, hitp2.CertificationEdition)
	}
	hitp2.CertificationEdition = hitp1.CertificationEdition

	hitp2.LastModifiedInCHPL = hitp2.LastModifiedInCHPL.Add(500)
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. LastModifiedInCHPL should be different.")
	}
	hitp2.LastModifiedInCHPL = hitp1.LastModifiedInCHPL

	hitp2.CHPLID = "other"
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. CHPLID should be different. %s vs %s", hitp1.CHPLID, hitp2.CHPLID)
	}
	hitp2.CHPLID = hitp1.CHPLID

	hitp2.Name = "other"
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. Name should be different. %s vs %s", hitp1.Name, hitp2.Name)
	}
	hitp2.Name = hitp1.Name

	hitp2 = nil
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect hitp1 to equal nil hitp 2.")
	}
	hitp2 = hitp1

	hitp1 = nil
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect nil hitp1 to equal hitp 2.")
	}

	hitp2 = nil
	if !hitp1.Equal(hitp2) {
		t.Errorf("Nil hitp 1 should equal nil hitp 2.")
	}
}
