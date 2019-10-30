package main

import (
	"encoding/json"
	"time"

	"github.com/google/go-cmp/cmp"
)

// HealthITProduct represents a health IT vendor product such as an
// EHR. This information is gathered from the Certified Health IT Products List
// (CHPL).
type HealthITProduct struct {
	id                    int
	Name                  string
	Version               string
	Developer             string    // the name of the vendor that creates the product.
	Location              *Location // the address listed in CHPL for the Developer.
	AuthorizationStandard string    // examples: OAuth 2.0, Basic, etc.
	APISyntax             string    // the format of the information provided by the API, for example, REST, FHIR STU3, etc.
	APIURL                string    // the URL to the API documentation for the product.
	CertificationCriteria []string  // the ONC criteria that the product was certified to, for example, ["170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"]
	CertificationStatus   string    // the ONC certification status, for example, "Active", "Retired", "Suspended by ONC", etc.
	CertificationDate     time.Time
	CertificationEdition  string // the product's certification edition for the ONC Health IT certification program, for example, "2014", "2015".
	LastModifiedInCHPL    time.Time
	CHPLID                string // the product's unique ID within the CHPL system.
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// GetHealthITProduct gets a HealthITProduct from the database using the database ID as a key.
// If the HealthITProduct does not exist in the database, sql.ErrNoRows will be returned.
func GetHealthITProduct(id int) (*HealthITProduct, error) {
	var hitp HealthITProduct
	var locationJSON []byte
	var certificationCriteriaJSON []byte

	sqlStatement := `
	SELECT
		id,
		name,
		version,
		developer,
		location,
		authorization_standard,
		api_syntax,
		api_url,
		certification_criteria,
		certification_status,
		certification_date,
		certification_edition,
		last_modified_in_chpl,
		chpl_id,
		created_at,
		updated_at
	FROM healthit_products WHERE id=$1`
	row := db.QueryRow(sqlStatement, id)

	err := row.Scan(
		&hitp.id,
		&hitp.Name,
		&hitp.Version,
		&hitp.Developer,
		&locationJSON,
		&hitp.AuthorizationStandard,
		&hitp.APISyntax,
		&hitp.APIURL,
		&certificationCriteriaJSON,
		&hitp.CertificationStatus,
		&hitp.CertificationDate,
		&hitp.CertificationEdition,
		&hitp.LastModifiedInCHPL,
		&hitp.CHPLID,
		&hitp.CreatedAt,
		&hitp.UpdatedAt)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &hitp.Location)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(certificationCriteriaJSON, &hitp.CertificationCriteria)

	return &hitp, err
}

// GetHealthITProductUsingNameAndVersion gets a HealthITProduct from the database using the healthit product's name and version as a key.
// If the HealthITProduct does not exist in the database, sql.ErrNoRows will be returned.
func GetHealthITProductUsingNameAndVersion(name string, version string) (*HealthITProduct, error) {
	var hitp HealthITProduct
	var locationJSON []byte
	var certificationCriteriaJSON []byte

	sqlStatement := `
	SELECT
		id,
		name,
		version,
		developer,
		location,
		authorization_standard,
		api_syntax,
		api_url,
		certification_criteria,
		certification_status,
		certification_date,
		certification_edition,
		last_modified_in_chpl,
		chpl_id,
		created_at,
		updated_at
	FROM healthit_products WHERE name=$1 AND version=$2`
	row := db.QueryRow(sqlStatement, name, version)

	err := row.Scan(
		&hitp.id,
		&hitp.Name,
		&hitp.Version,
		&hitp.Developer,
		&locationJSON,
		&hitp.AuthorizationStandard,
		&hitp.APISyntax,
		&hitp.APIURL,
		&certificationCriteriaJSON,
		&hitp.CertificationStatus,
		&hitp.CertificationDate,
		&hitp.CertificationEdition,
		&hitp.LastModifiedInCHPL,
		&hitp.CHPLID,
		&hitp.CreatedAt,
		&hitp.UpdatedAt)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(locationJSON, &hitp.Location)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(certificationCriteriaJSON, &hitp.CertificationCriteria)

	return &hitp, err
}

// GetID returns the database ID for the HealthITProduct.
func (hitp *HealthITProduct) GetID() int {
	return hitp.id
}

