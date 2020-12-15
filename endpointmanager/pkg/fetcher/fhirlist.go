package fetcher

// FHIRList implements the Endpoints interface for epic's endpoint list in FHIR
type FHIRList struct{}

// GetEndpoints takes the list of cerner endpoints and formats it into a ListOfEndpoints
// Assumed Structure:
/**
{ ... entry: [ {
		fullUrl: URI for resource
		resource: {
			...
			name: <name of the endpoint>
			managingOrganiation: { display: <text for resource>, reference: <organization name> },
			address: <FHIR url>,
		}
	  }, ...
] }
*/
func (fl FHIRList) GetEndpoints(fhirList []map[string]interface{}, listURL string) ListOfEndpoints {
	var finalList ListOfEndpoints
	var innerList []EndpointEntry

	for entry := range fhirList {
		fhirEntry := EndpointEntry{}
		if listURL != "" {
			fhirEntry.ListSource = listURL
		} else {
			fhirEntry.ListSource = string(FHIR)
		}

		resource, ok := fhirList[entry]["resource"].(map[string]interface{})
		if ok {
			// Save both name & managing organization in the array since both could be used
			// for storing the organization name
			managingOrg, orgOk := resource["managingOrganization"].(map[string]interface{})
			if orgOk {
				orgName, orgOk := managingOrg["display"].(string)
				if orgOk {
					fhirEntry.OrganizationNames = append(fhirEntry.OrganizationNames, orgName)
				}
				alternateName, orgOk := managingOrg["reference"].(string)
				if orgOk {
					fhirEntry.OrganizationNames = append(fhirEntry.OrganizationNames, alternateName)
				}
			}
			nameEndpt, nameOk := resource["name"].(string)
			if nameOk {
				fhirEntry.OrganizationNames = append(fhirEntry.OrganizationNames, nameEndpt)
			}
			uri, uriOk := resource["address"].(string)
			if uriOk {
				fhirEntry.FHIRPatientFacingURI = uri
			}
			innerList = append(innerList, fhirEntry)
		}
	}

	finalList.Entries = innerList
	return finalList
}
