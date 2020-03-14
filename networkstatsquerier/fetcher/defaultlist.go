package fetcher

// DefaultList implements the Endpoints interface for endpoint lists in the default format
// which is the format that matches the ListOfEndpoints struct
type DefaultList struct{}

// GetEndpoints takes the list of cerner endpoints and formats it into a ListOfEndpoints
func (dl DefaultList) GetEndpoints(defaultList []map[string]interface{}) (ListOfEndpoints, error) {
	var finalList ListOfEndpoints
	var innerList []EndpointEntry

	for entry := range defaultList {
		fhirEntry := EndpointEntry{
			ListSource: "https://open.epic.com/MyApps/EndpointsJson",
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
	return finalList, nil
}
