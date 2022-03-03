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
	smartRsp smartparser.SMARTResponse,
	requestedFhirVersion string,
	defaultFhirVersion string) endpointmanager.Validation {
	var validationResults []endpointmanager.Rule

	returnedRule := bv.CapStatExists(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = bv.MimeTypeValid(mimeTypes, fhirVersion)
	validationResults = append(validationResults, returnedRule)

	returnedRules := bv.KindValid(capStat)
	validationResults = append(validationResults, returnedRules[0])

	returnedRule = bv.DescribeEndpointValid(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = bv.DocumentSetValid(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = bv.EndpointFunctionValid(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = bv.MessagingEndpointValid(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = bv.UniqueResources(capStat)
	validationResults = append(validationResults, returnedRule)

	validations := endpointmanager.Validation{
		Results: validationResults,
	}

	return validations
}

// CapStatExists checks if the capability statement exists
func (bv *baseVal) CapStatExists(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseComment := "Servers SHALL provide a Conformance Resource that specifies which interactions and resources are supported."
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.CapStatExistRule,
		Valid:     true,
		Expected:  "true",
		Actual:    "true",
		Reference: "http://hl7.org/fhir/http.html",
		Comment:   "The Conformance Resource exists. " + baseComment,
	}

	if capStat != nil {
		return ruleError
	}

	ruleError.Valid = false
	ruleError.Actual = "false"
	ruleError.Comment = "The Conformance Resource does not exist. " + baseComment
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
		Reference: "http://hl7.org/fhir/http.html",
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
		validMIMETypes := true

		// If no fhirVersion returned and all MIME types are valid, these MIME types saved from previous successful responses, so do not show them in actual result
		for _, mt := range mimeTypes {
			if mt != fhir2LessJSONMIMEType && mt != fhir3PlusJSONMIMEType {
				validMIMETypes = false
			}
		}
		if validMIMETypes {
			ruleError.Actual = ""
		}
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
		ruleError.Comment = "Conformance Resource does not exist; cannot check kind value. " + baseComment
		returnVal := []endpointmanager.Rule{
			ruleError,
		}
		return returnVal
	}
	kind, err := capStat.GetKind()
	if err != nil || len(kind) == 0 {
		ruleError.Valid = false
	}

	if kind != "instance" {
		ruleError.Valid = false
	}
	ruleError.Actual = kind
	returnVal := []endpointmanager.Rule{
		ruleError,
	}
	return returnVal
}

func (bv *baseVal) MessagingEndpointValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseComment := "Messaging end-point is required (and is only permitted) when a statement is for an implementation. This endpoint must be an implementation."
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.MessagingEndptRule,
		Valid:     false,
		Expected:  "true",
		Actual:    "false",
		Comment:   baseComment,
		Reference: "http://hl7.org/fhir/DSTU2/conformance.html",
		ImplGuide: "USCore 3.1",
	}

	kindRule := bv.KindValid(capStat)
	if !kindRule[0].Valid {
		ruleError.Comment = kindRule[0].Comment + " " + baseComment
		return ruleError
	}
	messaging, err := capStat.GetMessaging()
	if err != nil || len(messaging) == 0 {
		ruleError.Comment = "Messaging does not exist. " + baseComment
		return ruleError
	}
	for _, message := range messaging {
		endpoints, err := capStat.GetMessagingEndpoint(message)
		if err != nil || len(endpoints) == 0 {
			ruleError.Comment = "Endpoint field in Messaging does not exist. " + baseComment
			return ruleError
		}
	}

	ruleError.Valid = true
	ruleError.Actual = "true"
	return ruleError
}

func (bv *baseVal) EndpointFunctionValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	var actualVal []string
	baseComment := "A Conformance Resource SHALL have at least one of REST, messaging or document element."
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.EndptFunctionRule,
		Valid:     true,
		Expected:  "rest,messaging,document",
		Comment:   baseComment,
		Reference: "http://hl7.org/fhir/DSTU2/conformance.html",
		ImplGuide: "USCore 3.1",
	}

	if capStat == nil {
		ruleError.Valid = false
		ruleError.Comment = "The Conformance Resource does not exist; cannot check REST, messaging or document elements."
		return ruleError
	}

	// If rest is not nil, add to actual list
	rest, err := capStat.GetRest()
	if err == nil && len(rest) > 0 {
		actualVal = append(actualVal, "rest")
	}
	// If messaging is not nil, add to actual list
	messaging, err := capStat.GetMessaging()
	if err == nil && len(messaging) > 0 {
		actualVal = append(actualVal, "messaging")
	}
	// if document is not nil, add to actual list
	document, err := capStat.GetDocument()
	if err == nil && len(document) > 0 {
		actualVal = append(actualVal, "document")
	}
	// If none of the above exist, the capability statement is not valid
	if len(actualVal) == 0 {
		ruleError.Actual = ""
		ruleError.Valid = false
		return ruleError
	}
	ruleError.Actual = strings.Join(actualVal, ",")
	return ruleError
}

