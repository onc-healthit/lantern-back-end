package endpointmanager

import (
	"strings"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/versionsoperatorparser"
)

// FHIREndpoint represents a fielded FHIR API endpoint hosted by a
// HealthITProduct and populated by a ProviderOrganization.
// Information about the FHIR API endpoint is populated by the FHIR
// capability statement found at that endpoint as well as information
// discovered about the IP address of the endpoint.
type FHIREndpoint struct {
	ID                int
	URL               string
	OrganizationNames []string
	NPIIDs            []string
	ListSource        string
	VersionsResponse  versionsoperatorparser.VersionsResponse
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Equal checks each field of the two FHIREndpoints except for the database ID, CreatedAt and UpdatedAt fields to see if they are equal.
func (e *FHIREndpoint) Equal(e2 *FHIREndpoint) bool {
	if e == nil && e2 == nil {
		return true
	} else if e == nil {
		return false
	} else if e2 == nil {
		return false
	}

	if e.URL != e2.URL {
		return false
	}
	if !helpers.StringArraysEqual(e.OrganizationNames, e2.OrganizationNames) {
		return false
	}
	if !helpers.StringArraysEqual(e.NPIIDs, e2.NPIIDs) {
		return false
	}
	if !e.VersionsResponse.Equal(e2.VersionsResponse) {
		return false
	}
	if e.ListSource != e2.ListSource {
		return false
	}

	return true
}

// Prepends url with https:// if needed
func NormalizeURL(url string) string {
	normalized := url
	// for cases such as foobar.com
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		normalized = "https://" + normalized
	}

	return normalized
}

// Prepends url with https:// and appends with metadata/ if needed
func NormalizeEndpointURL(url string) string {
	normalized := NormalizeURL(url)

	// for cases such as foobar.com/
	if !strings.HasSuffix(url, "/metadata") && !strings.HasSuffix(url, "/metadata/") {
		if !strings.HasSuffix(url, "/") {
			normalized = normalized + "/"
		}
		normalized = normalized + "metadata"
	}
	return normalized
}

// AddOrganizationName adds the name to the endpoint's OrganizationNames list if it's not present already. If it is, it does nothing.
func (e *FHIREndpoint) AddOrganizationName(orgName string) {
	if !helpers.StringArrayContains(e.OrganizationNames, orgName) {
		e.OrganizationNames = append(e.OrganizationNames, orgName)
	}
}

// AddNPIID adds the name to the endpoint's NPIIDs list if it's not present already. If it is, it does nothing.
func (e *FHIREndpoint) AddNPIID(npiID string) {
	if !helpers.StringArrayContains(e.NPIIDs, npiID) {
		e.NPIIDs = append(e.NPIIDs, npiID)
	}
}

// Prepends url with https:// and appends with .well-know/smart-configuration/ if needed
func NormalizeWellKnownURL(url string) string {
	normalized := NormalizeURL(url)

	if !strings.HasSuffix(url, "/.well-known/smart-configuration") && !strings.HasSuffix(url, "/.well-known/smart-configuration/") {
		if !strings.HasSuffix(url, "/") {
			normalized = normalized + "/"
		}
		normalized = normalized + ".well-known/smart-configuration"
	}
	return normalized
}

// Prepends url with https:// and appends with $versions if needed
func NormalizeVersionsURL(url string) string {
	normalized := NormalizeURL(url)

	if !strings.HasSuffix(url, "/$versions") && !strings.HasSuffix(url, "/$versions/") {
		if !strings.HasSuffix(url, "/") {
			normalized = normalized + "/"
		}
		normalized = normalized + "$versions"
	}
	return normalized
}
