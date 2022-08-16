package capabilityquerier

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	aq "github.com/onc-healthit/lantern-back-end/lanternmq/pkg/accessqueue"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type EndpointType string

const (
	metadata  EndpointType = "metadata"
	wellknown EndpointType = "well-known"
)

var fhir3PlusJSONMIMEType = "application/fhir+json"
var fhir2LessJSONMIMEType = "application/json+fhir"
var fhir2LessXMLMIMEType = "application/xml+fhir"
var fhir3PlusXMLMIMEType = "application/fhir+xml"

var ssl30 = "SSL 3.0"
var tls10 = "TLS 1.0"
var tls11 = "TLS 1.1"
var tls12 = "TLS 1.2"
var tls13 = "TLS 1.3"
var tlsUnknown = "TLS version unknown"
var tlsNone = "No TLS"

// Message is the structure that gets sent on the queue with capability statement inforation. It includes the URL of
// the FHIR API, any errors from making the FHIR API request, the MIME type, the TLS version, and the capability
// statement itself.
type Message struct {
	URL                      string      `json:"url"`
	Err                      string      `json:"err"`
	MIMETypes                []string    `json:"mimeTypes"`
	TLSVersion               string      `json:"tlsVersion"`
	HTTPResponse             int         `json:"httpResponse"`
	CapabilityStatement      interface{} `json:"capabilityStatement"`
	CapabilityStatementBytes []byte      `json:"capabilityStatementBytes"`
	SMARTHTTPResponse        int         `json:"smarthttpResponse"`
	SMARTResp                interface{} `json:"smartResp"`
	SMARTRespBytes           []byte      `json:"smartRespBytes"`
	ResponseTime             float64     `json:"responseTime"`
	RequestedFhirVersion     string      `json:"requestedFhirVersion"`
	DefaultFhirVersion       string      `json:"defaultFhirVersion"`
}

// VersionMessage is the structure that gets sent on the queue with $versions response inforation. It includes the URL of
// the FHIR API, any errors from making the FHIR $versions request, and the $versions response itself.
type VersionsMessage struct {
	URL              string      `json:"url"`
	Err              string      `json:"err"`
	VersionsResponse interface{} `json:"versionsResponse"`
}

// QuerierArgs is a struct of the queue connection information (MessageQueue, ChannelID, and QueueName) as well as
// the Client and FhirURL for querying
type QuerierArgs struct {
	FhirURL        string
	RequestVersion string
	DefaultVersion string
	Client         *http.Client
	MessageQueue   *lanternmq.MessageQueue
	ChannelID      *lanternmq.ChannelID
	QueueName      string
	UserAgent      string
	Store          *postgresql.Store
}

// GetAndSendVersionsResponse gets a $versions response from a FHIR API endpoint and then puts the versions
// response and accompanying data on a receiving queue.
func GetAndSendVersionsResponse(ctx context.Context, args *map[string]interface{}) error {
	var jsonResponse interface{}

	qa, ok := (*args)["querierArgs"].(QuerierArgs)
	if !ok {
		return fmt.Errorf("unable to cast querierArgs to type QuerierArgs from arguments")
	}

	message := VersionsMessage{
		URL: qa.FhirURL,
	}

	// If Finished message, pass on to versions response queue
	if qa.FhirURL != "FINISHED" {
		// Cast string url to type url then cast back to string to ensure url string in correct url format
		castURL, err := url.Parse(qa.FhirURL)
		if err != nil {
			return fmt.Errorf("endpoint URL parsing error: %s", err.Error())
		}
		versionsURL := endpointmanager.NormalizeVersionsURL(castURL.String())
		// Add a short time buffer before sending HTTP request to reduce burden on servers hosting multiple endpoints
		time.Sleep(time.Duration(500 * time.Millisecond))
		req, err := http.NewRequest("GET", versionsURL, nil)
		if err != nil {
			log.Errorf("unable to create new GET request from URL: " + versionsURL)
		} else {
			req.Header.Set("User-Agent", qa.UserAgent)
			trace := &httptrace.ClientTrace{}
			req = req.WithContext(httptrace.WithClientTrace(ctx, trace))

			httpResponseCode, _, _, versionsResponse, _, err := requestWithMimeType(req, "application/json", qa.Client)
			// If an error occurs with the version request we still want to proceed with the capability request
			if err != nil {
				log.Infof("Error requesting versions response: %s", err.Error())
			} else {
				if httpResponseCode == 200 && versionsResponse != nil {
					err = json.Unmarshal(versionsResponse, &(jsonResponse))
					if err != nil {
						log.Errorf("Error unmarshalling versions response: %s", err.Error())
					}
				}
			}
		}

		message.VersionsResponse = jsonResponse
	}
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return errors.Wrapf(err, "error marshalling json message for request to %s", qa.FhirURL)
	}
	msgStr := string(msgBytes)
	// Blank context passed in to SendToQueue to prevent terminating error due to an endpoint timeout
	tempCtx := context.Background()
	err = aq.SendToQueue(tempCtx, msgStr, qa.MessageQueue, qa.ChannelID, qa.QueueName)
	if err != nil {
		return errors.Wrapf(err, "error sending versions response for FHIR endpoint %s to queue '%s'", qa.FhirURL, qa.QueueName)
	}
	return nil
}

