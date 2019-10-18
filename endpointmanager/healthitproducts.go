package main

import "time"

// HealthITProduct represents a health IT vendor product such as an
// EHR. This information is gathered from the Certified Health IT Products List
// (CHPL).
type HealthITProduct struct {
	Name                  string
	Version               string
	Developer             string   // the name of the vendor that creates the product.
	Location              Location // the address listed in CHPL for the Developer.
	AuthorizationStandard string   // examples: OAuth 2.0, Basic, etc.
	APISyntax             string   // the format of the information provided by the API, for example, REST, FHIR STU3, etc.
	APIURL                string   // the URL to the API documentation for the product.
	CertificationCriteria []string // the ONC criteria that the product was certified to, for example, ["170.315 (g)(7)", "170.315 (g)(8)", "170.315 (g)(9)"]
	CertificationStatus   string   // the ONC certification status, for example, "Active", "Retired", "Suspended by ONC", etc.
	CertificationDate     time.Time
	CertificationEdition  string // the product's certification edition for the ONC Health IT certification program, for example, "2014", "2015".
	LastModifiedInCHPL    time.Time
	CHPLID                string // the product's unique ID within the CHPL system.
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
