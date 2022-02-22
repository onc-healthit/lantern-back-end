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
	return baseRule
}

// KindValid checks the rule that kind = instance since all of the endpoints we are looking
// at are for server instances, and then adds specific STU3 reference information
func (v *stu3Validation) KindValid(capStat capabilityparser.CapabilityStatement) []endpointmanager.Rule {
	baseRule := v.baseVal.KindValid(capStat)
	baseRule[0].Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"
	return baseRule
}

// DescribeEndpointValid checks the requirement: "A Capability Statement SHALL have at least one of description,
// software, or implementation element."
func (v *stu3Validation) DescribeEndpointValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.DescribeEndpointValid(capStat)
	baseRule.Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"
	baseRule.Comment = "A Capability Statement SHALL have at least one of description, software, or implementation element."
	return baseRule
}

// DocumentSetValid checks the requirement: "The set of documents must be unique by the combination of profile and mode."
func (v *stu3Validation) DocumentSetValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.DocumentSetValid(capStat)
	baseRule.Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"
	return baseRule
}

// EndpointFunctionValid checks the requirement "A Capability Statement SHALL have at least one of REST,
// messaging or document element."
func (v *stu3Validation) EndpointFunctionValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.EndpointFunctionValid(capStat)
	baseRule.Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"
	baseRule.Comment = "A Capability Statement SHALL have at least one of REST, messaging or document element."
	return baseRule
}

// MessagingEndpointValid checks the requirement "Messaging endpoint is required (and is only permitted) when a statement is for an implementation."
// Every endpoint we are testing should be an implementation, which means the endpoint field should be there.
func (v *stu3Validation) MessagingEndpointValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.MessagingEndpointValid(capStat)
	baseRule.Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"
	return baseRule
}

// UniqueResources checks the requirement: "A given resource can only be described once per RESTful mode."
func (v *stu3Validation) UniqueResources(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.UniqueResources(capStat)
	baseRule.Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"
	return baseRule
}
