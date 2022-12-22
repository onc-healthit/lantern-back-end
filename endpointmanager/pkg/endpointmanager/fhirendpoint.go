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
	OrgDatabaseMapID int
	OrganizationList []*FHIREndpointOrganization
	ListSource       string
	VersionsResponse versionsoperatorparser.VersionsResponse
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type FHIREndpointOrganization struct {
	ID                  int
	OrganizationName    string
	OrganizationZipCode string
	OrganizationNPIID   string
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

	if !organizationListEquals(e.OrganizationList, e2.OrganizationList) {
		return false
	}

	return true
}

func organizationListEquals(orgList1 []*FHIREndpointOrganization, orgList2 []*FHIREndpointOrganization) bool {

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
	if o == nil && o == nil {
		return true
	} else if o == nil {
		return false
	} else if o == nil {
		return false
	}

	if o.OrganizationName != o.OrganizationName {
		return false
	}
	if o.OrganizationZipCode != o.OrganizationZipCode {
		return false
	}
	if o.OrganizationNPIID != o.OrganizationNPIID {
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

// OrganizationsToAdd adds the Organizations to the endpoint's Organization list if they are not present already, and returns all the organizations that need to be added to the db.
func (e *FHIREndpoint) OrganizationsToAdd(orgList []*FHIREndpointOrganization) []*FHIREndpointOrganization {

	newOrgList := orgList
	existingOrgList := e.OrganizationList

	var newOrganizations []*FHIREndpointOrganization
	for _, org := range newOrgList {
		found := containsOrganization(existingOrgList, org)
		if !found {

			organizationEntry := FHIREndpointOrganization{
				OrganizationName:    org.OrganizationName,
				OrganizationNPIID:   org.OrganizationNPIID,
				OrganizationZipCode: org.OrganizationZipCode,
			}

			e.OrganizationList = append(e.OrganizationList, &organizationEntry)
			newOrganizations = append(newOrganizations, &organizationEntry)
		}
	}
	return newOrganizations
}

// OrganizationsToRemove removes the Organizations to the endpoint's Organization list if they are not present in the new list, and returns all the organizations that need to be removed from the db.
func (e *FHIREndpoint) OrganizationsToRemove(orgList[]*FHIREndpointOrganization) []*FHIREndpointOrganization {
	newOrgList := orgList
	existingOrgList := e.OrganizationList
	

	var oldOrganizations []*FHIREndpointOrganization
	for index, org := range existingOrgList {
		found := containsOrganization(newOrgList, org)
		if !found {
			organizationListLength := len(e.OrganizationList)
			if index < organizationListLength - 1 {
				e.OrganizationList = append(e.OrganizationList[:index], e.OrganizationList[index+1:]...)
			} else {
				e.OrganizationList = append(e.OrganizationList[:organizationListLength-1])
			}
			oldOrganizations = append(oldOrganizations, org)
		}
	}
	return oldOrganizations
}

// containsOrganization checks if organization list contains the specified organization
func containsOrganization(orgList []*FHIREndpointOrganization, org *FHIREndpointOrganization) bool {
	found := false
	
	for _, o := range orgList {
		if org.OrganizationName == o.OrganizationName && org.OrganizationNPIID == o.OrganizationNPIID && org.OrganizationZipCode == o.OrganizationZipCode {
			found = true
			break
		}
	}
	return found
}

// sortOrganizationList sorts an endpoint's list of Organizations
func sortOrganizationList(orgList []*FHIREndpointOrganization) {
	sort.Slice(orgList, func(i, j int) bool {
		return orgList[i].OrganizationName < orgList[j].OrganizationName
	})
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
