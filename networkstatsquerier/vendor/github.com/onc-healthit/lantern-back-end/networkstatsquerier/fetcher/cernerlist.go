package fetcher

import (
	"encoding/json"
	"fmt"
)

// CernerList implements the Endpoints interface for cerner endpoint lists
type CernerList struct{}

// GetEndpoints takes the list of cerner endpoints and formats it into a ListOfEndpoints
func (cl CernerList) GetEndpoints(cernerList interface{}) (ListOfEndpoints, error) {
	var finalList ListOfEndpoints
	var innerList []EndpointEntry
	var formatList []map[string]interface{}

	// For reformatting purposes
	bc, err := json.Marshal(cernerList)
	if err != nil {
		return finalList, fmt.Errorf("unable to format cerner list")
	}
	err = json.Unmarshal(bc, &formatList)
	if err != nil {
		return finalList, fmt.Errorf("unable to format cerner list")
	}

	for entry := range formatList {
		orgName, orgOk := formatList[entry]["name"].(string)
		if !orgOk {
			orgName = ""
		}
		uri, uriOk := formatList[entry]["baseUrl"].(string)
		if !uriOk {
			uri = ""
		}
		fhirEntry := EndpointEntry{
			OrganizationName:     orgName,
			FHIRPatientFacingURI: uri,
			ListSource:           "Cerner",
		}
		innerList = append(innerList, fhirEntry)
	}

	finalList.Entries = innerList
	return finalList, nil
}
