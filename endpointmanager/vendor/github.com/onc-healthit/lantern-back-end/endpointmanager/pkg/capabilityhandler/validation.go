package capabilityhandler

import "strconv"

var version3plus = []string{"3.0.0", "3.0.1", "4.0.0", "4.0.1"}
var fhir3PlusJSONMIMEType = "application/fhir+json"
var fhir2LessJSONMIMEType = "application/json+fhir"

// validationError is the structure for validation errors that are saved in the Validation
// JSON blob in fhir_endpoints for now.
type validationError struct {
	Correct  bool   `json:"correct"`
	Expected string `json:"expected"`
	Comment  string `json:"comment"`
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// This function takes in the array of accepted Mime Types by a specific endpoint and that endpoint's FHIR
// version. It returns whether this is an error or warning, whether it's correct, the expected value,
// and a comment on the result if necessary.
func mimeTypeValid(mimeTypes []string, fhirVersion string) validationError {
	if len(fhirVersion) == 0 {
		return validationError{
			Correct:  false,
			Expected: "",
			Comment:  "Cannot compare FHIR Version and Mime Type",
		}
	}

	var mimeError string
	for _, mimeType := range mimeTypes {
		if contains(version3plus, fhirVersion) {
			if mimeType == fhir3PlusJSONMIMEType {
				return validationError{
					Correct:  true,
					Expected: fhir3PlusJSONMIMEType,
					Comment:  "",
				}
			}
			mimeError = fhir3PlusJSONMIMEType
		} else {
			// The fhirVersion has to be valid in order to create a valid capability statement
			// so if it's gotten this far, the fhirVersion has to be less than 3
			if mimeType == fhir2LessJSONMIMEType {
				return validationError{
					Correct:  true,
					Expected: fhir2LessJSONMIMEType,
					Comment:  "",
				}
			}
			mimeError = fhir2LessJSONMIMEType
		}
	}

	errorMsg := "FHIR Version " + fhirVersion + " requires the Mime Type to be " + mimeError

	return validationError{
		Correct:  false,
		Expected: mimeError,
		Comment:  errorMsg,
	}
}

func httpResponseValid(httpResponse int) validationError {
	if httpResponse == 200 {
		return validationError{
			Correct:  true,
			Expected: "200",
			Comment:  "",
		}
	} else if httpResponse == 0 {
		return validationError{
			Correct:  false,
			Expected: "200",
			Comment:  "The GET request failed",
		}
	}
	s := strconv.Itoa(httpResponse)
	return validationError{
		Correct:  false,
		Expected: "200",
		Comment:  "The HTTP response code was " + s + " instead of 200",
	}
}
