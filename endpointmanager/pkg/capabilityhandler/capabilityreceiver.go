package capabilityhandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path"

	"github.com/onc-healthit/lantern-back-end/lanternmq"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	log "github.com/sirupsen/logrus"
)

func formatMessage(message []byte) (*endpointmanager.FHIREndpoint, error) {
	var msgJSON map[string]interface{}

	err := json.Unmarshal(message, &msgJSON)
	if err != nil {
		return nil, err
	}

	errs, ok := msgJSON["err"].(string)
	if !ok {
		return nil, fmt.Errorf("unable to cast message Error to string")
	}

	url, ok := msgJSON["url"].(string)
	if !ok {
		return nil, fmt.Errorf("unable to cast message URL to string")
	}

	tlsVersion, ok := msgJSON["tlsVersion"].(string)
	if !ok {
		return nil, fmt.Errorf("unable to cast TLS Version to string")
	}

	mimeType, ok := msgJSON["mimetype"].(string)
	if !ok {
		return nil, fmt.Errorf("unable to cast MIME Type to string")
	}

	// remove "metadata" from the url
	originalURL, file := path.Split(url)
	if file != "metadata" {
		originalURL = url
	}

	capJson, err := json.Marshal(msgJSON["capabilityStatement"].(string))
	if err != nil {
		return nil, fmt.Errorf("unable to marshal CapabilityStatement JSON")
	}

	capStat, err := capabilityparser.NewCapabilityStatement(capJson)
	if err != nil {
		return nil, fmt.Errorf("unable to parse CapabailtyStatement out of message"+err.Error())
	}

	fhirEndpoint := endpointmanager.FHIREndpoint{
		URL:                 originalURL,
		TLSVersion:          tlsVersion,
		MimeType:            mimeType,
		Errors:              errs,
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

	store := (*args)["store"].(endpointmanager.FHIREndpointStore)
	ctx := (*args)["ctx"].(context.Context)

	existingEndpt, err = store.GetFHIREndpointUsingURL(ctx, fhirEndpoint.URL)

	// If the URL doesn't exist, add it to the DB
	if err == sql.ErrNoRows {
		err = store.AddFHIREndpoint(ctx, fhirEndpoint)
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
		if fhirEndpoint.MimeType != "" {
			existingEndpt.MimeType = fhirEndpoint.MimeType
		}
		existingEndpt.Errors = fhirEndpoint.Errors
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
	store endpointmanager.FHIREndpointStore,
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
