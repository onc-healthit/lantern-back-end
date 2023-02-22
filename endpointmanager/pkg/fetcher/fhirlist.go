package fetcher

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

// FHIRList implements the Endpoints interface for an endpoint list in FHIR
type FHIRList struct{}

// GetEndpoints takes the list of endpoints in FHIR Bundle format and formats it into a ListOfEndpoints
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
func (fl FHIRList) GetEndpoints(fhirList []map[string]interface{}, source string, listURL string) ListOfEndpoints {
	var finalList ListOfEndpoints
	var innerList []EndpointEntry

	for entry := range fhirList {
		var organizationNames []string

		resource, ok := fhirList[entry]["resource"].(map[string]interface{})
		if ok {
			uri, uriOk := resource["address"].(string)
			if uriOk {

				nameEndpt, nameOk := resource["name"].(string)
				if nameOk {
					organizationNames = append(organizationNames, nameEndpt)
				}

				// Save both name & managing organization in the array since both could be used
				// for storing the organization name if managingOrganization boolean is true
				managingOrg, orgOk := resource["managingOrganization"].(map[string]interface{})
				if orgOk {
					orgName, orgOk := managingOrg["display"].(string)
					if orgOk {
						if !helpers.StringArrayContains(organizationNames, orgName) {
							organizationNames = append(organizationNames, orgName)
						}
					}
					orgReference, orgOk := managingOrg["reference"].(string)
					if orgOk {
						containedList, orgOk := resource["contained"].([]interface{})
						if orgOk {
							for index := range containedList {
								entry := containedList[index].(map[string]interface{})
								entryType, orgOk := entry["resourceType"].(string)
								if orgOk && entryType == "Organization" {
									entryID, orgOk := entry["id"].(string)
									if orgOk && entryID == orgReference {
										entryName, orgOk := resource["name"].(string)
										if orgOk {
											if !helpers.StringArrayContains(organizationNames, entryName) {
												organizationNames = append(organizationNames, entryName)
											}
										}
									}
								}
							}
						}
					}
				}
				if len(organizationNames) == 0 {
					log.Warnf("No associated organization name for the URL %s.", uri)
				} else {
					for _, orgName := range organizationNames {
						fhirEntry := EndpointEntry{}

						if listURL != "" {
							fhirEntry.ListSource = listURL
						} else if source != "" {
							fhirEntry.ListSource = source
						} else {
							fhirEntry.ListSource = "FHIR"
						}

						fhirEntry.FHIRPatientFacingURI = uri
						fhirEntry.OrganizationName = orgName

						innerList = append(innerList, fhirEntry)
					}
				}
			} else {
				log.Warnf("No address field in the resource. Ignoring resource.")
			}
		} else {
			log.Warnf("No resource field in FHIR list. Returning an empty list of entries.")
		}
	}

	finalList.Entries = innerList
	return finalList
}
