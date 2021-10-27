package validation

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
)

// from https://www.hl7.org/fhir/codesystem-FHIR-version.html
// looking at official and release versions only
var dstu2 = []string{"0.4.0", "0.5.0", "1.0.0", "1.0.1", "1.0.2"}
var stu3 = []string{"1.1.0", "1.2.0", "1.4.0", "1.6.0", "1.8.0", "3.0.0", "3.0.1", "3.0.2"}
var r4 = []string{"3.2.0", "3.3.0", "3.5.0", "3.5a.0", "4.0.0", "4.0.1"}

// Validator is an interface that can be implemented for each FHIR Version to run the correct
// version's validation checks
type Validator interface {
	RunValidation(capabilityparser.CapabilityStatement, int, []string, string, string, int, string, string) endpointmanager.Validation
	CapStatExists(capabilityparser.CapabilityStatement) endpointmanager.Rule
	MimeTypeValid([]string, string) endpointmanager.Rule
	HTTPResponseValid(int) endpointmanager.Rule
	FhirVersion(string) endpointmanager.Rule
	VersionResponseValid(string, string) endpointmanager.Rule
	TLSVersion(string) endpointmanager.Rule
	PatientResourceExists(capabilityparser.CapabilityStatement) endpointmanager.Rule
	OtherResourceExists(capabilityparser.CapabilityStatement) endpointmanager.Rule
	SmartHTTPResponseValid(int) endpointmanager.Rule
	KindValid(capabilityparser.CapabilityStatement) []endpointmanager.Rule
	MessagingEndpointValid(capabilityparser.CapabilityStatement) endpointmanager.Rule
	EndpointFunctionValid(capabilityparser.CapabilityStatement) endpointmanager.Rule
	DescribeEndpointValid(capabilityparser.CapabilityStatement) endpointmanager.Rule
	DocumentSetValid(capabilityparser.CapabilityStatement) endpointmanager.Rule
	UniqueResources(capabilityparser.CapabilityStatement) endpointmanager.Rule
	SearchParamsUnique(capabilityparser.CapabilityStatement) endpointmanager.Rule
}

// ValidatorForFHIRVersion checks the given fhir version and returns the specific validator
// for that version, which can be used for running the Validation checks.
// To note: All but the newR4Val() function returns the base validation currently
func ValidatorForFHIRVersion(fhirVersion string) Validator {
	if fhirVersion == "" {
		return newUnknownVal()
	}

	if helpers.StringArrayContains(dstu2, fhirVersion) {
		return newDSTU2Val()
	} else if helpers.StringArrayContains(stu3, fhirVersion) {
		return newSTU3Val()
	} else if helpers.StringArrayContains(r4, fhirVersion) {
		return newR4Val()
	}

	return newUnknownVal()
}