// DescribeEndpointValid checks the requirement: "A Conformance Resource/Capability Statement SHALL have at least one of description,
// software, or implementation element."
func (bv *baseVal) DescribeEndpointValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	var actualVal []string
	baseComment := "A Conformance Resource SHALL have at least one of description, software, or implementation element."
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.DescribeEndptRule,
		Valid:     true,
		Expected:  "description,software,implementation",
		Comment:   baseComment,
		Reference: "http://hl7.org/fhir/DSTU2/conformance.html",
		ImplGuide: "USCore 3.1",
	}

	if capStat == nil {
		ruleError.Valid = false
		ruleError.Comment = "The Conformance Resource does not exist; cannot check description, software, or implementation elements."
		return ruleError
	}

	// If description is not an empty string, add to actual list
	description, err := capStat.GetDescription()
	if err == nil && len(description) > 0 {
		actualVal = append(actualVal, "description")
	}
	// If software is not nil, add to actual list
	software, err := capStat.GetSoftware()
	if err == nil && len(software) > 0 {
		actualVal = append(actualVal, "software")
	}
	// if implementation is not nil, add to actual list
	implementation, err := capStat.GetImplementation()
	if err == nil && len(implementation) > 0 {
		actualVal = append(actualVal, "implementation")
	}
	// If none of the above exist, the capability statement is not valid
	if len(actualVal) == 0 {
		ruleError.Actual = ""
		ruleError.Valid = false
		return ruleError
	}
	ruleError.Actual = strings.Join(actualVal, ",")
	return ruleError
}

// DocumentSetValid checks the requirement: "The set of documents must be unique by the combination of profile and mode."
func (bv *baseVal) DocumentSetValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseComment := "The set of documents must be unique by the combination of profile and mode."
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.DocumentValidRule,
		Valid:     false,
		Expected:  "true",
		Actual:    "false",
		Comment:   baseComment,
		Reference: "http://hl7.org/fhir/DSTU2/conformance.html",
		ImplGuide: "USCore 3.1",
	}

	if capStat == nil {
		ruleError.Comment = "The Conformance Resource does not exist; cannot check documents."
		return ruleError
	}

	document, err := capStat.GetDocument()
	if err != nil {
		ruleError.Comment = "Document field is not formatted correctly. Cannot check if the set of documents are unique. " + baseComment
		return ruleError
	}
	if err == nil && len(document) == 0 {
		ruleError.Valid = true
		ruleError.Actual = "true"
		ruleError.Comment = "Document field does not exist, but is not required. " + baseComment
		return ruleError
	}
	var uniqueIDs []string
	invalid := false
	for _, doc := range document {
		mode := doc["mode"]
		if mode == nil {
			invalid = true
			break
		}
		modeStr, ok := mode.(string)
		if !ok {
			invalid = true
			break
		}
		profile := doc["profile"]
		if profile == nil {
			invalid = true
			break
		}
		profileStr, ok := profile.(string)
		if !ok {
			invalid = true
			break
		}
		// Combine profile & mode to compare against other defined documents
		id := profileStr + "." + modeStr
		if stringInList(id, uniqueIDs) {
			ruleError.Comment = "The set of documents are not unique. " + baseComment
			return ruleError
		}
		uniqueIDs = append(uniqueIDs, id)
	}
	if invalid {
		ruleError.Comment = "Document field is not formatted correctly. Cannot check if the set of documents are unique. " + baseComment
		return ruleError
	}
	ruleError.Valid = true
	ruleError.Actual = "true"
	return ruleError
}

func (bv *baseVal) UniqueResources(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseComment := "A given resource can only be described once per RESTful mode."
	returnVal := checkResourceList(capStat, endpointmanager.UniqueResourcesRule)
	returnVal.Comment = returnVal.Comment + baseComment
	returnVal.Reference = "http://hl7.org/fhir/DSTU2/conformance.html"
	return returnVal
}

func (bv *baseVal) SearchParamsUnique(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	var ruleError endpointmanager.Rule
	return ruleError
}

func (bv *baseVal) VersionResponseValid(fhirVersion string, defaultFhirVersion string) endpointmanager.Rule {
	var ruleError endpointmanager.Rule
	return ruleError
}