// GetAndSendCapabilityStatement gets a capability statement from a FHIR API endpoint and then puts the capability
// statement and accompanying data on a receiving queue.
// The args are expected to be a map of the string "querierArgs" to the above QuerierArgs struct. It is formatted
// this way in order for it to be able to be called by a worker (see endpointmanager/pkg/workers)
func GetAndSendCapabilityStatement(ctx context.Context, args *map[string]interface{}) error {
	// Get arguments

	qa, ok := (*args)["querierArgs"].(QuerierArgs)
	if !ok {
		return fmt.Errorf("unable to cast querierArgs to type QuerierArgs from arguments")
	}

	var err error

	endpt, err := qa.Store.GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx, qa.FhirURL, qa.RequestVersion)
	var mimeTypes []string
	if err == sql.ErrNoRows {
		mimeTypes = []string{}
	} else if err != nil {
		select {
		case <-ctx.Done():
			log.Warnf("Got error: could not connect to database: %s", qa.FhirURL)
			mimeTypes = []string{}
		default:
			log.Warnf("Got error:\n%s\n\nfrom URL: %s", err.Error(), qa.FhirURL)
			return err
		}
	} else {
		mimeTypes = endpt.MIMETypes
	}

	userAgent := qa.UserAgent
	message := Message{
		URL:                  qa.FhirURL,
		RequestedFhirVersion: qa.RequestVersion,
		DefaultFhirVersion:   qa.DefaultVersion,
		MIMETypes:            mimeTypes,
	}
	// Cast string url to type url then cast back to string to ensure url string in correct url format
	castURL, err := url.Parse(qa.FhirURL)
	if err != nil {
		return fmt.Errorf("endpoint URL parsing error: %s", err.Error())
	}
	metadataURL := endpointmanager.NormalizeEndpointURL(castURL.String())
	// Query fhir endpoint
	err = requestCapabilityStatementAndSmartOnFhir(ctx, metadataURL, metadata, qa.Client, userAgent, &message)
	if err != nil {
		select {
		case <-ctx.Done():
			log.Warnf("Got error: server could not be reached from URL: %s", qa.FhirURL)
			message.Err = "server could not be reached from URL: " + metadataURL
		default:
			log.Warnf("Got error:\n%s\n\nfrom URL: %s", err.Error(), qa.FhirURL)
			message.Err = err.Error()
		}
	}

	wellKnownURL := endpointmanager.NormalizeWellKnownURL(castURL.String())
	// Query well known endpoint
	err = requestCapabilityStatementAndSmartOnFhir(ctx, wellKnownURL, wellknown, qa.Client, userAgent, &message)
	if err != nil {
		log.Warnf("Got error:\n%s\n\nfrom wellknown URL: %s", err.Error(), wellKnownURL)
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return errors.Wrapf(err, "error marshalling json message for request to %s", qa.FhirURL)
	}
	msgStr := string(msgBytes)
	// Blank context passed in to SendToQueue to prevent terminating error due to an endpoint timeout
	tempCtx := context.Background()
	err = aq.SendToQueue(tempCtx, msgStr, qa.MessageQueue, qa.ChannelID, qa.QueueName)
	if err != nil {
		return errors.Wrapf(err, "error sending capability statement for FHIR endpoint %s to queue '%s'", qa.FhirURL, qa.QueueName)
	}

	return nil
}

