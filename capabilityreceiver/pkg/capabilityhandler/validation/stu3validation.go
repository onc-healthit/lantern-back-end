package validation

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/smartparser"
)

type stu3Validation struct {
	baseVal
}

func newSTU3Val() *stu3Validation {
	return &stu3Validation{
		baseVal: baseVal{},
	}
}

// RunValidation runs all of the defined validation checks
func (v *stu3Validation) RunValidation(capStat capabilityparser.CapabilityStatement,
	mimeTypes []string,
	fhirVersion string,
	tlsVersion string,
	smartRsp smartparser.SMARTResponse,
	requestedFhirVersion string,
	defaultFhirVersion string) endpointmanager.Validation {
	var validationResults []endpointmanager.Rule

	returnedRule := v.CapStatExists(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.MimeTypeValid(mimeTypes, fhirVersion)
	validationResults = append(validationResults, returnedRule)

	returnedRules := v.KindValid(capStat)
	validationResults = append(validationResults, returnedRules[0])

	validations := endpointmanager.Validation{
		Results: validationResults,
	}

	return validations
}

// CapStatExists checks if the capability statement exists using the base function, and then
// adds specific STU3 reference information
func (v *stu3Validation) CapStatExists(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.CapStatExists(capStat)
	baseRule.Reference = "http://hl7.org/fhir/http.html"
	return baseRule
}

// MimeTypeValid checks if the given mime types include the correct mime type for the given version
// using the base function, and then adds specific STU3 reference information
func (v *stu3Validation) MimeTypeValid(mimeTypes []string, fhirVersion string) endpointmanager.Rule {
	baseRule := v.baseVal.MimeTypeValid(mimeTypes, fhirVersion)
	baseRule.Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"
	return baseRule
}

// KindValid checks the rule that kind = instance since all of the endpoints we are looking
// at are for server instances, and then adds specific STU3 reference information
func (v *stu3Validation) KindValid(capStat capabilityparser.CapabilityStatement) []endpointmanager.Rule {
	baseRule := v.baseVal.KindValid(capStat)
	baseRule[0].Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"
	return baseRule
}
