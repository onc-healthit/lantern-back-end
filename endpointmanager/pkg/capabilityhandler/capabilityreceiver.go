package capabilityhandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/onc-healthit/lantern-back-end/capabilityquerier/pkg/capabilityquerier"

	"github.com/onc-healthit/lantern-back-end/lanternmq"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	log "github.com/sirupsen/logrus"
)

func formatMessage(message []byte) (*endpointmanager.FHIREndpoint, error) {
	var msgJSON capabilityquerier.Message

	err := json.Unmarshal(message, &(msgJSON))
	if err != nil {
		return nil, err
	}

	originalURL := strings.Replace(msgJSON.URL, "metadata", "", 1)

	fhirEndpoint := endpointmanager.FHIREndpoint{
		URL:                 originalURL,
		TLSVersion:          msgJSON.TLSVersion,
		MimeType:            msgJSON.MimeType,
		CapabilityStatement: msgJSON.CapabilityStatement,
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

	ctx := context.Background()
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
		// Add the new information and update the endpoint in the database
		existingEndpt.CapabilityStatement = fhirEndpoint.CapabilityStatement
		existingEndpt.TLSVersion = fhirEndpoint.TLSVersion
		existingEndpt.MimeType = fhirEndpoint.MimeType
		err = store.UpdateFHIREndpoint(ctx, existingEndpt)
		if err != nil {
			return err
		}
	}

	return err
}

// CapabilityReceiver receives the capability statement from the queue and adds it to the database
func CapabilityReceiver(store endpointmanager.FHIREndpointStore,
	messageQueue lanternmq.MessageQueue,
	channelID lanternmq.ChannelID,
	qName string) error {

	args := make(map[string]interface{})
	args["store"] = store

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
