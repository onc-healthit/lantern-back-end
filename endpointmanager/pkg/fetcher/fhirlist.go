package fetcher

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

// FHIRList implements the Endpoints interface for an endpoint list in FHIR
type FHIRList struct{}

// GetEndpoints takes the list of endpoints in FHIR Bundle format and formats it into a ListOfEndpoints
// managingOrganization is set to true if the endpoint list contains an organization name in the managingOrganization field, otherwise false
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
			fhirEntry.ListSource = "FHIR"
		}

		resource, ok := fhirList[entry]["resource"].(map[string]interface{})
		if ok {
			uri, uriOk := resource["address"].(string)
			if uriOk {
				fhirEntry.FHIRPatientFacingURI = uri

				nameEndpt, nameOk := resource["name"].(string)
				if nameOk {
					fhirEntry.OrganizationNames = append(fhirEntry.OrganizationNames, nameEndpt)
				}

				// Save both name & managing organization in the array since both could be used
				// for storing the organization name if managingOrganization boolean is true
				managingOrg, orgOk := resource["managingOrganization"].(map[string]interface{})
				if orgOk {
					orgName, orgOk := managingOrg["display"].(string)
					if orgOk {
						if !helpers.StringArrayContains(fhirEntry.OrganizationNames, orgName) {
							fhirEntry.OrganizationNames = append(fhirEntry.OrganizationNames, orgName)
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
											if !helpers.StringArrayContains(fhirEntry.OrganizationNames, entryName) {
												fhirEntry.OrganizationNames = append(fhirEntry.OrganizationNames, entryName)
											}
										}
									}
								}
							}
						}
					}
				}
				if fhirEntry.OrganizationNames == nil {
					log.Warnf("No associated organization name for the URL %s.", uri)
				}
				innerList = append(innerList, fhirEntry)
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
