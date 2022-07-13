package chplendpointquerier

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type FHIRBundle struct {
	Entries []BundleEntry `json:"entry"`
}

type BundleEntry struct {
	Resource BundleResource `json:"resource"`
}

type BundleResource struct {
	URL         string               `json:"address"`
	Name        string               `json:"name"`
	ManagingOrg ManagingOrgReference `json:"managingOrganization"`
	Orgs        []Organization       `json:"contained"`
}

type ManagingOrgReference struct {
	Reference string `json:"reference"`
	Display   string `json:"display"`
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

		entry.URL = bundleEntry.Resource.URL
		if bundleEntry.Resource.Name == "" {
			if bundleEntry.Resource.ManagingOrg.Display == "" {
				orgId := bundleEntry.Resource.ManagingOrg.Reference
				for _, org := range bundleEntry.Resource.Orgs {
					if org.Id == orgId {
						entry.OrganizationName = org.Name
					}
				}
			} else {
				entry.OrganizationName = bundleEntry.Resource.ManagingOrg.Display
			}
		} else {
			entry.OrganizationName = bundleEntry.Resource.Name
		}

		log.Info(entry.OrganizationName)

		lanternEntryList = append(lanternEntryList, entry)
	}

	return lanternEntryList
}
