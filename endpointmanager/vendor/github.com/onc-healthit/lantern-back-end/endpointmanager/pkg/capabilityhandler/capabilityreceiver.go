package capabilityhandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/chplmapper"

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

	mimeType, ok := msgJSON["mimetype"].(string)
	if !ok {
		return nil, fmt.Errorf("%s: unable to cast MIME Type to string", url)
	}

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
		if fhirEndpoint.MimeType != "" {
			existingEndpt.MimeType = fhirEndpoint.MimeType
		}
		existingEndpt.Errors = fhirEndpoint.Errors
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
