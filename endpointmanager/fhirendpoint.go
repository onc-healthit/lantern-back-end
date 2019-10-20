package main

import (
	"time"

	_ "github.com/lib/pq"
)

// FHIREndpoint represents a fielded FHIR API endpoint hosted by a
// HealthITProduct and populated by a ProviderOrganization.
// Information about the FHIR API endpoint is populated by the FHIR
// capability statement found at that endpoint as well as information
// discovered about the IP address of the endpoint.
type FHIREndpoint struct {
	id                    int
	URL                   string
	FHIRVersion           string
	AuthorizationStandard string    // examples: OAuth 2.0, Basic, etc.
	Location              *Location // location of the FHIR API endpoint's IP address from ipstack.com.
	Metadata              string    // the JSON representation of the FHIR capability statement
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// GetFHIREndpoint gets a FHIREndpoint from the database using the database id as a key.
func GetFHIREndpoint(id int) (*FHIREndpoint, error) {
	// TODO: missing metadata and location.
	sqlStatement := `SELECT id,
							url,
							fhir_version,
							authorization_standard,
							created_at,
							updated_at
					FROM fhir_endpoints WHERE id=$1`
	row := db.QueryRow(sqlStatement, id)
	var endpoint FHIREndpoint

	err := row.Scan(
		&endpoint.id,
		&endpoint.URL,
		&endpoint.FHIRVersion,
		&endpoint.AuthorizationStandard,
		&endpoint.CreatedAt,
		&endpoint.UpdatedAt)

	return &endpoint, err
}

// GetFHIREndpointUsingURL gets a FHIREndpoint from the database using the given url as a key.
func GetFHIREndpointUsingURL(url string) (*FHIREndpoint, error) {
	// TODO: missing metadata and location.
	sqlStatement := `SELECT id,
							url,
							fhir_version,
							authorization_standard,
							created_at,
							updated_at
					FROM fhir_endpoints WHERE url=$1`
	row := db.QueryRow(sqlStatement, url)
	var endpoint FHIREndpoint

	err := row.Scan(
		&endpoint.id,
		&endpoint.URL,
		&endpoint.FHIRVersion,
		&endpoint.AuthorizationStandard,
		&endpoint.CreatedAt,
		&endpoint.UpdatedAt)

	return &endpoint, err
}

// GetID returns the database ID for the FHIREndpoint.
func (e *FHIREndpoint) GetID() int {
	return e.id
}

// Add adds the FHIREndpoint to the database.
func (e *FHIREndpoint) Add() error {
	// TODO: missing metadata and location.
	sqlStatement := `
	INSERT INTO fhir_endpoints (url,
		fhir_version,
		authorization_standard)
	VALUES ($1, $2, $3)
	RETURNING id`

	row := db.QueryRow(sqlStatement,
		e.URL,
		e.FHIRVersion,
		e.AuthorizationStandard)

	err := row.Scan(&e.id)

	return err
}

// Update updates the FHIREndpoint in the database using the FHIREndpoint's URL as the key.
func (e *FHIREndpoint) Update() error {
	// TODO: missing metadata and location.
	sqlStatement := `
	UPDATE fhir_endpoints
	SET url = $1,
		fhir_version = $2,
		authorization_standard = $3
	WHERE id = $4`

	_, err := db.Exec(sqlStatement,
		e.URL,
		e.FHIRVersion,
		e.AuthorizationStandard,
		e.id)

	return err
}

// Delete deletes the FHIREndpoint from the databse using the FHIREndpoint's URL as the key.
func (e *FHIREndpoint) Delete() error {
	sqlStatement := `
	DELETE FROM fhir_endpoints
	WHERE id = $1`

	_, err := db.Exec(sqlStatement, e.id)

	return err
}

// Equal checks each field of the two FHIREndpoints except for the CreatedAt and UpdatedAt fields to see if they are equal.
func (e *FHIREndpoint) Equal(e2 *FHIREndpoint) bool {
	if e == nil && e2 == nil {
		return true
	} else if e == nil {
		return false
	} else if e2 == nil {
		return false
	}

	if e.id != e2.id {
		return false
	}
	if e.URL != e2.URL {
		return false
	}
	if e.FHIRVersion != e2.FHIRVersion {
		return false
	}
	if e.AuthorizationStandard != e2.AuthorizationStandard {
		return false
	}
	if !e.Location.Equal(e2.Location) {
		return false
	}
	if e.Metadata != e2.Metadata {
		return false
	}

	return true
}
