package fetcher

// LanternList implements the Endpoints interface for lantern endpoint lists
type LanternList struct{}

// GetEndpoints takes the list of lantern endpoints and formats it into a ListOfEndpoints
func (ll LanternList) GetEndpoints(lanternList []map[string]interface{}, source string, listURL string) ListOfEndpoints {
	var finalList ListOfEndpoints
	var innerList []EndpointEntry

	for entry := range lanternList {
		fhirEntry := EndpointEntry{}
		if listURL != "" {
			fhirEntry.ListSource = listURL
		} else if source != "" {
			fhirEntry.ListSource = source
		} else {
			fhirEntry.ListSource = "Lantern"
		}
		orgName, orgOk := lanternList[entry]["OrganizationName"].(string)
		if orgOk {
			fhirEntry.OrganizationNames = []string{orgName}
		}
		uri, uriOk := lanternList[entry]["URL"].(string)
		if uriOk {
			fhirEntry.FHIRPatientFacingURI = uri
		}
		npiID, npiIDOk := lanternList[entry]["NPIID"].(string)
		if npiIDOk {
			fhirEntry.NPIIDs = []string{npiID}
		}
		innerList = append(innerList, fhirEntry)
	}

	finalList.Entries = innerList
	return finalList
}
