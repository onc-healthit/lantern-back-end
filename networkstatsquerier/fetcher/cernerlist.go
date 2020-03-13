package fetcher

// CernerList implements the Endpoints interface for cerner endpoint lists
type CernerList struct{}

// GetEndpoints takes the list of cerner endpoints and formats it into a ListOfEndpoints
func (cl CernerList) GetEndpoints(cernerList []map[string]interface{}) (ListOfEndpoints, error) {
	var finalList ListOfEndpoints
	var innerList []EndpointEntry

	for entry := range cernerList {
		orgName, orgOk := cernerList[entry]["name"].(string)
		if !orgOk {
			orgName = ""
		}
		uri, uriOk := cernerList[entry]["baseUrl"].(string)
		if !uriOk {
			uri = ""
		}
		fhirEntry := EndpointEntry{
			OrganizationName:     orgName,
			FHIRPatientFacingURI: uri,
			ListSource:           "https://github.com/cerner/ignite-endpoints",
		}
		innerList = append(innerList, fhirEntry)
	}

	finalList.Entries = innerList
	return finalList, nil
}
