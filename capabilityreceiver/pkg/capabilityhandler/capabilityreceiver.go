package capabilityhandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/onc-healthit/lantern-back-end/capabilityreceiver/pkg/capabilityhandler/validation"
	"github.com/onc-healthit/lantern-back-end/capabilityreceiver/pkg/chplmapper"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/pkg/errors"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	log "github.com/sirupsen/logrus"
)

func formatMessage(message []byte) (*endpointmanager.FHIREndpointInfo, error) {
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

	smarthttpResponseFloat, ok := msgJSON["smarthttpResponse"].(float64)
	if !ok {
		return nil, fmt.Errorf("unable to cast smart http response to int")
	}
	smarthttpResponse := int(smarthttpResponseFloat)

	var capStat capabilityparser.CapabilityStatement
	var capInt map[string]interface{}
	if msgJSON["capabilityStatement"] != nil {
		capInt, ok = msgJSON["capabilityStatement"].(map[string]interface{})

		if !ok {
			return nil, fmt.Errorf("%s: unable to cast capability statement to map[string]interface{}", url)
		}

		capStat, err = capabilityparser.NewCapabilityStatementFromInterface(capInt)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("%s: unable to parse CapabilityStatement out of message", url))
		}
	}
	var smartResponse capabilityparser.SMARTResponse
	if msgJSON["smartResp"] != nil {
		smartInt, ok := msgJSON["smartResp"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s: unable to cast smart response body to map[string]interface{}", url)
		}
		smartResponse = capabilityparser.NewSMARTRespFromInterface(smartInt)
	}

	responseTime, ok := msgJSON["responseTime"].(float64)
	if !ok {
		return nil, fmt.Errorf("Response time is not a float")
	}

	fhirVersion := ""
	if capStat != nil {
		fhirVersion, _ = capStat.GetFHIRVersion()
	}
	validator := validation.ValidatorForFHIRVersion(fhirVersion)

	validationObj := validator.RunValidation(capStat, httpResponse, mimeTypes, fhirVersion, tlsVersion, smarthttpResponse)
	includedFields := RunIncludedFieldsAndExtensionsChecks(capInt)
	supportedResources := RunSupportedResourcesChecks(capInt)

	FHIREndpointMetadata := &endpointmanager.FHIREndpointMetadata{
		URL:               url,
		HTTPResponse:      httpResponse,
		Errors:            errs,
		SMARTHTTPResponse: smarthttpResponse,
		ResponseTime:      responseTime,
	}

	fhirEndpoint := endpointmanager.FHIREndpointInfo{
		URL:                 url,
		TLSVersion:          tlsVersion,
		MIMETypes:           mimeTypes,
		Validation:          validationObj,
		CapabilityStatement: capStat,
		SMARTResponse:       smartResponse,
		IncludedFields:      includedFields,
		SupportedResources:  supportedResources,
		Metadata:            FHIREndpointMetadata,
	}

	return &fhirEndpoint, nil
}

// saveMsgInDB formats the message data for the database and either adds a new entry to the database or
// updates a current one
func saveMsgInDB(message []byte, args *map[string]interface{}) error {
	var err error
	var fhirEndpoint *endpointmanager.FHIREndpointInfo
	var existingEndpt *endpointmanager.FHIREndpointInfo

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

	existingEndpt, err = store.GetFHIREndpointInfoUsingURL(ctx, fhirEndpoint.URL)

	if err == sql.ErrNoRows {

		// If the endpoint info entry doesn't exist, add it to the DB
		err = chplmapper.MatchEndpointToVendor(ctx, fhirEndpoint, store)
		if err != nil {
			return err
		}
		err = chplmapper.MatchEndpointToProduct(ctx, fhirEndpoint, store, fmt.Sprintf("%v", (*args)["chplMatchFile"]))
		if err != nil {
			return err
		}
		err = store.AddFHIREndpointInfo(ctx, fhirEndpoint)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {

		fhirEndpoint.VendorID = existingEndpt.VendorID
		fhirEndpoint.HealthITProductID = existingEndpt.HealthITProductID

		existingEndpt.Metadata.URL = fhirEndpoint.Metadata.URL
		existingEndpt.Metadata.HTTPResponse = fhirEndpoint.Metadata.HTTPResponse
		existingEndpt.Metadata.Errors = fhirEndpoint.Metadata.Errors
		existingEndpt.Metadata.ResponseTime = fhirEndpoint.Metadata.ResponseTime
		existingEndpt.Metadata.SMARTHTTPResponse = fhirEndpoint.Metadata.SMARTHTTPResponse

		// If the existing endpoint info does not equal the stored endpoint info, update it with the new information, otherwise only update metadata.
		if !existingEndpt.EqualExcludeMetadata(fhirEndpoint) {
			existingEndpt.CapabilityStatement = fhirEndpoint.CapabilityStatement
			existingEndpt.TLSVersion = fhirEndpoint.TLSVersion
			existingEndpt.MIMETypes = fhirEndpoint.MIMETypes
			existingEndpt.Validation = fhirEndpoint.Validation
			existingEndpt.SMARTResponse = fhirEndpoint.SMARTResponse
			existingEndpt.IncludedFields = fhirEndpoint.IncludedFields
			existingEndpt.SupportedResources = fhirEndpoint.SupportedResources

			err = store.UpdateFHIREndpointInfo(ctx, existingEndpt)
			if err != nil {
				return err
			}
		} else {
			metadataID, err := store.AddFHIREndpointMetadata(ctx, existingEndpt.Metadata)
			if err != nil {
				return err
			}

			err = store.UpdateMetadataIDInfo(ctx, metadataID, existingEndpt.URL)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
	args["chplMatchFile"] = "/etc/lantern/resources/CHPLProductMapping.json"

	messages, err := messageQueue.ConsumeFromQueue(channelID, qName)
	if err != nil {
		return err
	}

	errs := make(chan error)
	go messageQueue.ProcessMessages(ctx, messages, saveMsgInDB, &args, errs)

	for elem := range errs {
		log.Warn(elem)
	}

	return nil
}
