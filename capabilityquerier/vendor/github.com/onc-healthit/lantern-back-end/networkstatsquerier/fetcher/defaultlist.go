package fetcher

// GetEndpoints takes the a list of endpoints and formats it into a ListOfEndpoints
func getDefaultEndpoints(defaultList []map[string]interface{}, source string) ListOfEndpoints {
	var finalList ListOfEndpoints
	var innerList []EndpointEntry
	for entry := range defaultList {
		fhirEntry := EndpointEntry{
			ListSource: source,
		}
		orgName, orgOk := defaultList[entry]["OrganizationName"].(string)
		if orgOk {
			fhirEntry.OrganizationName = orgName
		}
		uri, uriOk := defaultList[entry]["FHIRPatientFacingURI"].(string)
		if uriOk {
			fhirEntry.FHIRPatientFacingURI = uri
		}
		innerList = append(innerList, fhirEntry)
	}

	finalList.Entries = innerList
	return finalList
}
