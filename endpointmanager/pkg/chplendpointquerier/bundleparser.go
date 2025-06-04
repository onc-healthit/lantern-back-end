package chplendpointquerier

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type FHIRBundle struct {
	Entries []BundleEntry `json:"entry"`
}

type BundleEntry struct {
	Resource BundleResource `json:"resource"`
	FullURL  string         `json:"fullUrl"`
}

type BundleResource struct {
	Address      interface{}          `json:"address"`
	Identifier   interface{}          `json:"identifier"`
	Active       interface{}          `json:"active"`
	Name         string               `json:"name"`
	ManagingOrg  ManagingOrgReference `json:"managingOrganization"`
	Orgs         []Organization       `json:"contained"`
	ResourceType string               `json:"resourceType"`
	OrgId        string               `json:"id"`
	Endpoint     interface{}          `json:"endpoint"`
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

func containsOrgId(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func BundleToLanternFormat(bundle []byte, chplURL string) []LanternEntry {
	var lanternEntryList []LanternEntry

	var endpointOrgMap = make(map[string][]string)
	var organizationZip = make(map[string]string)
	var organizationName = make(map[string]string)
	var organizationURL = make(map[string]string)
	var organizationAddresses = make(map[string][]string)
	var organizationIdentifiers = make(map[string][]string)
	var organizationActive = make(map[string]string)
	var organizationNPI = make(map[string]string)
	var seenIdentifiers = make(map[string]bool)

	var structBundle FHIRBundle
	err := json.Unmarshal(bundle, &structBundle)
	if err != nil {
		log.Warn("Handler is required for url ", chplURL)
		log.Fatal("More details about the error: ", err)
	}
	for _, bundleEntry := range structBundle.Entries {
		if strings.EqualFold(strings.TrimSpace(bundleEntry.Resource.ResourceType), "Organization") {

			isIdentifierPresent := false
			var identifierKey string

			if bundleEntry.Resource.Identifier != nil {
				identifierArr := bundleEntry.Resource.Identifier.([]interface{})

				for _, identifier := range identifierArr {
					identifierMap := identifier.(map[string]interface{})
					var identifierCode string

					if identifierMap["system"] != nil && identifierMap["system"].(string) != "" {
						if identifierMap["system"].(string) == "http://hl7.org/fhir/sid/us-npi" ||
							identifierMap["system"].(string) == "http://hl7.org.fhir/sid/us-npi" {
							identifierCode = "NPI"
						} else if identifierMap["system"].(string) == "urn:oid:2.16.840.1.113883.4.7" {
							identifierCode = "CLIA"
						} else if identifierMap["system"].(string) == "urn:oid:2.16.840.1.113883.6.300" {
							identifierCode = "NAIC"
						} else {
							identifierCode = "Other"
						}

						if identifierMap["value"] != nil && identifierMap["value"].(string) != "" {
							identifierStr := identifierCode + ": " + identifierMap["value"].(string)

							if !isIdentifierPresent && !seenIdentifiers[identifierStr] {
								isIdentifierPresent = true
								identifierKey = identifierStr
							}

							seenIdentifiers[identifierStr] = true

							if !containsOrgId(organizationIdentifiers[identifierKey], identifierStr) {
								organizationIdentifiers[identifierKey] = append(organizationIdentifiers[identifierKey], identifierStr)
							}

							if identifierCode == "NPI" {
								organizationNPI[identifierKey] = identifierMap["value"].(string)
							}
						}
					}
				}
			}

			if isIdentifierPresent && bundleEntry.Resource.Endpoint != nil {
				endpointArr := bundleEntry.Resource.Endpoint.([]interface{})
				for _, endpoint := range endpointArr {
					endpointMap := endpoint.(map[string]interface{})
					if endpointMap["reference"] != nil && endpointMap["reference"].(string) != "" {
						endpointId := endpointMap["reference"].(string)
						endpointId = strings.TrimPrefix(endpointId, "Endpoint/")
						endpointId = strings.TrimPrefix(endpointId, "endpoint/")

						// Store endpoint-to-organizations mapping (if not already present)
						if !containsOrgId(endpointOrgMap[endpointId], identifierKey) {
							endpointOrgMap[endpointId] = append(endpointOrgMap[endpointId], identifierKey)
						}
					}
				}
			}

			if isIdentifierPresent && bundleEntry.Resource.Address != nil {
				addressMapArr := bundleEntry.Resource.Address.([]interface{})
				for _, address := range addressMapArr {
					addressMap := address.(map[string]interface{})

					// Get the values inside "line" array of the address
					var result []string
					if addressMap["line"] != nil {
						lineMap := addressMap["line"].([]interface{})
						for _, line := range lineMap {
							if line != nil {
								result = append(result, fmt.Sprintf("%v", line))
							}
						}
					}

					// Get the rest of the values in address
					if addressMap["city"] != nil {
						result = append(result, fmt.Sprintf("%v", addressMap["city"]))
					}

					if addressMap["state"] != nil {
						result = append(result, fmt.Sprintf("%v", addressMap["state"]))
					}

					if addressMap["postalCode"] != nil {
						result = append(result, fmt.Sprintf("%v", addressMap["postalCode"]))
					}

					if addressMap["country"] != nil {
						result = append(result, fmt.Sprintf("%v", addressMap["country"]))
					}

					finalString := strings.Join(result, ", ")

					if !containsOrgId(organizationAddresses[identifierKey], finalString) {
						organizationAddresses[identifierKey] = append(organizationAddresses[identifierKey], finalString)
					}

					postalCode, ok := addressMap["postalCode"].(string)
					if ok {
						organizationZip[identifierKey] = postalCode
					}
				}
			}

			if isIdentifierPresent && bundleEntry.Resource.Active != nil {
				organizationActive[identifierKey] = strconv.FormatBool(bundleEntry.Resource.Active.(bool))
			}

			if isIdentifierPresent {
				organizationName[identifierKey] = bundleEntry.Resource.Name
			}

			if isIdentifierPresent {
				organizationURL[identifierKey] = bundleEntry.FullURL
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
				for _, org := range bundleEntry.Resource.Orgs {
					if len(org.Address) > 0 {
						address := org.Address[0]
						if address.PostalCode != "" {
							organizationZip[org.Id] = strings.TrimSpace(address.PostalCode)
						}
					}
				}

				var endpointId string
				if len(endpointOrgMap[bundleEntry.Resource.OrgId]) > 0 {
					endpointId = bundleEntry.Resource.OrgId
				} else {
					endpointId = bundleEntry.FullURL
				}

				isPersisted := false

				for _, identifierKey := range endpointOrgMap[endpointId] {

					isPersisted = true

					entry.URL = strings.TrimSpace(entryURL)

					orgName, ok := organizationName[identifierKey]
					if ok {
						entry.OrganizationName = strings.TrimSpace(orgName)
					}

					orgURL, ok := organizationURL[identifierKey]
					if ok {
						entry.OrganizationURL = strings.TrimSpace(orgURL)
					}

					address, ok := organizationAddresses[identifierKey]
					if ok {
						entry.OrganizationAddresses = address
					}

					identifier, ok := organizationIdentifiers[identifierKey]
					if ok {
						entry.OrganizationIdentifiers = identifier
					}

					npiID, ok := organizationNPI[identifierKey]
					if ok {
						entry.NPIID = strings.TrimSpace(npiID)
					}

					active, ok := organizationActive[identifierKey]
					if ok {
						entry.OrganizationActive = active
					}

					postalCode, ok := organizationZip[identifierKey]
					if ok {
						entry.OrganizationZipCode = strings.TrimSpace(postalCode)
					}

					lanternEntryList = append(lanternEntryList, entry)
				}

				// Append only the endpoint URL if the organization data is not parsed
				if !isPersisted {
					entry.URL = strings.TrimSpace(entryURL)
					lanternEntryList = append(lanternEntryList, entry)
				}
			}
		}
	}

	return lanternEntryList
}