// Add adds the HealthITProduct to the database.
func (hitp *HealthITProduct) Add() error {
	sqlStatement := `
	INSERT INTO healthit_products (
		name,
		version,
		developer,
		location,
		authorization_standard,
		api_syntax,
		api_url,
		certification_criteria,
		certification_status,
		certification_date,
		certification_edition,
		last_modified_in_chpl,
		chpl_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	RETURNING id`

	locationJSON, err := json.Marshal(hitp.Location)
	if err != nil {
		return err
	}

	certificationCriteriaJSON, err := json.Marshal(hitp.CertificationCriteria)
	if err != nil {
		return err
	}

	row := db.QueryRow(sqlStatement,
		hitp.Name,
		hitp.Version,
		hitp.Developer,
		locationJSON,
		hitp.AuthorizationStandard,
		hitp.APISyntax,
		hitp.APIURL,
		certificationCriteriaJSON,
		hitp.CertificationStatus,
		hitp.CertificationDate,
		hitp.CertificationEdition,
		hitp.LastModifiedInCHPL,
		hitp.CHPLID)

	err = row.Scan(&hitp.id)

	return err
}

// Update updates the HealthITProduct in the database using the HealthITProduct's database ID as the key.
func (hitp *HealthITProduct) Update() error {
	sqlStatement := `
	UPDATE healthit_products
	SET name = $1,
		version = $2,
		developer = $3,
		authorization_standard = $4,
		api_syntax = $5,
		api_url = $6,
		certification_status = $7,
		certification_date = $8,
		certification_edition = $9,
		last_modified_in_chpl = $10,
		chpl_id = $11,
		location = $12,
		certification_criteria = $13
	WHERE id=$14`

	locationJSON, err := json.Marshal(hitp.Location)
	if err != nil {
		return err
	}

	certificationCriteriaJSON, err := json.Marshal(hitp.CertificationCriteria)
	if err != nil {
		return err
	}

	_, err = db.Exec(sqlStatement,
		hitp.Name,
		hitp.Version,
		hitp.Developer,
		hitp.AuthorizationStandard,
		hitp.APISyntax,
		hitp.APIURL,
		hitp.CertificationStatus,
		hitp.CertificationDate,
		hitp.CertificationEdition,
		hitp.LastModifiedInCHPL,
		hitp.CHPLID,
		locationJSON,
		certificationCriteriaJSON,
		hitp.id)

	return err
}

// Delete deletes the HealthITProduct from the database using the HealthITProduct's database ID as the key.
func (hitp *HealthITProduct) Delete() error {
	sqlStatement := `
	DELETE FROM healthit_products
	WHERE id=$1`

	_, err := db.Exec(sqlStatement, hitp.id)

	return err
}

// Equal checks each field of the two HealthITProducts except for the database ID, CreatedAt and UpdatedAt fields to see if they are equal.
func (hitp *HealthITProduct) Equal(hitp2 *HealthITProduct) bool {
	if hitp == nil && hitp2 == nil {
		return true
	} else if hitp == nil {
		return false
	} else if hitp2 == nil {
		return false
	}

	if hitp.Name != hitp2.Name {
		return false
	}
	if hitp.Version != hitp2.Version {
		return false
	}
	if hitp.Developer != hitp2.Developer {
		return false
	}
	if !hitp.Location.Equal(hitp2.Location) {
		return false
	}
	if hitp.AuthorizationStandard != hitp2.AuthorizationStandard {
		return false
	}
	if hitp.APISyntax != hitp2.APISyntax {
		return false
	}
	if hitp.APIURL != hitp2.APIURL {
		return false
	}
	if !cmp.Equal(hitp.CertificationCriteria, hitp2.CertificationCriteria) {
		return false
	}
	if hitp.CertificationStatus != hitp2.CertificationStatus {
		return false
	}
	if !hitp.CertificationDate.Equal(hitp2.CertificationDate) {
		return false
	}
	if hitp.CertificationEdition != hitp2.CertificationEdition {
		return false
	}
	if !hitp.LastModifiedInCHPL.Equal(hitp2.LastModifiedInCHPL) {
		return false
	}
	if hitp.CHPLID != hitp2.CHPLID {
		return false
	}

	return true
}