// fills out message with http response code, tls version, capability statement, and supported mime types
func requestCapabilityStatementAndSmartOnFhir(ctx context.Context, fhirURL string, endptType EndpointType, client *http.Client, userAgent string, message *Message) error {
	var err error
	var httpErr error
	var httpResponseCode int
	var mimeTypeWorked bool
	var tlsVersion string
	var capResp []byte
	var jsonResponse interface{}
	var responseTime float64
	var triedMIMEType string

	// Add a short time buffer before sending HTTP request to reduce burden on servers hosting multiple endpoints
	time.Sleep(time.Duration(500 * time.Millisecond))
	req, err := http.NewRequest("GET", fhirURL, nil)
	if err != nil {
		return errors.Wrap(err, "unable to create new GET request from URL: "+fhirURL)
	}
	req.Header.Set("User-Agent", userAgent)
	trace := &httptrace.ClientTrace{}
	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))

	// If there is a requested fhir version, set the fhirVersion in the request header
	if message.RequestedFhirVersion != "None" {
		req.Header.Set("fhirVersion", message.RequestedFhirVersion)
	}

	// If there is a mime type saved in the database for this URL, try those ones first when requesting the capability statement
	if len(message.MIMETypes) == 1 {
		savedMIME := message.MIMETypes[0]
		httpResponseCode, tlsVersion, mimeTypeWorked, capResp, responseTime, httpErr = requestWithMimeType(req, savedMIME, client)
		if httpErr != nil && httpResponseCode != 0 {
			return err
		}
	}

	// If there was no MIME type saved in the database, or the saved MIME type did not work, go through process of trying others
	if len(message.MIMETypes) != 1 || httpResponseCode != http.StatusOK || !mimeTypeWorked {
		// If the endpoint is a well known endpoint and it did not already have MIME type saved, try the fhir3PlusJSONMIMEType
		if endptType == wellknown {
			if len(message.MIMETypes) == 0 {
				httpResponseCode, _, _, capResp, _, httpErr = requestWithMimeType(req, fhir3PlusJSONMIMEType, client)
				if httpErr != nil && httpResponseCode != 0 {
					return err
				}
			}
		} else if endptType == metadata {

			// If there was a MIME type saved in the database, remove it from the list of MIME types since it did not work
			oldMIMEType := ""
			if len(message.MIMETypes) == 1 {
				oldMIMEType = message.MIMETypes[0]
				message.MIMETypes = []string{}
			} else if len(message.MIMETypes) > 1 {
				message.MIMETypes = []string{}
			}

			// Try fhir3PlusJSONMIMEType first if it was not the MIME type saved in the database
			if oldMIMEType != fhir3PlusJSONMIMEType {
				httpResponseCode, tlsVersion, mimeTypeWorked, capResp, responseTime, httpErr = requestWithMimeType(req, fhir3PlusJSONMIMEType, client)
				if httpErr != nil && httpResponseCode != 0 {
					return err
				}
				triedMIMEType = fhir3PlusJSONMIMEType
			}
			// Try fhir2LessJSONMIMEType second if it was not the MIME type saved in the database and the first MIME type did not work
			if oldMIMEType != fhir2LessJSONMIMEType && (!mimeTypeWorked || httpResponseCode != http.StatusOK) {
				httpResponseCode, tlsVersion, mimeTypeWorked, capResp, responseTime, httpErr = requestWithMimeType(req, fhir2LessJSONMIMEType, client)
				if httpErr != nil && httpResponseCode != 0 {
					return err
				}
				triedMIMEType = fhir2LessJSONMIMEType
			}
			// Try fhir3PlusXMLMIMEType third if it was not the MIME type saved in the database and the first two MIME types did not work
			if oldMIMEType != fhir3PlusXMLMIMEType && (!mimeTypeWorked || httpResponseCode != http.StatusOK) {
				httpResponseCode, tlsVersion, mimeTypeWorked, capResp, responseTime, httpErr = requestWithMimeType(req, fhir3PlusXMLMIMEType, client)
				if httpErr != nil && httpResponseCode != 0 {
					return err
				}
				triedMIMEType = fhir3PlusXMLMIMEType
			}
			// Try fhir2LessXMLMIMEType last if it was not the MIME type saved in the database and the first three MIME types did not work
			if oldMIMEType != fhir2LessXMLMIMEType && (!mimeTypeWorked || httpResponseCode != http.StatusOK) {
				httpResponseCode, tlsVersion, mimeTypeWorked, capResp, responseTime, httpErr = requestWithMimeType(req, fhir2LessXMLMIMEType, client)
				if httpErr != nil && httpResponseCode != 0 {
					return err
				}
				triedMIMEType = fhir2LessXMLMIMEType
			}

			// If there are no MIME types saved, and a new MIME type worked and had a valid HTTP response, save it in the db
			if len(message.MIMETypes) != 1 && mimeTypeWorked && httpResponseCode == http.StatusOK {
				message.MIMETypes = append(message.MIMETypes, triedMIMEType)
			}
		}
	}

	if capResp != nil {
		if endptType == metadata {
			message.CapabilityStatementBytes = capResp
		} else if endptType == wellknown {
			message.SMARTRespBytes = capResp
		}
		err := json.Unmarshal(capResp, &jsonResponse)
		if err == nil {
			if endptType == metadata {
				message.CapabilityStatement = jsonResponse
			} else if endptType == wellknown {
				message.SMARTResp = jsonResponse
			}
		} else {
			if httpErr == nil {
				httpErr = err
			}
		}
	}

	switch endptType {
	case metadata:
		message.TLSVersion = tlsVersion
		message.HTTPResponse = httpResponseCode
		message.ResponseTime = responseTime
	case wellknown:
		message.SMARTHTTPResponse = httpResponseCode
	}

	return httpErr
}

