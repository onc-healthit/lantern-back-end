package capabilityhandler

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
)

// RunSupportedFieldsCheck returns an interface that contains information about what profiles the server supports
func RunSupportedFieldsCheck(capInt map[string]interface{}, fhirVersion string) []endpointmanager.SupportedProfile {
	if capInt == nil {
		return nil
	}

	var supportedProfiles []endpointmanager.SupportedProfile

	if helpers.StringArrayContains(dstu2, fhirVersion) {
		return getConformanceProfiles(capInt, supportedProfiles)
	} else if helpers.StringArrayContains(stu3, fhirVersion) || helpers.StringArrayContains(r4, fhirVersion) {
		return getCapabilityStatementProfiles(capInt, supportedProfiles)
	}

	return supportedProfiles
}

// getConformanceProfiles stores all the profiles found in the Conformance statement profile array
func getConformanceProfiles(capInt map[string]interface{}, supportedProfiles []endpointmanager.SupportedProfile) []endpointmanager.SupportedProfile {

	if capInt["profile"] == nil {
		return nil
	}

	supportedProfilesList := capInt["profile"].([]interface{})

	for _, profile := range supportedProfilesList {
		profileInt := profile.(map[string]interface{})
		var profileInfo endpointmanager.SupportedProfile

		if profileInt["display"] != nil {
			profileInfo.ProfileName = profileInt["display"].(string)
		}

		if profileInt["reference"] != nil {
			profileInfo.ProfileURL = profileInt["reference"].(string)
		}

		supportedProfiles = append(supportedProfiles, profileInfo)

	}

	return supportedProfiles
}

// getCapabilityStatementProfiles stores all the profiles found in the Capability statement rest->resource->profile field
func getCapabilityStatementProfiles(capInt map[string]interface{}, supportedProfiles []endpointmanager.SupportedProfile) []endpointmanager.SupportedProfile {

	if capInt["rest"] == nil {
		return nil
	}

	restArr := capInt["rest"].([]interface{})
	restInt := restArr[0].(map[string]interface{})
	if restInt["resource"] == nil {
		return nil
	}
	resourceArr := restInt["resource"].([]interface{})

	for _, resource := range resourceArr {
		resourceInt := resource.(map[string]interface{})

		resourceType := ""
		if resourceInt["type"] != nil {
			resourceType = resourceInt["type"].(string)
		}

		if resourceInt["supportedProfile"] != nil {
			supportedProfileArr := resourceInt["supportedProfile"].([]string)
			for _, profileURL := range supportedProfileArr {
				var profileInfo endpointmanager.SupportedProfile
				profileInfo.ProfileURL = profileURL
				profileInfo.Resource = resourceType

				supportedProfiles = append(supportedProfiles, profileInfo)
			}

		}
	}

	return supportedProfiles
}
