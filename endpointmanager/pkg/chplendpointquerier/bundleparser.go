package chplendpointquerier

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"strings"
)

type FHIRBundle struct {
	Entries []BundleEntry `json:"entry"`
}

type BundleEntry struct {
	Resource BundleResource `json:"resource"`
}

type BundleResource struct {
	URL          interface{}          `json:"address"`
	Name         string               `json:"name"`
	ManagingOrg  ManagingOrgReference `json:"managingOrganization"`
	Orgs         []Organization       `json:"contained"`
	ResourceType string               `json:"resourceType"`
}

type ManagingOrgReference struct {
	Reference string `json:"reference"`
	Display   string `json:"display"`
	Id        string `json:"id"`
}

type Organization struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func BundleToLanternFormat(bundle []byte) []LanternEntry {
	var lanternEntryList []LanternEntry

	var structBundle FHIRBundle
	err := json.Unmarshal(bundle, &structBundle)
	if err != nil {
		log.Fatal(err)
	}

	for _, bundleEntry := range structBundle.Entries {
		var entry LanternEntry

		if bundleEntry.Resource.ResourceType == "Endpoint" {
			entryURL := bundleEntry.Resource.URL.(string)
			// Do not add entries that do not have URLs
			if entryURL != "" {
				entry.URL = strings.TrimSpace(entryURL)
				if bundleEntry.Resource.Name == "" {
					if bundleEntry.Resource.ManagingOrg.Display == "" {

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
					} else {
						entry.OrganizationName = strings.TrimSpace(bundleEntry.Resource.ManagingOrg.Display)
					}
				} else {
					entry.OrganizationName = strings.TrimSpace(bundleEntry.Resource.Name)
				}

				lanternEntryList = append(lanternEntryList, entry)
			}
		}
	}

	return lanternEntryList
}
