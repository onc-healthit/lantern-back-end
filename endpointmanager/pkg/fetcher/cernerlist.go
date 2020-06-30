package fetcher

// CernerList implements the Endpoints interface for cerner endpoint lists
type CernerList struct{}

// GetEndpoints takes the list of cerner endpoints and formats it into a ListOfEndpoints
func (cl CernerList) GetEndpoints(cernerList []map[string]interface{}) ListOfEndpoints {
	var finalList ListOfEndpoints
	var innerList []EndpointEntry

	for entry := range cernerList {
		fhirEntry := EndpointEntry{
			ListSource: string(Cerner),
		}
		orgName, orgOk := cernerList[entry]["name"].(string)
		if orgOk {
			fhirEntry.OrganizationNames = []string{orgName}
		}
		uri, uriOk := cernerList[entry]["baseUrl"].(string)
		if uriOk {
			fhirEntry.FHIRPatientFacingURI = uri
		}
		innerList = append(innerList, fhirEntry)
	}

	finalList.Entries = innerList
	return finalList
}
