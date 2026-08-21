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
	Telecom      interface{}          `json:"telecom"`
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
	Id           string      `json:"id"`
	Name         string      `json:"name"`
	Address      interface{} `json:"address"`
	Identifier   interface{} `json:"identifier"`
	Active       interface{} `json:"active"`
	Telecom      interface{} `json:"telecom"`
	ResourceType string      `json:"resourceType"`
}

// orgMaps groups the per-organization-keyCount maps populated by
// extractOrganizationFields, so callers only need to pass one value.
type orgMaps struct {
	zip         map[int]string
	name        map[int]string
	url         map[int]string
	addresses   map[int][]string
	identifiers map[int][]string
	active      map[int]string
	npi         map[int]string
}

// extractOrganizationFields pulls name/identifier/address/active/telecom(url)
// data out of a FHIR Organization resource (top-level or contained) into the
// maps in m, keyed by keyCount. identifier, address, and active are raw
// json.Unmarshal results (interface{}) since FHIR allows varying shapes.
func extractOrganizationFields(m orgMaps, keyCount int, name string, identifier, address, active, telecom interface{}) {
	if identifier != nil {
		identifierArr, ok := identifier.([]interface{})
		if ok {
			for _, id := range identifierArr {
				identifierMap, ok := id.(map[string]interface{})
				if !ok {
					continue
				}
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

						if !containsOrgId(m.identifiers[keyCount], identifierStr) {
							m.identifiers[keyCount] = append(m.identifiers[keyCount], identifierStr)
						}

						if identifierCode == "NPI" {
							m.npi[keyCount] = identifierMap["value"].(string)
						}
					}
				}
			}
		}
	}

	if address != nil {
		addressMapArr, ok := address.([]interface{})
		if ok {
			for _, addr := range addressMapArr {
				addressMap, ok := addr.(map[string]interface{})
				if !ok {
					continue
				}

				// Get the values inside "line" array of the address
				var result []string
				if addressMap["line"] != nil {
					lineMap, ok := addressMap["line"].([]interface{})
					if ok {
						for _, line := range lineMap {
							if line != nil {
								result = append(result, fmt.Sprintf("%v", line))
							}
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

				if !containsOrgId(m.addresses[keyCount], finalString) {
					m.addresses[keyCount] = append(m.addresses[keyCount], finalString)
				}

				postalCode, ok := addressMap["postalCode"].(string)
				if ok {
					m.zip[keyCount] = postalCode
				}
			}
		}
	}

	if active != nil {
		activeBool, ok := active.(bool)
		if ok {
			m.active[keyCount] = strconv.FormatBool(activeBool)
		}
	}

	m.name[keyCount] = name

	// Extract organization URL from telecom field when system is "url"
	if telecom != nil {
		telecomArr, ok := telecom.([]interface{})
		if ok {
			for _, tc := range telecomArr {
				telecomMap, ok := tc.(map[string]interface{})
				if !ok {
					continue
				}
				if telecomMap["system"] != nil && telecomMap["system"].(string) == "url" {
					if telecomMap["value"] != nil && telecomMap["value"].(string) != "" {
						m.url[keyCount] = strings.TrimSpace(telecomMap["value"].(string))
						break // Use the first URL found
					}
				}
			}
		}
	}
}

func containsOrgId(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func containsKey(s []int, i int) bool {
	for _, v := range s {
		if v == i {
			return true
		}
	}
	return false
}

func BundleToLanternFormat(bundle []byte, chplURL string) ([]LanternEntry, []string, int, bool) {
	var lanternEntryList []LanternEntry

	var endpointOrgMap = make(map[string][]int)
	var organizationZip = make(map[int]string)
	var organizationName = make(map[int]string)
	var organizationURL = make(map[int]string)
	var organizationAddresses = make(map[int][]string)
	var organizationIdentifiers = make(map[int][]string)
	var organizationActive = make(map[int]string)
	var organizationNPI = make(map[int]string)

	keyCount := 0

	var structBundle FHIRBundle

	// --- Logging for empty or invalid bundle input ---
	trimmed := strings.TrimSpace(string(bundle))
	preview := trimmed
	if len(preview) > 50 {
		preview = preview[:30]
	}
	// Minimal cleanup: only remove newlines so the log stays single-line
	preview = strings.ReplaceAll(preview, "\n", " ")
	preview = strings.ReplaceAll(preview, "\r", " ")

	// Case 1: empty response body
	if len(trimmed) == 0 {
		log.Warnf("Empty FHIR bundle detected. \n URL=%s \n Size=0 bytes", chplURL)

		// Case 2: HTML/XML (starts with '<')
	} else if strings.HasPrefix(trimmed, "<") {
		log.Warnf("Non-JSON response received. \n URL=%s. \n FirstBytes=\"%s\"", chplURL, preview)

		// Case 3: Not JSON object or array
	} else if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		log.Warnf("Unexpected response format. \n URL=%s. Not JSON. \n FirstBytes=\"%s\"", chplURL, preview)
	}

	// --- Attempt JSON unmarshal ---
	err := json.Unmarshal(bundle, &structBundle)
	if err != nil {
		// Additional logging for empty or malformed JSON
		log.Errorf("Failed to unmarshal FHIR bundle. \n URL=%s \n Size=%d bytes \n Error=%v",
			chplURL, len(bundle), err)

		log.Warn("Handler is required for url ", chplURL)
		log.Fatal("More details about the error: ", err)
	}

	// --- Bundle parsed but contains no entries ---
	if len(structBundle.Entries) == 0 {
		log.Warnf("Parsed FHIR bundle contains 0 entries. URL=%s Size=%d bytes",
			chplURL, len(bundle),
		)
	}

	usedOrgKeys := make(map[int]bool)

	maps := orgMaps{
		zip:         organizationZip,
		name:        organizationName,
		url:         organizationURL,
		addresses:   organizationAddresses,
		identifiers: organizationIdentifiers,
		active:      organizationActive,
		npi:         organizationNPI,
	}

	for _, bundleEntry := range structBundle.Entries {
		if strings.EqualFold(strings.TrimSpace(bundleEntry.Resource.ResourceType), "Organization") {
			extractOrganizationFields(maps, keyCount, bundleEntry.Resource.Name,
				bundleEntry.Resource.Identifier, bundleEntry.Resource.Address,
				bundleEntry.Resource.Active, bundleEntry.Resource.Telecom)

			if bundleEntry.Resource.Endpoint != nil {
				endpointArr := bundleEntry.Resource.Endpoint.([]interface{})
				for _, endpoint := range endpointArr {
					endpointMap := endpoint.(map[string]interface{})
					if endpointMap["reference"] != nil && endpointMap["reference"].(string) != "" {
						endpointId := endpointMap["reference"].(string)
						endpointId = strings.TrimPrefix(endpointId, "Endpoint/")
						endpointId = strings.TrimPrefix(endpointId, "endpoint/")

						// Store endpoint-to-organizations mapping (if not already present)
						if !containsKey(endpointOrgMap[endpointId], keyCount) {
							endpointOrgMap[endpointId] = append(endpointOrgMap[endpointId], keyCount)
						}
					}
				}
			}

			keyCount++
		} else if strings.EqualFold(strings.TrimSpace(bundleEntry.Resource.ResourceType), "Endpoint") && len(bundleEntry.Resource.Orgs) > 0 {
			containedOrg := bundleEntry.Resource.Orgs[0]

			extractOrganizationFields(maps, keyCount, containedOrg.Name,
				containedOrg.Identifier, containedOrg.Address,
				containedOrg.Active, containedOrg.Telecom)

			endpointId := strings.TrimSpace(bundleEntry.Resource.OrgId)
			if endpointId != "" && !containsKey(endpointOrgMap[endpointId], keyCount) {
				endpointOrgMap[endpointId] = append(endpointOrgMap[endpointId], keyCount)
			}

			keyCount++
		}
	}

	totalOrgCount := keyCount

	// --- No Organization resources found in the bundle at all ---
	// Only flag this when the bundle actually has entries; an empty bundle
	// is already reported separately (see "0 entries" warning above).
	noOrganizationsFound := totalOrgCount == 0 && len(structBundle.Entries) > 0
	if noOrganizationsFound {
		log.Warnf("Parsed FHIR bundle contains no Organization resources. URL=%s Size=%d bytes",
			chplURL, len(bundle),
		)
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
				var endpointId string
				if len(endpointOrgMap[bundleEntry.Resource.OrgId]) > 0 {
					endpointId = bundleEntry.Resource.OrgId
				} else {
					endpointId = bundleEntry.FullURL
				}

				isPersisted := false
				hadAnyMapped := len(endpointOrgMap[endpointId]) > 0

				for _, keyCount := range endpointOrgMap[endpointId] {
					var entry LanternEntry
					isPersisted = true
					usedOrgKeys[keyCount] = true

					entry.URL = strings.TrimSpace(entryURL)

					active, ok := organizationActive[keyCount]
					if ok {
						entry.OrganizationActive = active
					}

					// If active present and false -> skip this org
					if ok && active == "false" {
						lanternEntryList = append(lanternEntryList, entry)
						continue
					}

					orgName, ok := organizationName[keyCount]
					if ok {
						entry.OrganizationName = strings.TrimSpace(orgName)
					}

					orgURL, ok := organizationURL[keyCount]
					if ok {
						entry.OrganizationURL = strings.TrimSpace(orgURL)
					}

					address, ok := organizationAddresses[keyCount]
					if ok {
						entry.OrganizationAddresses = address
					}

					identifier, ok := organizationIdentifiers[keyCount]
					if ok {
						entry.OrganizationIdentifiers = identifier
					}

					npiID, ok := organizationNPI[keyCount]
					if ok {
						entry.NPIID = strings.TrimSpace(npiID)
					}

					postalCode, ok := organizationZip[keyCount]
					if ok {
						entry.OrganizationZipCode = strings.TrimSpace(postalCode)
					}

					lanternEntryList = append(lanternEntryList, entry)
				}

				// Append only the endpoint URL if the organization data is not parsed
				if !isPersisted {
					if !hadAnyMapped {
						entry.URL = strings.TrimSpace(entryURL)
						lanternEntryList = append(lanternEntryList, entry)
					}
				}
			}
		}
	}

	var unpopulatedOrgs []string
	inactiveOrgCount := 0
	for i := 0; i < totalOrgCount; i++ {
		if organizationActive[i] == "false" {
			inactiveOrgCount++
		}

		if usedOrgKeys[i] {
			continue
		}
		name := strings.TrimSpace(organizationName[i])
		if name == "" {
			name = fmt.Sprintf("organization #%d", i)
		}
		unpopulatedOrgs = append(unpopulatedOrgs, name)
	}

	return lanternEntryList, unpopulatedOrgs, inactiveOrgCount, noOrganizationsFound
}
