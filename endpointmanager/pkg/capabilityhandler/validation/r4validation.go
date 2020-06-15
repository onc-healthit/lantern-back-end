package validation

import (
	"strconv"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

type r4Validation struct {
	baseVal
}

func newR4Val() *r4Validation {
	return &r4Validation{
		baseVal: baseVal{},
	}
}

func (v *r4Validation) RunValidation(capStat capabilityparser.CapabilityStatement, httpResponse int, mimeTypes []string, fhirVersion string) endpointmanager.Validation {
	var validationResults []endpointmanager.Rule
	validationWarnings := make([]endpointmanager.Rule, 0)

	returnedRule := v.CapStatExists(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.MimeTypeValid(mimeTypes, fhirVersion)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.HTTPResponseValid(httpResponse)
	validationResults = append(validationResults, returnedRule)

	validations := endpointmanager.Validation{
		Results:  validationResults,
		Warnings: validationWarnings,
	}

	return validations
}

func (v *r4Validation) CapStatExists(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.CapStatExists(capStat)
	baseRule.Comment = "Servers SHALL provide a Capability Statement that specifies which interactions and resources are supported."
	baseRule.Reference = "http://hl7.org/fhir/http.html"
	baseRule.ImplGuide = "USCore 3.1"
	return baseRule
}

func (v *r4Validation) MimeTypeValid(mimeTypes []string, fhirVersion string) endpointmanager.Rule {
	baseRule := v.baseVal.MimeTypeValid(mimeTypes, fhirVersion)
	baseRule.Reference = "http://hl7.org/fhir/http.html"
	baseRule.ImplGuide = "USCore 3.1"
	return baseRule
}

func (v *r4Validation) HTTPResponseValid(httpResponse int) endpointmanager.Rule {
	baseRule := v.baseVal.HTTPResponseValid(httpResponse)
	baseRule.Reference = "http://hl7.org/fhir/http.html"
	baseRule.ImplGuide = "USCore 3.1"
	if httpResponse != 0 && httpResponse != 200 {
		strResp := strconv.Itoa(httpResponse)
		baseRule.Comment = "The HTTP response code was " + strResp + " instead of 200. Applications SHALL return a resource that describes the functionality of the server end-point."
	}
	return baseRule
}
