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
	fhirVersion string,
	tlsVersion string,
	smartRsp smartparser.SMARTResponse,
	requestedFhirVersion string,
	defaultFhirVersion string) endpointmanager.Validation {
	var validationResults []endpointmanager.Rule

	returnedRule := v.CapStatExists(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRules := v.KindValid(capStat)
	validationResults = append(validationResults, returnedRules[0])

	returnedRule = v.DescribeEndpointValid(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.DocumentSetValid(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.EndpointFunctionValid(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.MessagingEndpointValid(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.UniqueResources(capStat)
	validationResults = append(validationResults, returnedRule)

	validations := endpointmanager.Validation{
		Results: validationResults,
	}

	return validations
}

// CapStatExists checks if the capability statement exists using the base function, and then
// adds specific STU3 reference information
func (v *stu3Validation) CapStatExists(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseComment := "Servers SHALL provide a Capability Statement that specifies which interactions and resources are supported."

	baseRule := v.baseVal.CapStatExists(capStat)
	baseRule.Reference = "http://hl7.org/fhir/http.html"

	if baseRule.Valid {
		baseRule.Comment = "The Capability Statement exists. " + baseComment
	} else {
		baseRule.Comment = "The Capability Statement does not exist. " + baseComment
	}
	return baseRule
}

// KindValid checks the rule that kind = instance since all of the endpoints we are looking
// at are for server instances, and then adds specific STU3 reference information
func (v *stu3Validation) KindValid(capStat capabilityparser.CapabilityStatement) []endpointmanager.Rule {
	baseComment := "Kind value should be set to 'instance' because this is a specific system instance."

	baseRule := v.baseVal.KindValid(capStat)
	baseRule[0].Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"

	if capStat == nil {
		baseRule[0].Comment = "Capability Statement does not exist; cannot check kind value. " + baseComment
	}

	return baseRule
}

// DescribeEndpointValid checks the requirement: "A Capability Statement SHALL have at least one of description,
// software, or implementation element."
func (v *stu3Validation) DescribeEndpointValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.DescribeEndpointValid(capStat)
	baseRule.Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"
	baseRule.Comment = "A Capability Statement SHALL have at least one of description, software, or implementation element."

	if capStat == nil {
		baseRule.Comment = "The Capability Statement does not exist; cannot check description, software, or implementation elements."
	}

	return baseRule
}

// DocumentSetValid checks the requirement: "The set of documents must be unique by the combination of profile and mode."
func (v *stu3Validation) DocumentSetValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.DocumentSetValid(capStat)
	baseRule.Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"

	if capStat == nil {
		baseRule.Comment = "The Capability Statement does not exist; cannot check documents."
	}

	return baseRule
}

// EndpointFunctionValid checks the requirement "A Capability Statement SHALL have at least one of REST,
// messaging or document element."
func (v *stu3Validation) EndpointFunctionValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.EndpointFunctionValid(capStat)
	baseRule.Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"
	baseRule.Comment = "A Capability Statement SHALL have at least one of REST, messaging or document element."

	if capStat == nil {
		baseRule.Comment = "The Capability Statement does not exist; cannot check REST, messaging or document elements."
	}

	return baseRule
}

// MessagingEndpointValid checks the requirement "Messaging endpoint is required (and is only permitted) when a statement is for an implementation."
// Every endpoint we are testing should be an implementation, which means the endpoint field should be there.
func (v *stu3Validation) MessagingEndpointValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseKindComment := "Kind value should be set to 'instance' because this is a specific system instance."
	baseMessagingComment := "Messaging end-point is required (and is only permitted) when a statement is for an implementation. This endpoint must be an implementation."

	baseRule := v.baseVal.MessagingEndpointValid(capStat)
	baseRule.Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"

	if capStat == nil {
		baseRule.Comment = "Capability Statement does not exist; cannot check kind value. " + baseKindComment + " " + baseMessagingComment
	}
	return baseRule
}

// UniqueResources checks the requirement: "A given resource can only be described once per RESTful mode."
func (v *stu3Validation) UniqueResources(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.UniqueResources(capStat)
	baseRule.Reference = "http://hl7.org/fhir/STU3/capabilitystatement.html"
	return baseRule
}
