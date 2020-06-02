package capabilityhandler

import "strconv"

var version3plus = []string{"3.0.0", "3.0.1", "4.0.0", "4.0.1"}
var fhir3PlusJSONMIMEType = "application/fhir+json"
var fhir2LessJSONMIMEType = "application/json+fhir"

type RuleOption string

const (
	r4MimeType  RuleOption = "r4MimeType"
	generalMimeType RuleOption = "generalMimeType"
	httpResponse RuleOption = "httpResponse"
)

// Validations holds all of the errors and warnings from running the validation checks
// it is saved in JSON format to the fhir_endpoints_info database table
type Validations struct {
	Errors rule[] `json:"errors"`
	Warnings rule[] `json:"warnings"`
}

// rule is the structure for both validation errors and warnings that are saved in
// the Validations struct
type rule struct {
	ruleName RuleOption `json:"ruleName"`
	expected string `json:"expected"`
	comment  string `json:"comment"`
	reference string `json:"reference"`
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func RunValidationChecks(capStat capabilityparser.CapabilityStatement, httpResponse int, mimeTypes []string) Validations {
	var validations Validations

	returnedRule := r4MimeTypeValid(mimeTypes)
	if returnedRule != nil {
		validations.Errors = append(validations.Errors, returnedRule)
	}

	// @TODO This might not work, will have to test first
	if capStat != nil {
		fhirVersion, err := capStat.GetFHIRVersion()
		if err != nil {
			return nil, err
		}
		returnedRule = generalMimeTypeValid(mimeTypes, fhirVersion)
	} else {
		returnedRule = generalMimeTypeValid(mimeTypes, "")
	}
	if returnedRule != nil {
		validations.Errors = append(validations.Errors, returnedRule)
	}

	returnedRule = httpResponseValid(httpResponse)
	if returnedRule != nil {
		validations.Errors = append(validations.Errors, returnedRule)
	}

	return Validations
}

// r4MimeTypeValid checks to see if the application/fhir+json mime type was a valid mime type for this endpoint
func r4MimeTypeValid(mimeTypes []string) rule {
	for _, mt := range mimeTypes {
		if mt == fhir3PlusJSONMIMEType {
			return nil
		}
	}

	mimeTypeComment := `The formal MIME-type for FHIR resources is application/fhir+json for FHIR
		version STU3 and above. The correct mime type SHALL be used by clients and servers.`
	return rule{
		ruleName: r4MimeType,
		expected: fhir3PlusJSONMIMEType,
		comment: mimeTypeComment,
		reference: "http://hl7.org/fhir/http.html",
	}
}

// generalMimeTypeValid checks if the mime type is valid for the given fhirVersion.
// @TODO We might not care about this if endpoints are supposed to be version R4
func generalMimeTypeValid(mimeTypes []string, fhirVersion string) rule {
	if len(fhirVersion) == 0 {
		return rule{
			ruleName: generalMimeType,
			expected: "N/A",
			comment: "Unknown FHIR Version; cannot validate mime type.",
		}
	}

	var mimeError string
	for _, mt := range mimeTypes {
		if contains(version3plus, fhirVersion) {
			if mt == fhir3PlusJSONMIMEType {
				return nil
			}
			mimeError = fhir3PlusJSONMIMEType
		} else {
			// The fhirVersion has to be valid in order to create a valid capability statement
			// so if it's gotten this far, the fhirVersion has to be less than 3
			if mt == fhir2LessJSONMIMEType {
				return nil
			}
			mimeError = fhir2LessJSONMIMEType
		}
	}

	errorMsg := "FHIR Version " + fhirVersion + " requires the Mime Type to be " + mimeError

	return rule{
		ruleName: generalMimeType,
		expected: mimeError,
		comment: errorMsg,
		reference: "http://hl7.org/fhir/http.html",
	}
}

func httpResponseValid(httpResponse int) rule {
	if httpResponse == 200 {
		return nil
	} else if httpResponse == 0 {
		httpComment := `The GET request failed with no returned HTTP response status code.
			Applications SHALL return a resource that describes the functionality of the server end-point.`
		return rule{
			ruleName: httpResponse,
			expected: "200",
			comment: httpComment,
			reference: "http://hl7.org/fhir/http.html",
		}
	}
	s := strconv.Itoa(httpResponse)
	statusComment := `The HTTP response code was ` + s + ` instead of 200.
			Applications SHALL return a resource that describes the functionality of the server end-point.`
	return validationError{
		ruleName: httpResponse,
		expected: "200",
		comment: statusComment,
		reference: "http://hl7.org/fhir/http.html",
	}
}
