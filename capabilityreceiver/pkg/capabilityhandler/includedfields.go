package capabilityhandler

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
	return false
}
