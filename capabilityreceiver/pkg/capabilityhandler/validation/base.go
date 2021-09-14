package validation

import (
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/smartparser"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// from https://www.hl7.org/fhir/codesystem-FHIR-version.html
// looking at official and release versions only
var version3plus = []string{"3.0.0", "3.0.1", "3.0.2", "3.2.0", "3.3.0", "3.5.0", "3.5a.0", "4.0.0", "4.0.1"}
var fhir3PlusJSONMIMEType = "application/fhir+json"
var fhir2LessJSONMIMEType = "application/json+fhir"

type baseVal struct {
}

// RunValidation runs all of the defined validation checks
func (bv *baseVal) RunValidation(capStat capabilityparser.CapabilityStatement,
	mimeTypes []string,
	fhirVersion string,
	tlsVersion string,
	smartRsp smartparser.SMARTResponse) endpointmanager.Validation {
	var validationResults []endpointmanager.Rule

	returnedRule := bv.CapStatExists(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = bv.MimeTypeValid(mimeTypes, fhirVersion)
	validationResults = append(validationResults, returnedRule)

	returnedRules := bv.KindValid(capStat)
	validationResults = append(validationResults, returnedRules[0])

	validations := endpointmanager.Validation{
		Results: validationResults,
	}

	return validations
}

// CapStatExists checks if the capability statement exists
func (bv *baseVal) CapStatExists(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.CapStatExistRule,
		Valid:     true,
		Expected:  "true",
		Actual:    "true",
		Reference: "http://hl7.org/fhir/DSTU2/conformance.html",
		Comment:   "The Capability Statement exists.",
	}

	if capStat != nil {
		return ruleError
	}

	ruleError.Valid = false
	ruleError.Actual = "false"
	ruleError.Comment = "The Capability Statement does not exist."
	return ruleError
}

// MimeTypeValid checks if the given mime types include the correct mime type for the given version
func (bv *baseVal) MimeTypeValid(mimeTypes []string, fhirVersion string) endpointmanager.Rule {
	mimeString := strings.Join(mimeTypes, ",")
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.GeneralMimeTypeRule,
		Valid:     true,
		Expected:  "",
		Actual:    mimeString,
		Reference: "http://hl7.org/fhir/DSTU2/conformance.html",
		Comment:   "",
	}

	if len(mimeTypes) == 0 {
		ruleError.Valid = false
		ruleError.Expected = "N/A"
		ruleError.Comment = "No mime type given; cannot validate mime type."
		return ruleError
	}

	if len(fhirVersion) == 0 {
		ruleError.Valid = false
		ruleError.Expected = "N/A"
		ruleError.Comment = "Unknown FHIR Version; cannot validate mime type."
		return ruleError
	}

	var mimeError string
	for _, mt := range mimeTypes {
		if helpers.StringArrayContains(version3plus, fhirVersion) {
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

func (bv *baseVal) TLSVersion(tlsVersion string) endpointmanager.Rule {
	var ruleError endpointmanager.Rule
	return ruleError
}

func (bv *baseVal) PatientResourceExists(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	var ruleError endpointmanager.Rule
	return ruleError
}

func (bv *baseVal) OtherResourceExists(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	var ruleError endpointmanager.Rule
	return ruleError
}

func (bv *baseVal) SmartResponseExists(smartRsp smartparser.SMARTResponse) endpointmanager.Rule {
	var ruleError endpointmanager.Rule
	return ruleError
}

// KindValid checks the rule that kind = instance since all of the endpoints we are looking
// at are for server instances.
func (bv *baseVal) KindValid(capStat capabilityparser.CapabilityStatement) []endpointmanager.Rule {
	baseComment := "Kind value should be set to 'instance' because this is a specific system instance."
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.KindRule,
		Valid:     true,
		Expected:  "instance",
		Reference: "http://hl7.org/fhir/DSTU2/conformance.html",
		Comment:   baseComment,
	}
	if capStat == nil {
		ruleError.Valid = false
		ruleError.Comment = "Capability Statement does not exist; cannot check kind value. " + baseComment
		returnVal := []endpointmanager.Rule{
			ruleError,
		}
		return returnVal
	}
	kind, err := capStat.GetKind()
	if err != nil || len(kind) == 0 {
		ruleError.Valid = false
	}
	ruleError.Actual = kind
	returnVal := []endpointmanager.Rule{
		ruleError,
	}
	return returnVal
}

func (bv *baseVal) MessagingEndpointValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	var ruleError endpointmanager.Rule
	return ruleError
}

func (bv *baseVal) EndpointFunctionValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	var ruleError endpointmanager.Rule
	return ruleError
}

func (bv *baseVal) DescribeEndpointValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	var ruleError endpointmanager.Rule
	return ruleError
}

func (bv *baseVal) DocumentSetValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	var ruleError endpointmanager.Rule
	return ruleError
}

func (bv *baseVal) UniqueResources(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	var ruleError endpointmanager.Rule
	return ruleError
}

func (bv *baseVal) SearchParamsUnique(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	var ruleError endpointmanager.Rule
	return ruleError
}
