package capabilityhandler

import "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"

// RunIncludedFieldsChecks stores whether each field in capability statement is populated or not populated
func RunIncludedFieldsChecks(capInt map[string]interface{}) []endpointmanager.IncludedField {
	if capInt == nil {
		return nil
	}
	var includedFields []endpointmanager.IncludedField

	fieldsList := [][]string{
		{"url"},
		{"version"},
		{"name"},
		{"title"},
		{"status"},
		{"experimental"},
		{"date"},
		{"publisher"},
		{"contact"},
		{"description"},
		{"requirements"},
		{"useContext"},
		{"jurisdiction"},
		{"purpose"},
		{"copyright"},
		{"kind"},
		{"instantiates"},
		{"imports"},
		{"software", "name"},
		{"software", "version"},
		{"software", "releaseDate"},
		{"implementation", "description"},
		{"implementation", "url"},
		{"implementation", "custodian"},
		{"fhirVersion"},
		{"format"},
		{"patchFormat"},
		{"acceptUnknown"},
		{"implementationGuide"},
		{"profile"},
		{"messaging"},
		{"document"},
	}

	for _, fieldNames := range fieldsList {
		var stringIndex string
		if len(fieldNames) != 1 {
			for index, name := range fieldNames {
				if index == (len(fieldNames) - 1) {
					stringIndex = stringIndex + name
				} else if index == 0 {
					stringIndex = name + "."
				} else {
					stringIndex = stringIndex + "." + name
				}
			}
		} else {
			stringIndex = fieldNames[0]
		}
		fieldObj := endpointmanager.IncludedField{
			Field:  stringIndex,
			Exists: checkField(capInt, fieldNames),
			Extension: false
		}
		includedFields = append(includedFields, fieldObj)
	}

	return includedFields
}

// Checks whether the given field is populated in the capability statement
func checkField(capInt map[string]interface{}, fieldNames []string) bool {
	for index, name := range fieldNames {
		if capInt[name] == nil {
			return false
		}

		field := capInt[name]

		if index == (len(fieldNames) - 1) {
			return field != nil
		}

		capInt = field.(map[string]interface{})

	}

	return false
}

func RunIncludedExtensionsChecks(capInt map[string]interface{}) []endpointmanager.IncludedField {
	if capInt == nil {
		return nil
	}
	var includedFields []endpointmanager.IncludedField

	fieldsList := [][]string{
		{"rest", "security", "http://fhir-registry.smarthealthit.org/StructureDefinition/capabilities"},
		{"rest", "resource", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-search-parameter-combination"},
		
	}

	for _, fieldNames := range fieldsList {
		var stringIndex string
		if len(fieldNames) != 1 {
			for index, name := range fieldNames {
				if index == (len(fieldNames) - 1) {
					stringIndex = stringIndex + name
				} else if index == 0 {
					stringIndex = name + "."
				} else {
					stringIndex = stringIndex + "." + name
				}
			}
		} else {
			stringIndex = fieldNames[0]
		}
		fieldObj := endpointmanager.IncludedField{
			Field:  stringIndex,
			Exists: checkField(capInt, fieldNames),
			Extension: false
		}
		includedFields = append(includedFields, fieldObj)
	}

	return includedFields
}	
