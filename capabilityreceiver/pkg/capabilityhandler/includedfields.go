package capabilityhandler

import "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"

var arrayFields = []string{"resource", "interaction", "searchParam", "operation", "document"}

func RunIncludedFieldsAndExtensionsChecks(capInt map[string]interface{}) []endpointmanager.IncludedField {
	if capInt == nil {
		return nil
	}
	var includedFields []endpointmanager.IncludedField
	includedFields = RunIncludedFieldsChecks(capInt, includedFields)
	includedFields = RunIncludedExtensionsChecks(capInt, includedFields)
	return includedFields
}

// RunIncludedFieldsChecks stores whether each field in capability statement is populated or not populated
func RunIncludedFieldsChecks(capInt map[string]interface{}, includedFields []endpointmanager.IncludedField) []endpointmanager.IncludedField {
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
			Field:     stringIndex,
			Exists:    checkField(capInt, fieldNames),
			Extension: false,
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

func RunIncludedExtensionsChecks(capInt map[string]interface{}, includedFields []endpointmanager.IncludedField) []endpointmanager.IncludedField {
	extensionList := [][]string{
		{"rest", "security", "extension", "http://fhir-registry.smarthealthit.org/StructureDefinition/capabilities"},
		{"rest", "resource", "extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-search-parameter-combination"},
		{"extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-supported-system"},
		{"rest", "extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-websocket"},
		{"rest", "security", "extension", "http://fhir-registry.smarthealthit.org/StructureDefinition/oauth-uris"},
		{"extension", "http://hl7.org/fhir/StructureDefinition/replaces"},
		{"extension", "http://hl7.org/fhir/StructureDefinition/resource-approvalDate"},
		{"extension", "http://hl7.org/fhir/StructureDefinition/resource-effectivePeriod"},
		{"extension", "http://hl7.org/fhir/StructureDefinition/resource-lastReviewDate"},
	}

	multipleFieldsExtensionList := [][]string{
		{"expectation", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-expectation"},
		{"prohibited", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-prohibited"},
	}

	for _, extensionPath := range extensionList {
		extensionURL := extensionPath[len(extensionPath)-1]
		extensionObj := endpointmanager.IncludedField{
			Field:     extensionURL,
			Exists:    checkExtension(capInt, extensionPath, extensionURL),
			Extension: true,
		}
		includedFields = append(includedFields, extensionObj)
	}

	for _, multipleExtensionPath := range multipleFieldsExtensionList {
		extensionName := multipleExtensionPath[0]
		extensionURL := multipleExtensionPath[1]
		extensionObj := endpointmanager.IncludedField{
			Field:     extensionURL,
			Exists:    checkMultipleFieldsExtension(capInt, extensionURL, extensionName),
			Extension: true,
		}
		includedFields = append(includedFields, extensionObj)
	}

	return includedFields
}

func checkExtension(capInt map[string]interface{}, fieldNames []string, url string) bool {
	for index, name := range fieldNames {
		if capInt[name] == nil {
			return false
		}

		field := capInt[name]

		if name == "extension" {
			extensionArr := field.([]interface{})
			return checkExtensionURL(extensionArr, url)
		} else if arrContains(arrayFields, name) {
			fieldArr := field.([]interface{})
			return checkArrFieldExtension(fieldNames[index+1:len(fieldNames)], fieldArr, url, false)
		} else if name == "rest" {
			restArr := field.([]interface{})
			capInt = restArr[0].(map[string]interface{})
		} else {
			capInt = field.(map[string]interface{})
		}
	}

	return false
}

func checkArrFieldExtension(fieldNames []string, fieldArr []interface{}, url string, found bool) bool {
	for _, resource := range fieldArr {
		name := fieldNames[0]
		resourceMap := resource.(map[string]interface{})
		extensionField := resourceMap[name]
		if extensionField == nil {
			continue
		} else if name != "extension" {
			extensionArr := extensionField.([]interface{})
			found = checkArrFieldExtension(fieldNames[1:len(fieldNames)], extensionArr, url, found)
			if found {
				return found
			}
		} else {
			extensionArr := extensionField.([]interface{})
			found = checkExtensionURL(extensionArr, url)
			if found {
				return found
			}
		}
	}
	return found
}

func checkExtensionURL(extensionArr []interface{}, url string) bool {
	found := false
	for _, extension := range extensionArr {
		extensionMap := extension.(map[string]interface{})
		urlField := extensionMap["url"]
		if urlField == url {
			found = true
			break
		}
	}
	return found
}

func checkMultipleFieldsExtension(capInt map[string]interface{}, url string, extensionString string) bool {
	extensionList := [][]string{
		{"rest", "resource", "interaction", "extension"},
		{"rest", "resource", "searchParam", "extension"},
		{"rest", "searchParam", "extension"},
		{"rest", "operation", "extension"},
		{"document", "extension"},
		{"rest", "interaction", "extension"},
	}
	/*if extensionString == "expectation" {
		row1 := []string{"resource", "searchInclude", "extension"}
		row2 := []string{"rest", "resource", "searchRevInclude", "extension"}
		extensionList = append(extensionList, row1, row2)
	}*/

	found := false
	for _, extensionPath := range extensionList {
		found = checkExtension(capInt, extensionPath, url)
		if found {
			break
		}
	}

	return found
}

func arrContains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
