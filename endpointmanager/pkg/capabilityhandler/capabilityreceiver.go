package capabilityhandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/chplmapper"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/pkg/errors"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	log "github.com/sirupsen/logrus"
)

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

	var mimeTypeValidObj validationError
	if msgJSON["capabilityStatement"] != nil {
		fhirVersion, err := capStat.GetFHIRVersion()
		if err != nil {
			return nil, err
		}
		mimeTypeValidObj = mimeTypeValid(mimeTypes, fhirVersion)
	} else {
		mimeTypeValidObj = mimeTypeValid(mimeTypes, "")
	}

	httpCodeObj := httpResponseValid(httpResponse)

	validationObj := map[string]interface{}{
		"mimeType": mimeTypeValidObj,
		"httpCode": httpCodeObj,
	}

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

	store, ok := (*args)["store"].(*postgresql.Store)
	if !ok {
		return fmt.Errorf("unable to cast postgresql store from arguments")
	}
	ctx, ok := (*args)["ctx"].(context.Context)
	if !ok {
		return fmt.Errorf("unable to cast context from arguments")
	}

	existingEndpt, err = store.GetFHIREndpointUsingURL(ctx, fhirEndpoint.URL)

	// If the URL doesn't exist, add it to the DB
	if err == sql.ErrNoRows {
		err = chplmapper.MatchEndpointToVendorAndProduct(ctx, fhirEndpoint, store)
		if err != nil {
			return err
		}
		err = store.AddFHIREndpoint(ctx, fhirEndpoint)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// Add the new information if it's valid and update the endpoint in the database
		existingEndpt.CapabilityStatement = fhirEndpoint.CapabilityStatement
		existingEndpt.TLSVersion = fhirEndpoint.TLSVersion
		existingEndpt.MIMETypes = fhirEndpoint.MIMETypes
		existingEndpt.HTTPResponse = fhirEndpoint.HTTPResponse
		existingEndpt.Errors = fhirEndpoint.Errors
		existingEndpt.Validation = fhirEndpoint.Validation
		err = chplmapper.MatchEndpointToVendorAndProduct(ctx, existingEndpt, store)
		if err != nil {
			return err
		}
		err = store.UpdateFHIREndpoint(ctx, existingEndpt)
		if err != nil {
			return err
		}
	}

	return err
}

// ReceiveCapabilityStatements connects to the given message queue channel and receives the capability
// statements from it. It then adds the capability statements to the given store.
func ReceiveCapabilityStatements(ctx context.Context,
	store *postgresql.Store,
	messageQueue lanternmq.MessageQueue,
	channelID lanternmq.ChannelID,
	qName string) error {

	args := make(map[string]interface{})
	args["store"] = store
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
