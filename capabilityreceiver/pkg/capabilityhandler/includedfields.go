package capabilityhandler

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
)

// List of capability statement fields that are arrays of interfaces
var arrayFields = []string{"rest", "resource", "interaction", "searchParam", "operation", "document", "_searchInclude", "_searchRevInclude", "messaging", "supportedMessage"}

// from https://www.hl7.org/fhir/codesystem-FHIR-version.html
// looking at official and release versions only
var dstu2 = []string{"0.4.0", "0.5.0", "1.0.0", "1.0.1", "1.0.2"}
var stu3 = []string{"1.1.0", "1.2.0", "1.4.0", "1.6.0", "1.8.0", "3.0.0", "3.0.1", "3.0.2"}
var r4 = []string{"3.2.0", "3.3.0", "3.5.0", "3.5a.0", "4.0.0", "4.0.1"}

// RunIncludedFieldsAndExtensionsChecks returns an interface that contains information about whether fields and extensions are supported or not
func RunIncludedFieldsAndExtensionsChecks(capInt map[string]interface{}, fhirVersion string) []endpointmanager.IncludedField {
	if capInt == nil {
		return nil
	}

	var includedFields []endpointmanager.IncludedField
	includedFields = RunIncludedFieldsChecks(capInt, includedFields, fhirVersion)
	includedFields = RunIncludedExtensionsChecks(capInt, includedFields)
	return includedFields
}

