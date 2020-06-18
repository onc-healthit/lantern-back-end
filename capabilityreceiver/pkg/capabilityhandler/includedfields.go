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
	"fmt"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
)

var DSTU2 = []string{"1.0.0", "1.0.1", "1.0.2"}
var STU3 = []string{"3.0.0", "3.0.1", "3.0.2"}
var R4 = []string{"4.0.0", "4.0.1"}

func RunIncludedFieldsChecks(capStat capabilityparser.CapabilityStatement, capInt map[string]interface{}) map[string]bool {
	includedFields := make(map[string]bool)
	var fhirVersion string
	if capStat != nil {
		fhirVersion, _ = capStat.GetFHIRVersion()
	} else {
		return nil
	}

	if contains(DSTU2, fhirVersion) {
		includedFields = DSTU2IncludedFields(capStat, capInt)
	} else if contains(STU3, fhirVersion) {
		includedFields = STU3IncludedFields(capStat, capInt)
	} else if contains(R4, fhirVersion) {
		includedFields = R4IncludedFields(capStat, capInt)
	} else {
		//do something
	}

	return includedFields
}

func DSTU2IncludedFields(capStat capabilityparser.CapabilityStatement, capInt map[string]interface{}) map[string]bool {
	includedFields := make(map[string]bool)
	includedFields = commonFieldsChecksAll(capStat, capInt)
	return includedFields
}

func STU3IncludedFields(capStat capabilityparser.CapabilityStatement, capInt map[string]interface{}) map[string]bool {
	includedFields := make(map[string]bool)
	includedFields = commonFieldsChecksAll(capStat, capInt)

	return includedFields
}

func R4IncludedFields(capStat capabilityparser.CapabilityStatement, capInt map[string]interface{}) map[string]bool {
	includedFields := make(map[string]bool)
	includedFields = commonFieldsChecksAll(capStat, capInt)

	return includedFields
}

func commonFieldsChecksAll(capStat capabilityparser.CapabilityStatement, capInt map[string]interface{}) map[string]bool {
	includedFields := make(map[string]bool)
	includedFields["url"] = checkIncluded(GetURL(capInt))
	includedFields["version"] = checkIncluded(GetVersion(capInt))
	includedFields["name"] = checkIncluded(GetName(capInt))
	includedFields["title"] = checkIncluded(GetTitle(capInt))
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

// GetURL returns the url field from the conformance/capability statement.
func GetURL(capInt map[string]interface{}) (string, error) {
	url := capInt["url"]
	if url == nil {
		return "", nil
	}
	urlStr, ok := url.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement url value to a string", capInt["url"])
	}
	return urlStr, nil
}

// GetURL returns the version field from the conformance/capability statement.
func GetVersion(capInt map[string]interface{}) (string, error) {
	version := capInt["version"]
	if version == nil {
		return "", nil
	}
	versionStr, ok := version.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement version value to a string", capInt["version"])
	}
	return versionStr, nil
}

// GetURL returns the name field from the conformance/capability statement.
func GetName(capInt map[string]interface{}) (string, error) {
	name := capInt["name"]
	if name == nil {
		return "", nil
	}
	nameStr, ok := name.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement name value to a string", capInt["name"])
	}
	return nameStr, nil
}

// GetURL returns the title field from the conformance/capability statement.
func GetTitle(capInt map[string]interface{}) (string, error) {
	title := capInt["title"]
	if title == nil {
		return "", nil
	}
	titleStr, ok := title.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement title value to a string", capInt["title"])
	}
	return titleStr, nil
}

/*
// GetFHIRVersion returns the FHIR version specifiedin the conformance/capability statement.
func (cp *baseParser) GetFHIRVersion() (string, error) {
	fhirVersion := cp.capStat["fhirVersion"]
	if fhirVersion == nil {
		return "", nil
	}
	fhirVersionStr, ok := fhirVersion.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement fhirVersion value to a string", cp.version)
	}
	return fhirVersionStr, nil
}

// GetSoftwareName returns the software name specified in the conformance/capability statement.
func (cp *baseParser) GetSoftwareName() (string, error) {
	software := cp.capStat["software"]
	if software == nil {
		return "", nil
	}
	softwareMap, ok := software.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement software value to a map[string]interface{}", cp.version)
	}
	name := softwareMap["name"]
	if name == nil {
		return "", nil
	}
	nameStr, ok := name.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement software.name value to a string", cp.version)
	}
	return nameStr, nil
}*/
