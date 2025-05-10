package chplendpointquerier

import (
	"encoding/json"
	"strings"

	log "github.com/sirupsen/logrus"
)

type FHIRBundle struct {
	Entries []BundleEntry `json:"entry"`
}

type BundleEntry struct {
	Resource BundleResource `json:"resource"`
}

type BundleResource struct {
	Address      interface{}          `json:"address"`
	Name         string               `json:"name"`
	ManagingOrg  ManagingOrgReference `json:"managingOrganization"`
	Orgs         []Organization       `json:"contained"`
	ResourceType string               `json:"resourceType"`
	OrgId        string               `json:"id"`
}

type ManagingOrgReference struct {
	Reference string `json:"reference"`
	Display   string `json:"display"`
	Id        string `json:"id"`
}

type Organization struct {
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	Address      []Address `json:"address"`
	ResourceType string    `json:"resourceType"`
}

type Address struct {
	PostalCode string `json:"postalCode"`
}

func BundleToLanternFormat(bundle []byte, chplURL string) []LanternEntry {
	var lanternEntryList []LanternEntry
	var organizationZip = make(map[string]string)

	var structBundle FHIRBundle
	err := json.Unmarshal(bundle, &structBundle)
	if err != nil {
		log.Warn("Handler is required for url ", chplURL)
		log.Fatal("More details about the error: ", err)
	}
	for _, bundleEntry := range structBundle.Entries {
		if strings.EqualFold(strings.TrimSpace(bundleEntry.Resource.ResourceType), "Organization") {
			if bundleEntry.Resource.Address != nil {
				addressMapArr := bundleEntry.Resource.Address.([]interface{})
				for _, address := range addressMapArr {
					addressMap := address.(map[string]interface{})
					postalCode, ok := addressMap["postalCode"].(string)
					if ok {
						organizationZip[bundleEntry.Resource.OrgId] = postalCode
					}
				}
			}
		}
	}

	for _, bundleEntry := range structBundle.Entries {
		var entry LanternEntry

		if strings.EqualFold(strings.TrimSpace(bundleEntry.Resource.ResourceType), "Endpoint") {
			if bundleEntry.Resource.Address == nil {
				continue
			}
			entryURL := bundleEntry.Resource.Address.(string)
			// Do not add entries that do not have URLs
			if entryURL != "" {
				entry.URL = strings.TrimSpace(entryURL)
				if bundleEntry.Resource.ManagingOrg.Display == "" {
					if bundleEntry.Resource.Name != "" {
						entry.OrganizationName = strings.TrimSpace(bundleEntry.Resource.Name)
					} else {
						orgId := bundleEntry.Resource.ManagingOrg.Reference

						if orgId == "" {
							orgId = bundleEntry.Resource.ManagingOrg.Id
						}

						orgId = strings.TrimPrefix(orgId, "#")

						for _, org := range bundleEntry.Resource.Orgs {
							if org.Id == orgId {
								entry.OrganizationName = strings.TrimSpace(org.Name)
							}
						}
					}
				} else {
					entry.OrganizationName = strings.TrimSpace(bundleEntry.Resource.ManagingOrg.Display)
				}

				orgZipAdded := false

				for _, org := range bundleEntry.Resource.Orgs {
					if len(org.Address) > 0 {
						address := org.Address[0]
						if address.PostalCode != "" {
							entry.OrganizationZipCode = strings.TrimSpace(address.PostalCode)
							orgZipAdded = true
						}
					}

					if !orgZipAdded && len(organizationZip) > 0 {
						postalCode, ok := organizationZip[org.Id]
						if ok {
							entry.OrganizationZipCode = strings.TrimSpace(postalCode)
						}
					}
				}

				lanternEntryList = append(lanternEntryList, entry)
			}
		}
	}

	return lanternEntryList
}