// RunIncludedFieldsChecks stores whether each field in capability statement is populated or not populated
func RunIncludedFieldsChecks(capInt map[string]interface{}, includedFields []endpointmanager.IncludedField, fhirVersion string) []endpointmanager.IncludedField {
	fieldsList := getFieldsList(fhirVersion)
	// Get name of field
	for _, fieldNames := range fieldsList {
		var stringIndex string
		length := len(fieldNames)
		if length != 1 {
			for index, name := range fieldNames {
				if index == length-1 {
					stringIndex = stringIndex + name
				} else if index == 0 {
					stringIndex = name + "."
				} else {
					stringIndex = stringIndex + name + "."
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
		length := len(fieldNames)

		if index == (len(fieldNames) - 1) {
			return field != nil
		} else if arrContains(arrayFields, name) {
			fieldArr := field.([]interface{})
			nextIndex := index + 1
			return checkFieldArr(fieldNames[nextIndex:length], fieldArr)
		} else {
			capInt = field.(map[string]interface{})
		}
	}

	return false
}

// Given an array of interface objects, loops through each object to check whether the field is populated following the path of fieldNames
func checkFieldArr(fieldNames []string, fieldArr []interface{}) bool {
	var found bool
	length := len(fieldNames)
	name := fieldNames[0]
	interfaceArrTrue := arrContains(arrayFields, name)
	// Loop through the array of interface objects
	for _, obj := range fieldArr {
		// For each object in interface array, get desired field using name in fieldNames
		objMap := obj.(map[string]interface{})
		field := objMap[name]
		if field == nil {
			// If the desired field does not exist in that object, continue to the next object within the array of interface objects
			continue
		} else if length != 1 && interfaceArrTrue {
			// If the desired field is is not the last field to check and an array of interface objects, call checkFieldArr with this new array
			fieldArr := field.([]interface{})
			found = checkFieldArr(fieldNames[1:length], fieldArr)
			if found {
				return found
			}
		} else if length != 1 && !interfaceArrTrue {
			// If the desired field is not the last field to check and not an array of interface objects, call checkField with this field map[string]interface
			field := field.(map[string]interface{})
			found = checkField(field, fieldNames[1:length])
			if found {
				return found
			}
		} else {
			found = true
			return found
		}
	}
	return found
}

// RunIncludedExtensionsChecks stores whether each extension in capability statement is populated or not populated
func RunIncludedExtensionsChecks(capInt map[string]interface{}, includedFields []endpointmanager.IncludedField) []endpointmanager.IncludedField {

	extensionList := [][]string{
		{"extension", "http://hl7.org/fhir/StructureDefinition/conformance-supported-system", "conformance-supported-system"},
		{"rest", "resource", "extension", "http://hl7.org/fhir/StructureDefinition/conformance-search-parameter-combination", "conformance-search-parameter-combination"},
		{"rest", "resource", "interaction", "extension", "http://hl7.org/fhir/StructureDefinition/conformance-expectation", "conformance-expectation"},
		{"rest", "resource", "searchParam", "extension", "http://hl7.org/fhir/StructureDefinition/conformance-expectation", "conformance-expectation"},
		{"rest", "searchParam", "extension", "http://hl7.org/fhir/StructureDefinition/conformance-expectation", "conformance-expectation"},
		{"rest", "operation", "extension", "http://hl7.org/fhir/StructureDefinition/conformance-expectation", "conformance-expectation"},
		{"document", "extension", "http://hl7.org/fhir/StructureDefinition/conformance-expectation", "conformance-expectation"},
		{"rest", "interaction", "extension", "http://hl7.org/fhir/StructureDefinition/conformance-expectation", "conformance-expectation"},
		{"rest", "resource", "interaction", "modifierExtension", "http://hl7.org/fhir/StructureDefinition/conformance-prohibited", "conformance-prohibited"},
		{"rest", "resource", "searchParam", "modifierExtension", "http://hl7.org/fhir/StructureDefinition/conformance-prohibited", "conformance-prohibited"},
		{"rest", "searchParam", "modifierExtension", "http://hl7.org/fhir/StructureDefinition/conformance-prohibited", "conformance-prohibited"},
		{"rest", "operation", "modifierExtension", "http://hl7.org/fhir/StructureDefinition/conformance-prohibited", "conformance-prohibited"},
		{"document", "modifierExtension", "http://hl7.org/fhir/StructureDefinition/conformance-prohibited", "conformance-prohibited"},
		{"rest", "interaction", "modifierExtension", "http://hl7.org/fhir/StructureDefinition/conformance-prohibited", "conformance-prohibited"},
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
		{"rest", "resource", "interaction", "extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-expectation", "capabilitystatement-expectation"},
		{"rest", "resource", "searchParam", "extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-expectation", "capabilitystatement-expectation"},
		{"rest", "searchParam", "extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-expectation", "capabilitystatement-expectation"},
		{"rest", "operation", "extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-expectation", "capabilitystatement-expectation"},
		{"document", "extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-expectation", "capabilitystatement-expectation"},
		{"rest", "interaction", "extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-expectation", "capabilitystatement-expectation"},
		{"rest", "resource", "_searchInclude", "extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-expectation", "capabilitystatement-expectation"},
		{"rest", "resource", "_searchRevInclude", "extension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-expectation", "capabilitystatement-expectation"},
		{"rest", "resource", "interaction", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-prohibited", "modifierExtension", "capabilitystatement-prohibited"},
		{"rest", "resource", "searchParam", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-prohibited", "modifierExtension", "capabilitystatement-prohibited"},
		{"rest", "searchParam", "modifierExtension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-prohibited", "capabilitystatement-prohibited"},
		{"rest", "operation", "modifierExtension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-prohibited", "capabilitystatement-prohibited"},
		{"document", "modifierExtension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-prohibited", "capabilitystatement-prohibited"},
		{"rest", "interaction", "modifierExtension", "http://hl7.org/fhir/StructureDefinition/capabilitystatement-prohibited", "capabilitystatement-prohibited"},
	}

	// Get name of extension and create extensionObj with extension name, if the extension exists, and if it is an extension
	for _, extensionPath := range extensionList {
		extensionName := extensionPath[len(extensionPath)-1]
		extensionURL := extensionPath[len(extensionPath)-2]
		// Check if includedFields already contains this extension
		index := includedFieldsContains(includedFields, extensionName)
		if index != -1 {
			//If includedFields contains extension but Exists is false, check next possible location to see if it exists
			if !includedFields[index].Exists {
				includedFields[index].Exists = checkExtension(capInt, extensionPath, extensionURL)
			} else {
				continue
			}
		} else {
			extensionObj := endpointmanager.IncludedField{
				Field:     extensionName,
				Exists:    checkExtension(capInt, extensionPath, extensionURL),
				Extension: true,
			}
			includedFields = append(includedFields, extensionObj)
		}
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
		length := len(fieldNames)
		// Check if at an extension field in fieldNames
		if index == length-3 {
			extensionArr := field.([]interface{})
			return checkExtensionURL(extensionArr, url)
		} else if arrContains(arrayFields, name) {
			fieldArr := field.([]interface{})
			nextIndex := index + 1
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
	length := len(fieldNames)
	name := fieldNames[0]
	interfaceArrTrue := arrContains(arrayFields, name)
	// Loop through the array of interface objects
	for _, obj := range fieldArr {
		// For each object in interface array, get desired field using name in fieldNames
		objMap := obj.(map[string]interface{})
		extensionField := objMap[name]
		if extensionField == nil {
			// If the desired field does not exist in that object, continue to the next object within the array of interface objects
			continue
		} else if length-3 != 0 && interfaceArrTrue {
			// If the desired field is not extension or modifierExtension and it is also an array of interface objects, call checkArrFieldExtension with this new array
			fieldArr := extensionField.([]interface{})
			found = checkArrFieldExtension(fieldNames[1:length], fieldArr, url)
			if found {
				return found
			}
		} else if length-3 != 0 && !interfaceArrTrue {
			// If the desired field is not extension or modifierExtension and it is not an array of interface objects, call checkExtension with this field map[string]interface
			extensionField := extensionField.(map[string]interface{})
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

func arrContains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// Checks if includedFields array already contains an extension with extensionName
func includedFieldsContains(includedFields []endpointmanager.IncludedField, extensionName string) int {
	for index, fieldObj := range includedFields {
		if fieldObj.Field == extensionName {
			return index
		}
	}
	return -1
}

func getFieldsList(fhirVersion string) [][]string {
	baseFieldsList := [][]string{
		{"url"},
		{"version"},
		{"name"},
		{"status"},
		{"experimental"},
		{"publisher"},
		{"contact"},
		{"date"},
		{"description"},
		{"copyright"},
		{"kind"},
		{"software", "name"},
		{"software", "version"},
		{"software", "releaseDate"},
		{"implementation", "description"},
		{"implementation", "url"},
		{"fhirVersion"},
		{"format"},
		{"rest", "mode"},
		{"rest", "resource", "type"},
		{"rest", "resource", "interaction", "code"},
		{"rest", "resource", "versioning"},
		{"rest", "resource", "conditionalDelete"},
		{"rest", "resource", "searchParam", "type"},
		{"rest", "interaction", "code"},
		{"document"},
		{"document", "mode"},
		{"messaging"},
	}

	DSTU2OnlyFields := [][]string{
		{"requirements"},
	}

	DSTU2FieldsList := [][]string{
		{"profile"},
		{"acceptUnknown"},
	}
	STU3FieldsList := [][]string{
		{"title"},
		{"useContext"},
		{"jurisdiction"},
		{"purpose"},
		{"instantiates"},
		{"patchFormat"},
		{"implementationGuide"},
		{"rest", "resource", "conditionalRead"},
		{"rest", "resource", "referencePolicy"},
		{"messaging", "supportedMessage", "mode"},
	}

	R4FieldsList := [][]string{
		{"imports"},
		{"implementation", "custodian"},
	}

	if helpers.StringArrayContains(dstu2, fhirVersion) {
		DSTU2Fields := append(baseFieldsList, DSTU2OnlyFields...)
		DSTU2Fields = append(DSTU2Fields, DSTU2FieldsList...)
		return DSTU2Fields
	} else if helpers.StringArrayContains(stu3, fhirVersion) {
		STU3Fields := append(baseFieldsList, DSTU2FieldsList...)
		STU3Fields = append(STU3Fields, STU3FieldsList...)
		return STU3Fields
	} else if helpers.StringArrayContains(r4, fhirVersion) {
		R4Fields := append(baseFieldsList, STU3FieldsList...)
		R4Fields = append(R4Fields, R4FieldsList...)
		return R4Fields
	} else {
		return baseFieldsList
	}
}
