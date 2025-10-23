package capabilityhandler

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	"github.com/spf13/viper"

	"github.com/onc-healthit/lantern-back-end/capabilityreceiver/pkg/capabilityhandler/validation"
	"github.com/onc-healthit/lantern-back-end/capabilityreceiver/pkg/chplmapper"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/versionsoperatorparser"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/pkg/errors"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/smartparser"
	log "github.com/sirupsen/logrus"
)

// versionsQueryArgs is a struct to hold the args that will be consumed by the
// saveVersionResponseMsgInDB function
type versionsQueryArgs struct {
	store             *postgresql.Store
	ctx               context.Context
	capQueryChannelID lanternmq.ChannelID
	capQueryQueue     lanternmq.MessageQueue
}

// capStatQueryArgs is a struct to hold the args that will be consumed by the
// saveMsgInDB function
type capStatQueryArgs struct {
	store                    *postgresql.Store
	ctx                      context.Context
	chplMatchFile            string
	chplEndpointListInfoFile string
}

func formatMessage(message []byte) (*endpointmanager.FHIREndpointInfo, *endpointmanager.Validation, error) {
	var msgJSON map[string]interface{}

	err := json.Unmarshal(message, &msgJSON)
	if err != nil {
		return nil, nil, err
	}

	url, ok := msgJSON["url"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("unable to cast message URL to string")
	}

	errs, ok := msgJSON["err"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("%s: unable to cast message Error to string", url)
	}

	tlsVersion, ok := msgJSON["tlsVersion"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("%s: unable to cast TLS Version to string", url)
	}

	requestedFhirVersion, ok := msgJSON["requestedFhirVersion"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("%s: unable to cast Requested Fhir Version to string", url)
	}

	defaultFhirVersion, ok := msgJSON["defaultFhirVersion"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("%s: unable to cast Default Fhir Version to string", url)
	}

	var mimeTypes []string
	if msgJSON["mimeTypes"] != nil {
		mimeTypesInt, ok := msgJSON["mimeTypes"].([]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("%s: unable to cast MIME Types to []interface{}", url)
		}
		for _, mimeTypeInt := range mimeTypesInt {
			mimeType, ok := mimeTypeInt.(string)
			if !ok {
				return nil, nil, fmt.Errorf("unable to cast mime type to string")
			}
			mimeTypes = append(mimeTypes, mimeType)
		}
	}

	// JSON numbers are golang float64s
	httpResponseFloat, ok := msgJSON["httpResponse"].(float64)
	if !ok {
		return nil, nil, fmt.Errorf("unable to cast http response to int")
	}
	httpResponse := int(httpResponseFloat)

	smarthttpResponseFloat, ok := msgJSON["smarthttpResponse"].(float64)
	if !ok {
		return nil, nil, fmt.Errorf("unable to cast smart http response to int")
	}
	smarthttpResponse := int(smarthttpResponseFloat)

	var capStat capabilityparser.CapabilityStatement
	var capInt map[string]interface{}
	if msgJSON["capabilityStatement"] != nil {
		capInt, ok = msgJSON["capabilityStatement"].(map[string]interface{})

		if !ok {
			return nil, nil, fmt.Errorf("%s: unable to cast capability statement to map[string]interface{}", url)
		}

		capStat, err = capabilityparser.NewCapabilityStatementFromInterface(capInt)
		if err != nil {
			return nil, nil, errors.Wrap(err, fmt.Sprintf("%s: unable to parse CapabilityStatement out of message", url))
		}
	}

	var capStatBytes []byte
	if msgJSON["capabilityStatementBytes"] != nil {
		capStatStringBytes, ok := msgJSON["capabilityStatementBytes"].(string)
		if !ok {
			return nil, nil, fmt.Errorf("unable to cast capStatBytes to string")
		}

		rawDecodedCapStat, err := base64.StdEncoding.DecodeString(capStatStringBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to cast capStatBytes to decoded string")
		}

		capStatBytes = []byte(rawDecodedCapStat)
	}

	var smartResponseBytes []byte
	if msgJSON["smartRespBytes"] != nil {
		smartRespStringBytes, ok := msgJSON["smartRespBytes"].(string)
		if !ok {
			return nil, nil, fmt.Errorf("unable to cast smartRespBytes to string")
		}

		rawDecodedSmartResp, err := base64.StdEncoding.DecodeString(smartRespStringBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to cast smartRespBytes to decoded string")
		}

		smartResponseBytes = []byte(rawDecodedSmartResp)
	}

	var smartResponse smartparser.SMARTResponse
	if msgJSON["smartResp"] != nil {
		smartInt, ok := msgJSON["smartResp"].(map[string]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("%s: unable to cast smart response body to map[string]interface{}", url)
		}
		smartResponse = smartparser.NewSMARTRespFromInterface(smartInt)
	}

	responseTime, ok := msgJSON["responseTime"].(float64)
	if !ok {
		return nil, nil, fmt.Errorf("response time is not a float")
	}

	fhirVersion := ""
	if capStat != nil {
		fhirVersion, _ = capStat.GetFHIRVersion()
	}

	validator := validation.ValidatorForFHIRVersion(fhirVersion)

	validationObj := validator.RunValidation(capStat, fhirVersion, tlsVersion, smartResponse, requestedFhirVersion, defaultFhirVersion)
	includedFields := RunIncludedFieldsAndExtensionsChecks(capInt, fhirVersion)
	operationResource := RunSupportedResourcesChecks(capInt)
	supportedProfiles := RunSupportedProfilesCheck(capInt, fhirVersion)

	FHIREndpointMetadata := &endpointmanager.FHIREndpointMetadata{
		URL:                  url,
		HTTPResponse:         httpResponse,
		Errors:               errs,
		SMARTHTTPResponse:    smarthttpResponse,
		ResponseTime:         responseTime,
		RequestedFhirVersion: requestedFhirVersion,
	}

	fhirEndpoint := endpointmanager.FHIREndpointInfo{
		URL:                      url,
		TLSVersion:               tlsVersion,
		MIMETypes:                mimeTypes,
		CapabilityStatement:      capStat,
		SMARTResponse:            smartResponse,
		IncludedFields:           includedFields,
		OperationResource:        operationResource,
		Metadata:                 FHIREndpointMetadata,
		RequestedFhirVersion:     requestedFhirVersion,
		CapabilityFhirVersion:    fhirVersion,
		SupportedProfiles:        supportedProfiles,
		CapabilityStatementBytes: capStatBytes,
		SMARTResponseBytes:       smartResponseBytes,
	}

	return &fhirEndpoint, &validationObj, nil
}

