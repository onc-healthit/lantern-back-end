package capabilityhandler

import (
	"strconv"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

var version3plus = []string{"3.0.0", "3.0.1", "4.0.0", "4.0.1"}
var fhir3PlusJSONMIMEType = "application/fhir+json"
var fhir2LessJSONMIMEType = "application/json+fhir"

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// RunValidationChecks runs all of the validation checks based on the rule requirements from ONC
func RunValidationChecks(capStat capabilityparser.CapabilityStatement, httpResponse int, mimeTypes []string) endpointmanager.Validation {
	var validationResults []endpointmanager.Rule
	validationWarnings := make([]endpointmanager.Rule, 0)

	returnedRule := r4MimeTypeValid(mimeTypes)
	validationResults = append(validationResults, returnedRule)

	if capStat != nil {
		fhirVersion, err := capStat.GetFHIRVersion()
		if err != nil {
			returnedRule = generalMimeTypeValid(mimeTypes, "")
		} else {
			returnedRule = generalMimeTypeValid(mimeTypes, fhirVersion)
		}
	} else {
		returnedRule = generalMimeTypeValid(mimeTypes, "")
	}
	validationResults = append(validationResults, returnedRule)

	returnedRule = httpResponseValid(httpResponse)
	validationResults = append(validationResults, returnedRule)

	validations := endpointmanager.Validation{
		Results:  validationResults,
		Warnings: validationWarnings,
	}

	return validations
}

// r4MimeTypeValid checks to see if the R4 required mime type application/fhir+json was a valid mime type for this endpoint
func r4MimeTypeValid(mimeTypes []string) endpointmanager.Rule {
	mimeString := strings.Join(mimeTypes, ", ")
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.R4MimeTypeRule,
		Valid:     true,
		Expected:  fhir3PlusJSONMIMEType,
		Actual:    mimeString,
		Comment:   "The formal MIME-type for FHIR resources is application/fhir+json for FHIR version STU3 and above. The correct mime type SHALL be used by clients and servers.",
		Reference: "http://hl7.org/fhir/http.html",
		ImplGuide: "USCore 3.1",
	}

	for _, mt := range mimeTypes {
		if mt == fhir3PlusJSONMIMEType {
			return ruleError
		}
	}

	ruleError.Valid = false
	return ruleError
}

// generalMimeTypeValid checks if the mime type is valid for the given fhirVersion.
// @TODO We might not care about this if endpoints are supposed to be version R4
func generalMimeTypeValid(mimeTypes []string, fhirVersion string) endpointmanager.Rule {
	mimeString := strings.Join(mimeTypes, ",")
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.GeneralMimeTypeRule,
		Valid:     true,
		Expected:  "",
		Actual:    mimeString,
		Comment:   "",
		Reference: "http://hl7.org/fhir/http.html",
		ImplGuide: "USCore 3.1",
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

// httpReponseValid checks for the http response and returns
func httpResponseValid(httpResponse int) endpointmanager.Rule {
	strResp := strconv.Itoa(httpResponse)
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.HTTPResponseRule,
		Valid:     true,
		Expected:  "200",
		Actual:    strResp,
		Comment:   "",
		Reference: "http://hl7.org/fhir/http.html",
		ImplGuide: "USCore 3.1",
	}

	if httpResponse == 200 {
		return ruleError
	}

	ruleError.Comment = "The HTTP response code was " + strResp + " instead of 200. Applications SHALL return a resource that describes the functionality of the server end-point."

	if httpResponse == 0 {
		ruleError.Comment = "The GET request failed with no returned HTTP response status code. Applications SHALL return a resource that describes the functionality of the server end-point."
	}

	ruleError.Valid = false
	return ruleError
}
