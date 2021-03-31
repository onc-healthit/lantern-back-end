package capabilityquerier

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/versionsoperation"
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
	URL                 string      `json:"url"`
	Err                 string      `json:"err"`
	MIMETypes           []string    `json:"mimeTypes"`
	TLSVersion          string      `json:"tlsVersion"`
	HTTPResponse        int         `json:"httpResponse"`
	CapabilityStatement interface{} `json:"capabilityStatement"`
	SMARTHTTPResponse   int         `json:"smarthttpResponse"`
	SMARTResp           interface{} `json:"smartResp"`
	ResponseTime        float64     `json:"responseTime"`
}
type VersionsMessage struct {
	URL                 string      `json:"url"`
	Err                 string      `json:"err"`
	VersionsResponse versionsoperation.VersionsResponse	`json:"versionsResponse"`
}

// QuerierArgs is a struct of the queue connection information (MessageQueue, ChannelID, and QueueName) as well as
// the Client and FhirURL for querying
type QuerierArgs struct {
	FhirURL      string
	Client       *http.Client
	MessageQueue *lanternmq.MessageQueue
	ChannelID    *lanternmq.ChannelID
	QueueName    string
	UserAgent    string
	Store        *postgresql.Store
}


func GetAndSendVersionsResponse(ctx context.Context, args *map[string]interface{}) error {
	var jsonResponse versionsoperation.VersionsResponse

	qa, ok := (*args)["querierArgs"].(QuerierArgs)
	if !ok {
		return fmt.Errorf("unable to cast querierArgs to type QuerierArgs from arguments")
	}
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
		return errors.Wrap(err, "unable to create new GET request from URL: "+ versionsURL)
	}
	req.Header.Set("User-Agent", qa.UserAgent)
	trace := &httptrace.ClientTrace{}
	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))
	httpResponseCode, _, _, versionsResponse, _, err := requestWithMimeType(req, "application/json", qa.Client)

	message := VersionsMessage{
		URL:       qa.FhirURL,
	}

	if httpResponseCode == 200 && versionsResponse != nil {
		err = json.Unmarshal(versionsResponse, &(jsonResponse))
		if err != nil {
			return err
		}
	}

	message.VersionsResponse = jsonResponse

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

	endpt, err := qa.Store.GetFHIREndpointInfoUsingURL(ctx, qa.FhirURL)
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
		URL:       qa.FhirURL,
		MIMETypes: mimeTypes,
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
	var httpResponseCode int
	var mimeTypeWorked bool
	var otherMimeWorked bool
	var tlsVersion string
	var capResp []byte
	var jsonResponse interface{}
	var responseTime float64

	// Add a short time buffer before sending HTTP request to reduce burden on servers hosting multiple endpoints
	time.Sleep(time.Duration(500 * time.Millisecond))
	req, err := http.NewRequest("GET", fhirURL, nil)
	if err != nil {
		return errors.Wrap(err, "unable to create new GET request from URL: "+fhirURL)
	}
	req.Header.Set("User-Agent", userAgent)
	trace := &httptrace.ClientTrace{}
	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))

	randomMimeIdx := 0
	firstMIME := fhir3PlusJSONMIMEType

	// If there are mime types saved in the database for this URL
	if endptType == metadata && len(message.MIMETypes) > 0 {
		// Choose a random mime type in the list if there's more than one
		if len(message.MIMETypes) == 2 {
			rand.Seed(time.Now().UnixNano())
			randomMimeIdx = rand.Intn(2)
			firstMIME = message.MIMETypes[randomMimeIdx]
			httpResponseCode, tlsVersion, mimeTypeWorked, capResp, responseTime, err = requestWithMimeType(req, firstMIME, client)
		} else {
			firstMIME = message.MIMETypes[randomMimeIdx]
			httpResponseCode, tlsVersion, mimeTypeWorked, capResp, responseTime, err = requestWithMimeType(req, firstMIME, client)
		}
		if err != nil {
			return err
		}
	} else if endptType == wellknown && len(message.MIMETypes) > 0 {
		firstMIME = message.MIMETypes[0]
		httpResponseCode, _, _, capResp, _, err = requestWithMimeType(req, firstMIME, client)
		if err != nil {
			return err
		}
	} else {
		httpResponseCode, tlsVersion, mimeTypeWorked, capResp, responseTime, err = requestWithMimeType(req, fhir3PlusJSONMIMEType, client)
		if err != nil {
			return err
		}
	}

	otherMime := fhir2LessJSONMIMEType
	if endptType == metadata {
		if httpResponseCode != http.StatusOK || !mimeTypeWorked {
			// Try the other mime type and remove the mime type that was initially saved
			// but no longer works
			if len(message.MIMETypes) == 2 {
				otherMimeIdx := (randomMimeIdx + 1) % 2
				otherMime = message.MIMETypes[otherMimeIdx]
				message.MIMETypes = []string{otherMime}
			} else if len(message.MIMETypes) == 1 {
				if message.MIMETypes[0] == otherMime {
					otherMime = fhir3PlusJSONMIMEType
				}
				message.MIMETypes = []string{}
			}
			// replace all values based on the other mime type if there were any issues with the first mime type request
			httpResponseCode, tlsVersion, otherMimeWorked, capResp, responseTime, err = requestWithMimeType(req, otherMime, client)
			if err != nil {
				return err
			}
		} else if len(message.MIMETypes) == 0 {
			// only check fhir 2 mime type support if the first request worked and there were no
			// mimeTypes saved in the database
			_, _, otherMimeWorked, _, _, err = requestWithMimeType(req, otherMime, client)
			if err != nil {
				return err
			}
		}

		finalMimeList := []string{}
		// If there was a 2nd saved mime type and it also did not work, remove it from the MIMETypes array
		if len(message.MIMETypes) == 1 && (httpResponseCode != http.StatusOK || !otherMimeWorked) {
			message.MIMETypes = []string{}
		} else if otherMimeWorked {
			// If the 2nd tried mime type did work, add it to the MIMETypes array
			finalMimeList = append(finalMimeList, otherMime)
		}
		// If the first mimeType worked and it wasn't saved in the database, add it to MIMETypes array
		if mimeTypeWorked && len(message.MIMETypes) == 0 {
			finalMimeList = append(finalMimeList, firstMIME)
		}
		// Update the message.MIMETypes as long as nothing was already saved there
		if len(message.MIMETypes) == 0 {
			message.MIMETypes = finalMimeList
		}
	}

	if capResp != nil {
		err = json.Unmarshal(capResp, &(jsonResponse))
		if err != nil {
			return err
		}
	}

	switch endptType {
	case metadata:
		message.TLSVersion = tlsVersion
		message.HTTPResponse = httpResponseCode
		message.CapabilityStatement = jsonResponse
		message.ResponseTime = responseTime
	case wellknown:
		message.SMARTHTTPResponse = httpResponseCode
		message.SMARTResp = jsonResponse
	}

	return nil
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

// responds with
//   http status code
//   tls version
//   mime type match
//   capability statement
//   error
func requestWithMimeType(req *http.Request, mimeType string, client *http.Client) (int, string, bool, []byte, float64, error) {
	var httpResponseCode int
	var tlsVersion string
	var capStat []byte

	mimeMatches := false

	req.Header.Set("Accept", mimeType)

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		return -1, "", false, nil, -1, errors.Wrapf(err, "making the GET request to %s failed", req.URL.String())
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
