package fetcher

import (
	"encoding/json"
	"fmt"
)

// DefaultList implements the Endpoints interface for endpoint lists in the default format
// which is the format that matches the ListOfEndpoints struct
type DefaultList struct{}

// GetEndpoints takes the list of cerner endpoints and formats it into a ListOfEndpoints
func (dl DefaultList) GetEndpoints(defaultList interface{}) (ListOfEndpoints, error) {
	var finalList ListOfEndpoints
	var innerList []EndpointEntry
	var formatList []map[string]interface{}

	// For reformatting purposes
	bc, err := json.Marshal(defaultList)
	if err != nil {
		return finalList, fmt.Errorf("unable to marshal default list")
	}
	err = json.Unmarshal(bc, &formatList)
	if err != nil {
		return finalList, fmt.Errorf("unable to unmarshal default list")
	}

	for entry := range formatList {
		fhirEntry := EndpointEntry{
			ListSource: "Epic",
		}
		orgName, orgOk := formatList[entry]["OrganizationName"].(string)
		if orgOk {
			fhirEntry.OrganizationName = orgName
		}
		uri, uriOk := formatList[entry]["FHIRPatientFacingURI"].(string)
		if uriOk {
			fhirEntry.FHIRPatientFacingURI = uri
		}
		entryType, typeOk := formatList[entry]["Type"].(string)
		if typeOk {
			fhirEntry.Type = entryType
			// If the entry has a type field then it's a CareEvolution list
			fhirEntry.ListSource = "CareEvolution"
		}
		keywords, keyOk := formatList[entry]["Keywords"].([]OrgKeyword)
		if keyOk {
			fhirEntry.Keywords = keywords
		}
		innerList = append(innerList, fhirEntry)
	}

	finalList.Entries = innerList
	return finalList, nil
}
