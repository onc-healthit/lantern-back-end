package capabilityhandler

import "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"

// List of capability statement fields that are arrays of interfaces
var arrayFields = []string{"rest", "resource", "interaction", "searchParam", "operation", "document", "_searchInclude", "_searchRevInclude"}

// RunIncludedFieldsAndExtensionsChecks returns an interface that contains information about whether fields and extensions are supported or not
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

	// Get name of field
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

		// Create fieldObj with field name, if the field exists, and if it is an extension
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

// RunIncludedExtensionsChecks stores whether each extension in capability statement is populated or not populated
func RunIncludedExtensionsChecks(capInt map[string]interface{}, includedFields []endpointmanager.IncludedField) []endpointmanager.IncludedField {
	extensionList := [][]string{
		{"extension", "http://hl7.org/fhir/StructureDefinition/conformance-supported-system", "conformance-supported-system"},
		{"rest", "resource", "extension", "http://hl7.org/fhir/StructureDefinition/conformance-search-parameter-combination", "conformance-search-parameter-combination"},
		{"rest", "security", "extension", "http://DSTU2/fhir-registry.smarthealthit.org/StructureDefinition/oauth-uris", "DSTU2-oauth-uris"},
		{"rest", "security", "extension", "http://fhir-registry.smarthealthit.org/StructureDefinition/capabilities", "capabilities"},
		{"rest", "resource", "extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-search-parameter-combination", "capabilitystatement-search-parameter-combination"},
		{"extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-supported-system", "capabilitystatement-supported-system"},
		{"rest", "extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-websocket", "capabilitystatement-websocket"},
		{"rest", "security", "extension", "http://fhir-registry.smarthealthit.org/StructureDefinition/oauth-uris", "oauth-uris"},
		{"extension", "http://hl7.org/fhir/StructureDefinition/replaces", "replaces"},
		{"extension", "http://hl7.org/fhir/StructureDefinition/resource-approvalDate", "resource-approvalDate"},
		{"extension", "http://hl7.org/fhir/StructureDefinition/resource-effectivePeriod", "resource-effectivePeriod"},
		{"extension", "http://hl7.org/fhir/StructureDefinition/resource-lastReviewDate", "resource-lastReviewDate"},
	}

	multipleFieldsExtensionList := [][]string{
		{"http://hl7.org/fhir/StructureDefinition/conformance-expectation", "conformance-expectation"},
		{"http://hl7.org/fhir/StructureDefinition/conformance-prohibited", "conformance-prohibited"},
		{"http://hl7.org/fhir/StructureDefinition/capabilitystatement-expectation", "capabilitystatement-expectation"},
		{"http://hl7.org/fhir/StructureDefinition/capabilitystatement-prohibited", "capabilitystatement-prohibited"},
	}

	// Get name of extension and create extensionObj with extension name, if the extension exists, and if it is an extension
	for _, extensionPath := range extensionList {
		extensionName := extensionPath[len(extensionPath)-1]
		extensionURL := extensionPath[len(extensionPath)-2]
		extensionObj := endpointmanager.IncludedField{
			Field:     extensionName,
			Exists:    checkExtension(capInt, extensionPath, extensionURL),
			Extension: true,
		}
		includedFields = append(includedFields, extensionObj)
	}

	// Get name of extension with multiple possible locations and create extensionObj with extension name, if the extension exists, and if it is an extension
	for _, multipleExtensionPath := range multipleFieldsExtensionList {
		extensionName := multipleExtensionPath[1]
		extensionURL := multipleExtensionPath[0]
		extensionObj := endpointmanager.IncludedField{
			Field:     extensionName,
			Exists:    checkMultipleFieldsExtension(capInt, extensionURL, extensionName),
			Extension: true,
		}
		includedFields = append(includedFields, extensionObj)
	}

	return includedFields
}

// Checks whether the extension is populated in the capability statement given a path of fieldNames
func checkExtension(capInt map[string]interface{}, fieldNames []string, url string) bool {
	for index, name := range fieldNames {
		if capInt[name] == nil {
			return false
		}

		field := capInt[name]

		if name == "extension" || name == "modifierExtension" {
			extensionArr := field.([]interface{})
			return checkExtensionURL(extensionArr, url)
		} else if arrContains(arrayFields, name) {
			fieldArr := field.([]interface{})
			nextIndex := index + 1
			length := len(fieldNames)
			return checkArrFieldExtension(fieldNames[nextIndex:length], fieldArr, url)
		} else {
			capInt = field.(map[string]interface{})
		}
	}

	return false
}

// Given an array of interface objects, loops through each object to check whether the extension is populated following the path of fieldNames
func checkArrFieldExtension(fieldNames []string, fieldArr []interface{}, url string) bool {
	var found bool
	// Loop through the array of interface objects
	for _, obj := range fieldArr {
		// For each object in interface array, get desired field using name in fieldNames
		name := fieldNames[0]
		objMap := obj.(map[string]interface{})
		extensionField := objMap[name]
		if extensionField == nil {
			// If the desired field does not exist in that object, continue to the next object within the array of interface objects
			continue
		} else if name != "extension" && name != "modifierExtension" && arrContains(arrayFields, name) {
			// If the desired field is not extension or modifierExtension and it is also an array of interface objects, call checkArrFieldExtension with this new array
			fieldArr := extensionField.([]interface{})
			length := len(fieldNames)
			found = checkArrFieldExtension(fieldNames[1:length], fieldArr, url)
			if found {
				return found
			}
		} else if name != "extension" && name != "modifierExtension" && !arrContains(arrayFields, name) {
			// If the desired field is not extension or modifierExtension and it is not an array of interface objects, call checkExtension with this field map[string]interface
			extensionField := extensionField.(map[string]interface{})
			length := len(fieldNames)
			found = checkExtension(extensionField, fieldNames[1:length], url)
			if found {
				return found
			}
		} else {
			// If the desired field is extension or modifierExtension, check array of extension interface objects for correct url
			extensionArr := extensionField.([]interface{})
			found = checkExtensionURL(extensionArr, url)
			if found {
				return found
			}
		}
	}
	return found
}

// Checks whether the given extension array contains the correct extension url
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

// checkMultipleFieldsExtension loops through all the possible locations of the extension to see if it exists
func checkMultipleFieldsExtension(capInt map[string]interface{}, url string, extensionString string) bool {
	var extensionList [][]string
	if extensionString == "capabilitystatement-prohibited" || extensionString == "conformance-prohibited" {
		extensionList = [][]string{
			{"rest", "resource", "interaction", "modifierExtension"},
			{"rest", "resource", "searchParam", "modifierExtension"},
			{"rest", "searchParam", "modifierExtension"},
			{"rest", "operation", "modifierExtension"},
			{"document", "modifierExtension"},
			{"rest", "interaction", "modifierExtension"},
		}
	}

	if extensionString == "capabilitystatement-expectation" || extensionString == "conformance-expectation" {
		extensionList = [][]string{
			{"rest", "resource", "interaction", "extension"},
			{"rest", "resource", "searchParam", "extension"},
			{"rest", "searchParam", "extension"},
			{"rest", "operation", "extension"},
			{"document", "extension"},
			{"rest", "interaction", "extension"},
		}
	}

	if extensionString == "capabilitystatement-expectation" {
		row1 := []string{"rest", "resource", "_searchInclude", "extension"}
		row2 := []string{"rest", "resource", "_searchRevInclude", "extension"}
		extensionList = append(extensionList, row1, row2)
	}

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
