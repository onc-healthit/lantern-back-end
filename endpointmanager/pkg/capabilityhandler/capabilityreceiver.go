package capabilityhandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path"
	"strconv"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/chplmapper"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/pkg/errors"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	log "github.com/sirupsen/logrus"
)

// var version2minus = []string{"1.0.1", "1.0.2"}
var version3plus = []string{"3.0.0", "3.0.1", "4.0.0", "4.0.1"}
var fhir3PlusJSONMIMEType = "application/fhir+json"
var fhir2LessJSONMIMEType = "application/json+fhir"

// ValidationError is the structure for validation errors that are saved in the Validation JSON
// blob in fhir_endpoints for now.
// @TODO This will be moved
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

func formatMessage(message []byte) (*endpointmanager.FHIREndpoint, error) {
	var msgJSON map[string]interface{}

	err := json.Unmarshal(message, &msgJSON)
	if err != nil {
		return nil, err
	}

	url, ok := msgJSON["url"].(string)
	if !ok {
		return nil, fmt.Errorf("unable to cast message URL to string")
	}

	errs, ok := msgJSON["err"].(string)
	if !ok {
		return nil, fmt.Errorf("%s: unable to cast message Error to string", url)
	}

	tlsVersion, ok := msgJSON["tlsVersion"].(string)
	if !ok {
		return nil, fmt.Errorf("%s: unable to cast TLS Version to string", url)
	}

	// TODO: for some reason casting to []string doesn't work... need to do roundabout way
	// Could be investigated further
	var mimeTypes []string
	if msgJSON["mimeTypes"] != nil {
		mimeTypesInt, ok := msgJSON["mimeTypes"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("%s: unable to cast MIME Types to []interface{}", url)
		}
		for _, mimeTypeInt := range mimeTypesInt {
			mimeType, ok := mimeTypeInt.(string)
			if !ok {
				return nil, fmt.Errorf("unable to cast mime type to string")
			}
			mimeTypes = append(mimeTypes, mimeType)
		}
	}

	// JSON numbers are golang float64s
	httpResponseFloat, ok := msgJSON["httpResponse"].(float64)
	if !ok {
		return nil, fmt.Errorf("unable to cast http response to int")
	}
	httpResponse := int(httpResponseFloat)

	// remove "metadata" from the url
	originalURL, file := path.Split(url)
	if file != "metadata" {
		originalURL = url
	}

	var capStat capabilityparser.CapabilityStatement
	if msgJSON["capabilityStatement"] != nil {
		capInt, ok := msgJSON["capabilityStatement"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s: unable to cast capability statement to map[string]interface{}", url)
		}
		capStat, err = capabilityparser.NewCapabilityStatementFromInterface(capInt)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("%s: unable to parse CapabilityStatement out of message", url))
		}
	}

	/**
	Quick Validation
	*/

	// @TODO Check if MimeType is valid
	var mimeTypeValidObj validationError
	if msgJSON["capabilityStatement"] != nil {
		fhirVersion, err := capStat.GetFHIRVersion()
		if err != nil {
			return nil, err
		}
		mimeTypeValidObj = mimeTypeValid(mimeTypes, fhirVersion)
	} else {
		mimeTypeValidObj = validationError{
			Correct:  false,
			Expected: "",
			Comment:  "No Fhir Version to validate if the Mime Type is accurate",
		}
	}

	// @TODO Update once we have actual information
	httpCodeObj := httpResponseValid(httpResponse)
	// httpCodeObj := httpResponseValid(httpResponse)
	validationObj := map[string]interface{}{
		"mimeType": mimeTypeValidObj,
		"httpCode": httpCodeObj,
	}

	// @TODO Check if TLSVersion is valid

	fhirEndpoint := endpointmanager.FHIREndpoint{
		URL:          originalURL,
		TLSVersion:   tlsVersion,
		MIMETypes:    mimeTypes,
		HTTPResponse: httpResponse,
		Errors:       errs,
		Validation: map[string]interface{}{
			"errors": validationObj,
		},
		CapabilityStatement: capStat,
	}

	return &fhirEndpoint, nil
}

// This function takes in the array of accepted Mime Types by a specific endpoint and that endpoint's FHIR
// version. It returns whether this is an error or warning, whether it's correct, the expected value,
// and a comment on the result if necessary.
func mimeTypeValid(mimeTypes []string, fhirVersion string) validationError {
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

// saveMsgInDB formats the message data for the database and either adds a new entry to the database or
// updates a current one
func saveMsgInDB(message []byte, args *map[string]interface{}) error {
	var err error
	var fhirEndpoint *endpointmanager.FHIREndpoint
	var existingEndpt *endpointmanager.FHIREndpoint

	fhirEndpoint, err = formatMessage(message)
	if err != nil {
		return err
	}

	hitpStore, ok := (*args)["hitpStore"].(endpointmanager.HealthITProductStore)
	if !ok {
		return fmt.Errorf("unable to cast health it store from arguments")
	}
	epStore, ok := (*args)["epStore"].(endpointmanager.FHIREndpointStore)
	if !ok {
		return fmt.Errorf("unable to cast fhir endpoint store from argument")
	}
	ctx, ok := (*args)["ctx"].(context.Context)
	if !ok {
		return fmt.Errorf("unable to cast context from arguments")
	}

	existingEndpt, err = epStore.GetFHIREndpointUsingURL(ctx, fhirEndpoint.URL)

	// If the URL doesn't exist, add it to the DB
	if err == sql.ErrNoRows {
		err = chplmapper.MatchEndpointToVendorAndProduct(ctx, fhirEndpoint, hitpStore)
		if err != nil {
			return err
		}
		err = epStore.AddFHIREndpoint(ctx, fhirEndpoint)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// Add the new information if it's valid and update the endpoint in the database
		if fhirEndpoint.CapabilityStatement != nil {
			existingEndpt.CapabilityStatement = fhirEndpoint.CapabilityStatement
		}
		if fhirEndpoint.TLSVersion != "" {
			existingEndpt.TLSVersion = fhirEndpoint.TLSVersion
		}
		if fhirEndpoint.MIMETypes != nil {
			existingEndpt.MIMETypes = fhirEndpoint.MIMETypes
		}
		if fhirEndpoint.HTTPResponse != 0 {
			existingEndpt.HTTPResponse = fhirEndpoint.HTTPResponse
		}
		existingEndpt.Errors = fhirEndpoint.Errors
		existingEndpt.Validation = fhirEndpoint.Validation
		err = chplmapper.MatchEndpointToVendorAndProduct(ctx, existingEndpt, hitpStore)
		if err != nil {
			return err
		}
		err = epStore.UpdateFHIREndpoint(ctx, existingEndpt)
		if err != nil {
			return err
		}
	}

	return err
}

// ReceiveCapabilityStatements connects to the given message queue channel and receives the capability
// statements from it. It then adds the capability statements to the given store.
func ReceiveCapabilityStatements(ctx context.Context,
	epStore endpointmanager.FHIREndpointStore,
	hitpStore endpointmanager.HealthITProductStore,
	messageQueue lanternmq.MessageQueue,
	channelID lanternmq.ChannelID,
	qName string) error {

	args := make(map[string]interface{})
	args["hitpStore"] = hitpStore
	args["epStore"] = epStore
	args["ctx"] = ctx

	messages, err := messageQueue.ConsumeFromQueue(channelID, qName)
	if err != nil {
		return err
	}

	errs := make(chan error)
	go messageQueue.ProcessMessages(messages, saveMsgInDB, &args, errs)

	for elem := range errs {
		log.Warn(elem)
	}

	return nil
}
