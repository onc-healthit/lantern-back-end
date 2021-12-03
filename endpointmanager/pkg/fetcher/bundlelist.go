package fetcher

import (
	log "github.com/sirupsen/logrus"
)

// GetBundleEndpoints takes the list of endpoints in FHIR Bundle format and formats it into a ListOfEndpoints
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
func GetBundleEndpoints(bundleList []map[string]interface{}, source string, listURL string, managingOrganization bool) ListOfEndpoints {
	var finalList ListOfEndpoints
	var innerList []EndpointEntry

	for entry := range bundleList {
		fhirEntry := EndpointEntry{}
		if listURL != "" {
			fhirEntry.ListSource = listURL
		} else {
			fhirEntry.ListSource = source
		}

		resource, ok := bundleList[entry]["resource"].(map[string]interface{})
		if ok {
			uri, uriOk := resource["address"].(string)
			if uriOk {
				fhirEntry.FHIRPatientFacingURI = uri

				// Save both name & managing organization in the array since both could be used
				// for storing the organization name if managingOrganization boolean is true
				if managingOrganization {
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
				}
				nameEndpt, nameOk := resource["name"].(string)
				if nameOk {
					fhirEntry.OrganizationNames = append(fhirEntry.OrganizationNames, nameEndpt)
				}

				if fhirEntry.OrganizationNames == nil {
					log.Warnf("No associated organization name for the URL %s.", uri)
				}
				innerList = append(innerList, fhirEntry)
			} else {
				log.Warnf("No address field in the resource. Ignoring resource.")
			}
		} else {
			log.Warnf("No resource field in " + source + " list. Returning an empty list of entries.")
		}
	}

	finalList.Entries = innerList
	return finalList
}