// saveMsgInDB formats the message data for the database and either adds a new entry to the database or
// updates a current one
func saveMsgInDB(message []byte, args *map[string]interface{}) error {
	var err error
	var fhirEndpoint *endpointmanager.FHIREndpointInfo
	var existingEndpt *endpointmanager.FHIREndpointInfo
	var validation *endpointmanager.Validation

	// Get arguments
	qa, ok := (*args)["queryArgs"].(capStatQueryArgs)
	if !ok {
		return fmt.Errorf("unable to parse args into capStatQueryArgs")
	}

	fhirEndpoint, validation, err = formatMessage(message)
	if err != nil {
		return err
	}

	// This is a safety check to make sure the RequestedFhirVersion will always be populated
	if fhirEndpoint.RequestedFhirVersion == "" {
		fhirEndpoint.RequestedFhirVersion = "None"
	}

	store := qa.store
	ctx := qa.ctx

	softwareListMap, err := chplmapper.OpenCHPLEndpointListInfoFile(fmt.Sprintf("%v", qa.chplEndpointListInfoFile))
	if err != nil {
		return fmt.Errorf("Opening CHPL endpoint list info file failed, %s", err)
	}

	existingEndpt, err = store.GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx, fhirEndpoint.URL, fhirEndpoint.RequestedFhirVersion)

	log.Info("Inside saveMsgInDB - outer")

	if err == sql.ErrNoRows {

		log.Info("Inside sql.ErrNoRows")

		// If the endpoint info entry doesn't exist, add it to the DB
		metadataID, err := store.AddFHIREndpointMetadata(ctx, fhirEndpoint.Metadata)
		if err != nil {
			return fmt.Errorf("doesn't exist, add endpoint metadata failed, %s", err)
		}

		valResID, err := store.AddValidationResult(ctx)
		if err != nil {
			return fmt.Errorf("adding new validation result ID failed, %s", err)
		}
		fhirEndpoint.ValidationID = valResID

		err = store.AddValidation(ctx, validation, valResID)
		if err != nil {
			return fmt.Errorf("error adding validation rows to table, %s", err)
		}

		// Pull url, list_source pairs from the db (url will be the same, list_source values will differ)
		fhirEndpointList, err := store.GetFHIREndpointUsingURL(ctx, fhirEndpoint.URL)
		if err != nil {
			return errors.Wrap(err, "error getting fhir endpoints from DB")
		}

		log.Info("Inside saveMsgInDB")

		log.Info("fhirEndpoint.URL: ", fhirEndpoint.URL, "\n")

		// For each list_source
		for _, fhirEp := range fhirEndpointList {

			log.Infof("Processing list source: ", fhirEp.ListSource, "\n")

			log.Info("softwareListMap[fhirEp.ListSource]: ", softwareListMap[fhirEp.ListSource])

			developerNames := softwareListMap[fhirEp.ListSource].ChplDeveloper
			productIds := softwareListMap[fhirEp.ListSource].ChplProductIDs

			if len(developerNames) == 0 {
				err = store.AddFHIREndpointInfo(ctx, fhirEndpoint, metadataID)
				if err != nil {
					return fmt.Errorf("doesn't exist, add to fhir_endpoints_info failed, %s", err)
				}
			} else {
				// Track the developers that have already been processed
				isDeveloperSeen := make(map[string]bool)

				for _, developerName := range developerNames {

					log.Infof("Processing developer: ", developerName, "\n")

					if !isDeveloperSeen[developerName] {
						err = chplmapper.MatchEndpointToVendor(ctx, fhirEndpoint, store, developerName)
						if err != nil {
							return fmt.Errorf("doesn't exist, match endpoint to vendor failed, %s", err)
						}

						var productIdsPerDeveloper []string
						for idx, productId := range productIds {
							log.Info("Processing product ID: ", productId, "\n")
							if developerNames[idx] == developerName {
								log.Info("developerNames[idx]: ", developerNames[idx], "\n")
								productIdsPerDeveloper = append(productIdsPerDeveloper, productId)
							}
						}

						fhirEndpoint.HealthITProductID = 0 // Reset HealthITProductID to 0 before matching to product
						err = chplmapper.MatchEndpointToProduct(ctx, fhirEndpoint, store, fmt.Sprintf("%v", qa.chplMatchFile), productIdsPerDeveloper)
						if err != nil {
							return fmt.Errorf("doesn't exist, match endpoint to product failed, %s", err)
						}

						err = store.AddFHIREndpointInfo(ctx, fhirEndpoint, metadataID)
						if err != nil {
							return fmt.Errorf("doesn't exist, add to fhir_endpoints_info failed, %s", err)
						}

						isDeveloperSeen[developerName] = true
					}
				}
			}
		}
	} else if err != nil {
		return err
	} else {

		log.Info("Inside else")

		fhirEndpoint.VendorID = existingEndpt.VendorID
		fhirEndpoint.HealthITProductID = existingEndpt.HealthITProductID

		existingEndpt.Metadata.URL = fhirEndpoint.Metadata.URL
		existingEndpt.Metadata.HTTPResponse = fhirEndpoint.Metadata.HTTPResponse
		existingEndpt.Metadata.Errors = fhirEndpoint.Metadata.Errors
		existingEndpt.Metadata.ResponseTime = fhirEndpoint.Metadata.ResponseTime
		existingEndpt.Metadata.SMARTHTTPResponse = fhirEndpoint.Metadata.SMARTHTTPResponse
		existingEndpt.Metadata.RequestedFhirVersion = fhirEndpoint.Metadata.RequestedFhirVersion

		// Set fhirEndpoint.ValidationID to existingEndpt value because they should have the same ValidationID
		// until there's a reason to update it
		fhirEndpoint.ValidationID = existingEndpt.ValidationID

		// If the existing endpoint info does not equal the stored endpoint info, update it with the new information, otherwise only update metadata.
		if !existingEndpt.EqualExcludeMetadata(fhirEndpoint) {
			existingEndpt.CapabilityStatement = fhirEndpoint.CapabilityStatement
			existingEndpt.CapabilityStatementBytes = fhirEndpoint.CapabilityStatementBytes
			existingEndpt.SMARTResponseBytes = fhirEndpoint.SMARTResponseBytes
			existingEndpt.TLSVersion = fhirEndpoint.TLSVersion
			existingEndpt.MIMETypes = fhirEndpoint.MIMETypes
			existingEndpt.SMARTResponse = fhirEndpoint.SMARTResponse
			existingEndpt.IncludedFields = fhirEndpoint.IncludedFields
			existingEndpt.OperationResource = fhirEndpoint.OperationResource
			existingEndpt.SupportedProfiles = fhirEndpoint.SupportedProfiles
			existingEndpt.CapabilityFhirVersion = fhirEndpoint.CapabilityFhirVersion

			log.Info("Updating other fields in existing endpoints")

			metadataID, err := store.AddFHIREndpointMetadata(ctx, existingEndpt.Metadata)
			if err != nil {
				return fmt.Errorf("does exist, add endpoint metadata failed, %s", err)
			}

			valResID, err := store.AddValidationResult(ctx)
			if err != nil {
				return fmt.Errorf("adding new validation result ID failed, %s", err)
			}
			existingEndpt.ValidationID = valResID

			err = store.AddValidation(ctx, validation, valResID)
			if err != nil {
				return fmt.Errorf("error adding validation rows to table, %s", err)
			}

			store.DeleteFHIREndpointInfo(ctx, existingEndpt)

			fhirEndpointList, err := store.GetFHIREndpointUsingURL(ctx, existingEndpt.URL)
			if err != nil {
				return errors.Wrap(err, "error getting fhir endpoints from DB")
			}

			for _, fhirEp := range fhirEndpointList {

				log.Info("inside for loop for fhirEndpointList")
				developerNames := softwareListMap[fhirEp.ListSource].ChplDeveloper
				productIds := softwareListMap[fhirEp.ListSource].ChplProductIDs

				if len(developerNames) == 0 {
					err = store.AddFHIREndpointInfo(ctx, existingEndpt, metadataID)
					if err != nil {
						return fmt.Errorf("does exist, add to fhir_endpoints_info failed, %s", err)
					}
				} else {
					isDeveloperSeen := make(map[string]bool)

					for _, developerName := range developerNames {

						log.Info("inside for loop for developerNames")
						if !isDeveloperSeen[developerName] {

							log.Info("inside !isdeveloperseen")

							err = chplmapper.MatchEndpointToVendor(ctx, existingEndpt, store, developerName)
							if err != nil {
								return fmt.Errorf("doesn't exist, match endpoint to vendor failed, %s", err)
							}

							var productIdsPerDeveloper []string
							for idx, productId := range productIds {
								if developerNames[idx] == developerName {
									productIdsPerDeveloper = append(productIdsPerDeveloper, productId)
								}
							}

							existingEndpt.HealthITProductID = 0 // Reset HealthITProductID to 0 before matching to product
							err = chplmapper.MatchEndpointToProduct(ctx, existingEndpt, store, fmt.Sprintf("%v", qa.chplMatchFile), productIdsPerDeveloper)
							if err != nil {
								return fmt.Errorf("doesn't exist, match endpoint to product failed, %s", err)
							}

							log.Info("Updating fhir endpoint data")

							err = store.AddFHIREndpointInfo(ctx, existingEndpt, metadataID)
							if err != nil {
								return fmt.Errorf("does exist, add to fhir_endpoints_info failed, %s", err)
							}

							isDeveloperSeen[developerName] = true
						}
					}
				}
			}
		} else {
			metadataID, err := store.AddFHIREndpointMetadata(ctx, existingEndpt.Metadata)
			if err != nil {
				return fmt.Errorf("just adding endpoint metadata failed, %s", err)
			}

			err = store.UpdateMetadataIDInfo(ctx, metadataID, existingEndpt.ID)
			if err != nil {
				return fmt.Errorf("just adding the Metadata ID failed, %s", err)
			}
		}
	}

	return nil
}

