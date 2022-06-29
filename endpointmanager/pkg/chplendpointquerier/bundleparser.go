package chplendpointquerier

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type FHIRBundle struct {
	Entries []BundleEntry `json:"entry"`
}

type BundleEntry struct {
	Resource BundleResource  `json:"resource"`
}

type BundleResource struct {
	URL string `json:"address"`
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
		entry.OrganizationName = bundleEntry.Resource.Name

		lanternEntryList = append(lanternEntryList, entry)
	}

	return lanternEntryList
}