package capabilityhandler

<<<<<<< HEAD
// RunIncludedFieldsChecks stores whether each field in capability statement is populated or not populated
func RunIncludedFieldsChecks(capInt map[string]interface{}) map[string]bool {
	if capInt == nil {
		return nil
	}
	includedFields := make(map[string]bool)

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

		includedFields[stringIndex] = checkField(capInt, fieldNames)
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

=======
import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
)

var DSTU2 = []string{"1.0.0", "1.0.1", "1.0.2"}
var STU3 = []string{"3.0.0", "3.0.1", "3.0.2"}
var R4 = []string{"4.0.0", "4.0.1"}

func RunIncludedFieldsChecks(capStat capabilityparser.CapabilityStatement) map[string]interface{} {
	includedFields := make(map[string]interface{})
	var fhirVersion string
	if capStat != nil {
		fhirVersion, _ = capStat.GetFHIRVersion()
	} else {
		return nil
	}

	if contains(DSTU2, fhirVersion) {
		includedFields = DSTU2IncludedFields(capStat)
	} else if contains(STU3, fhirVersion) {
		includedFields = STU3IncludedFields(capStat)
	} else if contains(R4, fhirVersion) {
		includedFields = R4IncludedFields(capStat)
	} else {
		//do something
	}

	return includedFields
}

func DSTU2IncludedFields(capStat capabilityparser.CapabilityStatement) map[string]interface{} {
	includedFields := make(map[string]interface{})
	includedFields = commonFieldsChecksAll(capStat)
	return includedFields
}

func STU3IncludedFields(capStat capabilityparser.CapabilityStatement) map[string]interface{} {
	includedFields := make(map[string]interface{})
	includedFields = commonFieldsChecksAll(capStat)

	return includedFields
}

func R4IncludedFields(capStat capabilityparser.CapabilityStatement) map[string]interface{} {
	includedFields := make(map[string]interface{})
	includedFields = commonFieldsChecksAll(capStat)

	return includedFields
}

func commonFieldsChecksAll(capStat capabilityparser.CapabilityStatement) map[string]interface{} {
	includedFields := make(map[string]interface{})
	includedFields["publisher"] = checkIncluded(capStat.GetPublisher())
	includedFields["fhirVersion"] = checkIncluded(capStat.GetFHIRVersion())
	includedFields["software.name"] = checkIncluded(capStat.GetSoftwareName())
	includedFields["software.version"] = checkIncluded(capStat.GetSoftwareVersion())
	includedFields["copyright"] = checkIncluded(capStat.GetCopyright())
	return includedFields
}

func checkIncluded(resp string, err error) bool {
	if resp != "" && err == nil {
		return true
	}
>>>>>>> 975432d... Creating file for checking for included fields
	return false
}