func removeNoLongerExistingVersionsInfos(ctx context.Context, store *postgresql.Store, url string, supportedVersions []string) error {
	// If there is a requestedVersion for a URL in fhir_endpoints_info that is no longer in supportedVersions
	// then we need to remove those fhir_endpoint_info entries
	endptInfos, err := store.GetFHIREndpointInfosByURLWithDifferentRequestedVersion(ctx, url, supportedVersions)
	if err != nil {
		return err
	}
	for _, infoEntry := range endptInfos {
		err = store.DeleteFHIREndpointInfo(ctx, infoEntry)
		if err != nil {
			return err
		}
	}
	return nil
}

func saveVersionResponseMsgInDB(message []byte, args *map[string]interface{}) error {
	var err error
	var existingEndpts []*endpointmanager.FHIREndpoint
	var msgJSON map[string]interface{}
	// Get arguments
	qa, ok := (*args)["queryArgs"].(versionsQueryArgs)
	if !ok {
		return fmt.Errorf("unable to parse args into versionsQueryArgs")
	}

	err = json.Unmarshal(message, &msgJSON)
	if err != nil {
		return err
	}

	url, ok := msgJSON["url"].(string)
	if !ok {
		return fmt.Errorf("unable to cast message URL to string")
	}

	if err != nil {
		return err
	}

	store := qa.store
	ctx := qa.ctx

	existingEndpts, err = store.GetFHIREndpointUsingURL(ctx, url)
	if err != nil {
		return err
	}

	resp, _ := msgJSON["versionsResponse"].(map[string]interface{})
	var vsr versionsoperatorparser.VersionsResponse
	vsr.Response = resp
	for _, endpt := range existingEndpts {
		// Only update if versions have changed
		if !endpt.VersionsResponse.Equal(vsr) {
			endpt.VersionsResponse = vsr
			err = store.UpdateFHIREndpoint(ctx, endpt)
			if err != nil {
				return err
			}
		}
	}

	// Dispatch query for CapabilityStatement here
	// Set up the queue for sending messages to capabilityquerier
	mq := qa.capQueryQueue
	channelID := qa.capQueryChannelID
	capQueryEndptQName := viper.GetString("endptinfo_capquery_qname")
	var supportedVersions []string
	supportedVersions = vsr.GetSupportedVersions()

	defaultVersion := vsr.GetDefaultVersion()

	supportedVersions = append(supportedVersions, "None")

	err = removeNoLongerExistingVersionsInfos(ctx, store, url, supportedVersions)
	if err != nil {
		return err
	}

	for _, version := range supportedVersions {
		// send URL and version of FHIR version to request
		var message map[string]string = make(map[string]string)
		message["url"] = url
		message["requestVersion"] = version
		message["defaultVersion"] = defaultVersion
		var msgBytes []byte
		msgBytes, err = json.Marshal(message)
		if err != nil {
			return err
		}
		err = accessqueue.SendToQueue(ctx, string(msgBytes), &mq, &channelID, capQueryEndptQName)
		if err != nil {
			return err
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
	args["queryArgs"] = capStatQueryArgs{
		store:                    store,
		ctx:                      ctx,
		chplMatchFile:            "/etc/lantern/resources/CHPLProductMapping.json",
		chplEndpointListInfoFile: "/etc/lantern/resources/CHPLProductsInfo.json",
	}

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

// ReceiveVersionResponses connects to the given message queue channel (qname) and receives the
// versions response from it. It then saves the versions response and queries the versions advertized
func ReceiveVersionResponses(ctx context.Context,
	store *postgresql.Store,
	messageQueue lanternmq.MessageQueue,
	channelID lanternmq.ChannelID,
	qName string,
	capQueryQueue lanternmq.MessageQueue,
	capQueryChannelID lanternmq.ChannelID) error {
	args := make(map[string]interface{})

	args["queryArgs"] = versionsQueryArgs{
		ctx:               ctx,
		capQueryChannelID: capQueryChannelID,
		capQueryQueue:     capQueryQueue,
		store:             store,
	}

	messages, err := messageQueue.ConsumeFromQueue(channelID, qName)
	if err != nil {
		return err
	}

	errs := make(chan error)
	go messageQueue.ProcessMessages(ctx, messages, saveVersionResponseMsgInDB, &args, errs)

	for elem := range errs {
		log.Warn(elem)
	}

	return nil
}
