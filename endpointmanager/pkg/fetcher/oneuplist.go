package fetcher

import "strings"

// OneUpList implements the Endpoints interface for cerner endpoint lists
type OneUpList struct{}

// GetEndpoints takes the list of 1Up endpoints and formats it into a ListOfEndpoints
func (ul OneUpList) GetEndpoints(oneupList []map[string]interface{}, listURL string) ListOfEndpoints {
	var finalList ListOfEndpoints
	var innerList []EndpointEntry

	for entry := range oneupList {
		fhirEntry := EndpointEntry{}
		if listURL != "" {
			fhirEntry.ListSource = listURL
		} else {
			fhirEntry.ListSource = "1Up"
		}
		orgName, orgOk := oneupList[entry]["name"].(string)
		if orgOk {
			fhirEntry.OrganizationNames = []string{orgName}
		}
		uri, uriOk := oneupList[entry]["resource_url"].(string)
		if uriOk {
			fhirEntry.FHIRPatientFacingURI = uri
		}

		if !strings.Contains(uri, "https://api.") && !strings.Contains(uri, "https://fhir.healow.com/FHIRServer/fhir/") {
			innerList = append(innerList, fhirEntry)
		}
	}

	finalList.Entries = innerList
	return finalList
}
