package capabilityhandler

import (
	"strconv"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
)

var version3plus = []string{"3.0.0", "3.0.1", "4.0.0", "4.0.1"}
var fhir3PlusJSONMIMEType = "application/fhir+json"
var fhir2LessJSONMIMEType = "application/json+fhir"

// RuleOption is an enum of the names given to the rule validation checks
type RuleOption string

const (
	r4MimeTypeRule      RuleOption = "r4MimeType"
	generalMimeTypeRule RuleOption = "generalMimeType"
	httpResponseRule    RuleOption = "httpResponse"
)

// Rule is the structure for both validation errors and warnings that are saved in
// the Validations struct
type Rule struct {
	RuleName  RuleOption `json:"ruleName"`
	Expected  string     `json:"expected"`
	Comment   string     `json:"comment"`
	Reference string     `json:"reference"`
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// RunValidationChecks runs all of the validation checks based on the rule requirements from ONC
// The final Validation object is formatted as a map of strings to arrays of the rule struct
// return value: 	{
// 						"Errors":   []rule,
//						"Warnings": []rule,
// 					}
func RunValidationChecks(capStat capabilityparser.CapabilityStatement, httpResponse int, mimeTypes []string) map[string]interface{} {
	var validationErrors []Rule
	validationWarnings := make([]Rule, 0)

	returnedRule := r4MimeTypeValid(mimeTypes)
	if returnedRule != (Rule{}) {
		validationErrors = append(validationErrors, returnedRule)
	}

	if capStat != nil {
		fhirVersion, err := capStat.GetFHIRVersion()
		if err != nil {
			returnedRule = generalMimeTypeValid(mimeTypes, "")
		}
		returnedRule = generalMimeTypeValid(mimeTypes, fhirVersion)
	} else {
		returnedRule = generalMimeTypeValid(mimeTypes, "")
	}
	if returnedRule != (Rule{}) {
		validationErrors = append(validationErrors, returnedRule)
	}

	returnedRule = httpResponseValid(httpResponse)
	if returnedRule != (Rule{}) {
		validationErrors = append(validationErrors, returnedRule)
	}

	validations := map[string]interface{}{
		"Errors":   validationErrors,
		"Warnings": validationWarnings,
	}

	return validations
}

// r4MimeTypeValid checks to see if the application/fhir+json mime type was a valid mime type for this endpoint
func r4MimeTypeValid(mimeTypes []string) Rule {
	var ruleError Rule

	for _, mt := range mimeTypes {
		if mt == fhir3PlusJSONMIMEType {
			return ruleError
		}
	}

	ruleError.RuleName = r4MimeTypeRule
	ruleError.Expected = fhir3PlusJSONMIMEType
	ruleError.Comment = `The formal MIME-type for FHIR resources is application/fhir+json for FHIR
	version STU3 and above. The correct mime type SHALL be used by clients and servers.`
	ruleError.Reference = "http://hl7.org/fhir/http.html"
	return ruleError
}

// generalMimeTypeValid checks if the mime type is valid for the given fhirVersion.
// @TODO We might not care about this if endpoints are supposed to be version R4
func generalMimeTypeValid(mimeTypes []string, fhirVersion string) Rule {
	var ruleError Rule

	if len(fhirVersion) == 0 {
		ruleError.RuleName = generalMimeTypeRule
		ruleError.Expected = "N/A"
		ruleError.Comment = "Unknown FHIR Version; cannot validate mime type."
		return ruleError
	}

	var mimeError string
	for _, mt := range mimeTypes {
		if contains(version3plus, fhirVersion) {
			if mt == fhir3PlusJSONMIMEType {
				return ruleError
			}
			mimeError = fhir3PlusJSONMIMEType
		} else {
			// The fhirVersion has to be valid in order to create a valid capability statement
			// so if it's gotten this far, the fhirVersion has to be less than 3
			if mt == fhir2LessJSONMIMEType {
				return ruleError
			}
			mimeError = fhir2LessJSONMIMEType
		}
	}

	errorMsg := "FHIR Version " + fhirVersion + " requires the Mime Type to be " + mimeError

	ruleError.RuleName = generalMimeTypeRule
	ruleError.Expected = mimeError
	ruleError.Comment = errorMsg
	ruleError.Reference = "http://hl7.org/fhir/http.html"
	return ruleError
}

// httpReponseValid checks for the http response and returns
func httpResponseValid(httpResponse int) Rule {
	var ruleError Rule

	if httpResponse == 200 {
		return ruleError
	}

	s := strconv.Itoa(httpResponse)
	ruleError.RuleName = httpResponseRule
	ruleError.Expected = "200"
	ruleError.Reference = "http://hl7.org/fhir/http.html"
	ruleError.Comment = `The HTTP response code was ` + s + ` instead of 200. 
	Applications SHALL return a resource that describes the functionality of the server end-point.`

	if httpResponse == 0 {
		ruleError.Comment = `The GET request failed with no returned HTTP response status code.
		Applications SHALL return a resource that describes the functionality of the server end-point.`
	}

	return ruleError
}
