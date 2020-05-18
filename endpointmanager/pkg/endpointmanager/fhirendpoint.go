package endpointmanager

import (
	"time"
	"strings"
)

// FHIREndpoint represents a fielded FHIR API endpoint hosted by a
// HealthITProduct and populated by a ProviderOrganization.
// Information about the FHIR API endpoint is populated by the FHIR
// capability statement found at that endpoint as well as information
// discovered about the IP address of the endpoint.
type FHIREndpoint struct {
	ID               int
	URL              string
	OrganizationName string
	ListSource       string
	CreatedAt        time.Time
	UpdatedAt        time.Time
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
	if e.OrganizationName != e2.OrganizationName {
		return false
	}
	if e.ListSource != e2.ListSource {
		return false
	}

	return true
}

// Prepends url with https://www. or https:// and appends with metadata/ if needed
func NormalizeURL(url string) string{
	normalized := ""
    // for cases such as foobar.com
    if !strings.HasPrefix(url, "https://www.") && !strings.HasPrefix(url, "http://www.")  {
        normalized = "https://www." + url
    }
    // for cases such as www.foobar.com
    if strings.HasPrefix(url, "www.") {
        normalized = "https://" +  url
	}

	// for cases such as foobar.com/
	if !strings.HasSuffix(url, "/metadata") && !strings.HasSuffix(url, "/metadata/") {
		if !strings.HasSuffix(url, "/") {
			normalized = normalized + "/"
		}
		normalized = normalized + "metadata"
	}
    return normalized
}
