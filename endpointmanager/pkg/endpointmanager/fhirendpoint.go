package endpointmanager

import (
	"sort"
	"strings"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/versionsoperatorparser"
)

// FHIREndpoint represents a fielded FHIR API endpoint hosted by a
// HealthITProduct and populated by a ProviderOrganization.
// Information about the FHIR API endpoint is populated by the FHIR
// capability statement found at that endpoint as well as information
// discovered about the IP address of the endpoint.
type FHIREndpoint struct {
	ID               int
	URL              string
	OrganizationList []*FHIREndpointOrganization
	ListSource       string
	VersionsResponse versionsoperatorparser.VersionsResponse
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type FHIREndpointOrganization struct {
	ID                      int
	OrganizationName        string
	OrganizationZipCode     string
	OrganizationNPIID       string
	OrganizationIdentifiers []interface{}
	OrganizationAddresses   []interface{}
	OrganizationActive      string
	UpdatedAt               time.Time
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
	if !e.VersionsResponse.Equal(e2.VersionsResponse) {
		return false
	}
	if e.ListSource != e2.ListSource {
		return false
	}

	if !OrganizationListEquals(e.OrganizationList, e2.OrganizationList) {
		return false
	}

	return true
}

// Checks if the two lists of organizations are equal
func OrganizationListEquals(orgList1 []*FHIREndpointOrganization, orgList2 []*FHIREndpointOrganization) bool {

	if len(orgList1) != len(orgList2) {
		return false
	}

	sortOrganizationList(orgList1)
	sortOrganizationList(orgList2)

	for i := 0; i < len(orgList1); i++ {
		equals := orgList1[i].Equal(orgList2[i])
		if !equals {
			return false
		}
	}

	return true

}

// Equal checks each field of the two FHIREndpointOrganizations except for the database ID to see if they are equal.
func (o *FHIREndpointOrganization) Equal(o2 *FHIREndpointOrganization) bool {
	if o == nil && o2 == nil {
		return true
	} else if o == nil {
		return false
	} else if o2 == nil {
		return false
	}

	if o.OrganizationName != o2.OrganizationName {
		return false
	}
	if o.OrganizationZipCode != o2.OrganizationZipCode {
		return false
	}
	if o.OrganizationNPIID != o2.OrganizationNPIID {
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

// sortOrganizationList sorts an endpoint's list of Organizations
func sortOrganizationList(orgList []*FHIREndpointOrganization) {
	sort.Slice(orgList, func(i, j int) bool {
		return orgList[i].OrganizationName < orgList[j].OrganizationName
	})
}

// Gets all the NPI IDs for an endpoint
func (e *FHIREndpoint) GetNPIIDs() []string {
	var NPIIDs []string
	for _, org := range e.OrganizationList {
		NPIIDs = append(NPIIDs, org.OrganizationNPIID)
	}

	return NPIIDs
}

// Gets all the organization names for an endpoint
func (e *FHIREndpoint) GetOrganizationNames() []string {
	var OrganizationNames []string
	for _, org := range e.OrganizationList {
		OrganizationNames = append(OrganizationNames, org.OrganizationName)
	}

	return OrganizationNames
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
