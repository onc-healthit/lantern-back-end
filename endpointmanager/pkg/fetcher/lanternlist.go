package fetcher

import log "github.com/sirupsen/logrus"

// LanternList implements the Endpoints interface for lantern endpoint lists
type LanternList struct{}

// GetEndpoints takes the list of lantern endpoints and formats it into a ListOfEndpoints
func (ll LanternList) GetEndpoints(lanternList []map[string]interface{}, source string, listURL string) ListOfEndpoints {
	var finalList ListOfEndpoints
	var innerList []EndpointEntry

	for entry := range lanternList {
		uri, uriOk := lanternList[entry]["URL"].(string)
		if uriOk {
			fhirEntry := EndpointEntry{}
			fhirEntry.FHIRPatientFacingURI = uri

			if listURL != "" {
				fhirEntry.ListSource = listURL
			} else if source != "" {
				fhirEntry.ListSource = source
			} else {
				fhirEntry.ListSource = "Lantern"
			}
			orgName, orgOk := lanternList[entry]["OrganizationName"].(string)
			if orgOk {
				fhirEntry.OrganizationName = orgName
			}
			npiID, npiIDOk := lanternList[entry]["NPIID"].(string)
			if npiIDOk {
				fhirEntry.NPIID = npiID
			}
			zipCode, zipCodeOk := lanternList[entry]["OrganizationZipCode"].(string)
			if zipCodeOk {
				if len(zipCode) > 5 {
					zipCode = zipCode[:5]
				}
				fhirEntry.OrganizationZipCode = zipCode
			}
			orgIdentifiers, orgIdOk := lanternList[entry]["OrganizationIdentifiers"]
			if orgIdOk && orgIdentifiers != nil {
				fhirEntry.OrganizationIdentifiers = orgIdentifiers.([]interface{})
			}
			orgAddresses, orgAddOk := lanternList[entry]["OrganizationAddresses"]
			if orgAddOk && orgAddresses != nil {
				fhirEntry.OrganizationAddresses = orgAddresses.([]interface{})
			}
			orgActive, orgActOk := lanternList[entry]["OrganizationActive"].(string)
			if orgActOk {
				fhirEntry.OrganizationActive = orgActive
			}
			innerList = append(innerList, fhirEntry)
		} else {
			log.Warnf("No URL field in Lantern list. Returning an empty list of entries.")
		}
	}

	finalList.Entries = innerList
	return finalList
}
