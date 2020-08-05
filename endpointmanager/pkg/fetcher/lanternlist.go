package fetcher

// LanternList implements the Endpoints interface for lantern endpoint lists
type LanternList struct{}

// GetEndpoints takes the list of lantern endpoints and formats it into a ListOfEndpoints
func (ll LanternList) GetEndpoints(lanternList []map[string]interface{}) ListOfEndpoints {
	var finalList ListOfEndpoints
	var innerList []EndpointEntry

	for entry := range lanternList {
		fhirEntry := EndpointEntry{
			ListSource: string(Lantern),
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
