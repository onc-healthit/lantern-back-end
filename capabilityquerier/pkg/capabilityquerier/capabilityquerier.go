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
		return err
	} else {
		mimeTypes = endpt.MIMETypes
	}

	userAgent := qa.UserAgent
	message := Message{
		URL: qa.FhirURL,
	}
	// Cast string url to type url then cast back to string to ensure url string in correct url format
	castURL, err := url.Parse(qa.FhirURL)
	if err != nil {
		log.Warnf("Error parsing URL string %s\n", qa.FhirURL)
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
	var supportsFHIR3MIMEType bool
	var supportsFHIR2MIMEType bool
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

	if endptType == wellknown && len(message.MIMETypes) > 0 {
		httpResponseCode, _, _, capResp, _, err = requestWithMimeType(req, message.MIMETypes[0], client)
		if err != nil {
			return err
		}
	} else {
		httpResponseCode, tlsVersion, supportsFHIR3MIMEType, capResp, responseTime, err = requestWithMimeType(req, fhir3PlusJSONMIMEType, client)
		if err != nil {
			return err
		}

	}

	if endptType == metadata {
		if httpResponseCode != http.StatusOK || !supportsFHIR3MIMEType {
			// replace all values based on fhir 2 mime type if there were any issues with fhir 3 mime type request
			httpResponseCode, tlsVersion, supportsFHIR2MIMEType, capResp, responseTime, err = requestWithMimeType(req, fhir2LessJSONMIMEType, client)
			if err != nil {
				return err
			}
		} else {
			// only chech fhir 2 mime type support
			_, _, supportsFHIR2MIMEType, _, _, err = requestWithMimeType(req, fhir2LessJSONMIMEType, client)
			if err != nil {
				return err
			}
		}
		if supportsFHIR2MIMEType {
			message.MIMETypes = append(message.MIMETypes, fhir2LessJSONMIMEType)
		}
		if supportsFHIR3MIMEType {
			message.MIMETypes = append(message.MIMETypes, fhir3PlusJSONMIMEType)
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
		defer resp.Body.Close()
		respMimeType := resp.Header.Get("Content-Type")
		// endpoints generally return an xml mime type by default.
		// checking that it's a json mime type confirms that it processes the JSON type request.
		// however, it doesn't necessarily match the request type exactly and seems to cache the
		// first JSON request type it receives and continues to respond with that.
		if isJSONMIMEType(respMimeType) {
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
