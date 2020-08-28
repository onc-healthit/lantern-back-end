package endpointmanager

import (
	"time"

	"testing"

	_ "github.com/lib/pq"
)

func Test_HealthITProductEqual(t *testing.T) {
	now := time.Now()
	var hitp1 = &HealthITProduct{
		ID:       1,
		Name:     "Health IT System 1",
		Version:  "1.0",
		VendorID: 2,
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		AuthorizationStandard: "OAuth 2.0",
		APISyntax:             "FHIR R4",
		APIURL:                "example.com",
		CertificationCriteria: makeTestCrit([]int{1, 2}),
		CertificationStatus:   "Active",
		CertificationDate:     now,
		CertificationEdition:  "2015",
		LastModifiedInCHPL:    now,
		CHPLID:                "ID"}
	var hitp2 = &HealthITProduct{
		ID:       1,
		Name:     "Health IT System 1",
		Version:  "1.0",
		VendorID: 2,
		Location: &Location{
			Address1: "123 Gov Way",
			Address2: "Suite 123",
			City:     "A City",
			State:    "AK",
			ZipCode:  "00000"},
		AuthorizationStandard: "OAuth 2.0",
		APISyntax:             "FHIR R4",
		APIURL:                "example.com",
		CertificationCriteria: makeTestCrit([]int{1, 2}),
		CertificationStatus:   "Active",
		CertificationDate:     now,
		CertificationEdition:  "2015",
		LastModifiedInCHPL:    now,
		CHPLID:                "ID"}

	if !hitp1.Equal(hitp2) {
		t.Errorf("Expected hitp1 to equal hitp2. They are not equal.")
	}

	hitp2.ID = 2
	if !hitp1.Equal(hitp2) {
		t.Errorf("Expect healthit product 1 to equal healthit product 2. ids should be ignored. %d vs %d", hitp1.ID, hitp2.ID)
	}
	hitp2.ID = hitp1.ID

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

	hitp2.VendorID = 3
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. Developer should be different. %d vs %d", hitp1.VendorID, hitp2.VendorID)
	}
	hitp2.VendorID = hitp1.VendorID

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

	hitp2.CertificationCriteria = makeTestCrit([]int{10, 2})
	if hitp1.Equal(hitp2) {
		t.Errorf("Did not expect healthit product 1 to equal healthit product 2. CertificationCriteria should be different. %d vs %d", hitp1.CertificationCriteria[0], hitp2.CertificationCriteria[0])
	}
	hitp2.CertificationCriteria = makeTestCrit([]int{1, 2})

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
	if !hitp1.Equal(hitp2) {
		t.Errorf("Expected healthit product 1 to equal healthit product 2. 'Equals' should ignore CHPLID.")
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

func makeTestCrit(critIDs []int) []interface{} {
	critInt := make([]interface{}, len(critIDs))
	for i, v := range critIDs {
		critInt[i] = v
	}
	return critInt
}