func getTLSVersion(resp *http.Response) string {
	if resp.TLS != nil {
		switch resp.TLS.Version {
		case tls.VersionSSL30: //nolint
			return ssl30
		case tls.VersionTLS10:
			return tls10
		case tls.VersionTLS11:
			return tls11
		case tls.VersionTLS12:
			return tls12
		case tls.VersionTLS13:
			return tls13
		default:
			return tlsUnknown
		}
	}
	return tlsNone
}

func isJSONMIMEType(mimeType string) bool {
	return strings.Contains(mimeType, "json")
}

func mimeTypesMatch(reqMimeType string, respMimeType string) bool {
	respMimeTypes := strings.Split(respMimeType, "; ")
	for _, rmt := range respMimeTypes {
		if rmt == reqMimeType {
			return true
		}
	}
	return false
}

// responds with:
// http status code
// tls version
// mime type match
// capability statement
// error
func requestWithMimeType(req *http.Request, mimeType string, client *http.Client) (int, string, bool, []byte, float64, error) {
	var httpResponseCode int
	var tlsVersion string
	var capStat []byte

	mimeMatches := false

	req.Header.Set("Accept", mimeType)

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		// Return http status code 0 on failure
		return 0, "", false, nil, -1, errors.Wrapf(err, "making the GET request to %s failed", req.URL.String())
	}

	var responseTime = float64(time.Since(start).Seconds())

	httpResponseCode = resp.StatusCode
	if httpResponseCode == http.StatusOK {
		respMimeType := resp.Header.Get("Content-Type")
		// endpoints generally return an xml mime type by default.
		// checking that it's a json mime type confirms that it processes the JSON type request.
		// however, it doesn't necessarily match the request type exactly and seems to cache the
		// first JSON request type it receives and continues to respond with that.
		if isJSONMIMEType(respMimeType) {
			defer resp.Body.Close()
			mimeMatches = true

			capStat, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return -1, "", false, nil, -1, errors.Wrapf(err, "reading the response from %s failed", req.URL.String())
			}
		}
	}

	tlsVersion = getTLSVersion(resp)

	return httpResponseCode, tlsVersion, mimeMatches, capStat, responseTime, nil
}
