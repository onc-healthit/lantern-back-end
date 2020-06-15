package validation

import (
	"strconv"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// from https://www.hl7.org/fhir/codesystem-FHIR-version.html
// looking at official and release versions only
var version3plus = []string{"3.0.0", "3.0.1", "4.0.0", "4.0.1"}
var fhir3PlusJSONMIMEType = "application/fhir+json"
var fhir2LessJSONMIMEType = "application/json+fhir"

type baseVal struct {
}

func (bv *baseVal) RunValidation(capStat capabilityparser.CapabilityStatement, httpResponse int, mimeTypes []string, fhirVersion string) endpointmanager.Validation {
	var validationResults []endpointmanager.Rule
	validationWarnings := make([]endpointmanager.Rule, 0)

	returnedRule := bv.CapStatExists(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = bv.MimeTypeValid(mimeTypes, fhirVersion)
	validationResults = append(validationResults, returnedRule)

	returnedRule = bv.HTTPResponseValid(httpResponse)
	validationResults = append(validationResults, returnedRule)

	validations := endpointmanager.Validation{
		Results:  validationResults,
		Warnings: validationWarnings,
	}

	return validations
}

func (bv *baseVal) CapStatExists(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	ruleError := endpointmanager.Rule{
		RuleName: endpointmanager.CapStatExistRule,
		Valid:    true,
		Expected: "true",
		Actual:   "true",
	}

	if capStat != nil {
		return ruleError
	}

	ruleError.Valid = false
	ruleError.Actual = "false"
	return ruleError
}

func (bv *baseVal) MimeTypeValid(mimeTypes []string, fhirVersion string) endpointmanager.Rule {
	mimeString := strings.Join(mimeTypes, ",")
	ruleError := endpointmanager.Rule{
		RuleName: endpointmanager.GeneralMimeTypeRule,
		Valid:    true,
		Expected: "",
		Actual:   mimeString,
		Comment:  "",
	}

	if len(fhirVersion) == 0 {
		ruleError.Valid = false
		ruleError.Expected = "N/A"
		ruleError.Comment = "Unknown FHIR Version; cannot validate mime type."
		return ruleError
	}

	var mimeError string
	for _, mt := range mimeTypes {
		if contains(version3plus, fhirVersion) {
			if mt == fhir3PlusJSONMIMEType {
				ruleError.Expected = fhir3PlusJSONMIMEType
				ruleError.Comment = "FHIR Version " + fhirVersion + " requires the Mime Type to be " + fhir3PlusJSONMIMEType
				return ruleError
			}
			mimeError = fhir3PlusJSONMIMEType
		} else {
			// The fhirVersion has to be valid in order to create a valid capability statement
			// so if it's gotten this far, the fhirVersion has to be less than 3
			if mt == fhir2LessJSONMIMEType {
				ruleError.Expected = fhir2LessJSONMIMEType
				ruleError.Comment = "FHIR Version " + fhirVersion + " requires the Mime Type to be " + fhir2LessJSONMIMEType
				return ruleError
			}
			mimeError = fhir2LessJSONMIMEType
		}
	}

	ruleError.Valid = false
	ruleError.Expected = mimeError
	ruleError.Comment = "FHIR Version " + fhirVersion + " requires the Mime Type to be " + mimeError
	return ruleError
}

func (bv *baseVal) HTTPResponseValid(httpResponse int) endpointmanager.Rule {
	strResp := strconv.Itoa(httpResponse)
	ruleError := endpointmanager.Rule{
		RuleName: endpointmanager.HTTPResponseRule,
		Valid:    true,
		Expected: "200",
		Actual:   strResp,
		Comment:  "",
	}

	if httpResponse == 200 {
		return ruleError
	}

	if httpResponse == 0 {
		ruleError.Comment = "The GET request failed with no returned HTTP response status code."
	}

	ruleError.Valid = false
	return ruleError
}
