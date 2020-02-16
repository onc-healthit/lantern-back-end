package capabilityquerier

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/pkg/errors"
)

var fhir3PlusJSONMIMEType = "application/fhir+json"
var fhir2LessJSONMIMEType = "application/json+fhir"

var ssl30 = "SSL 3.0"
var tls10 = "TLS 1.0"
var tls11 = "TLS 1.1"
var tls12 = "TLS 1.2"
var tls13 = "TLS 1.3"
var tlsUnknown = "TLS version unknown"

// Message is the structure that gets sent on the queue with capability statement inforation. It includes the URL of
// the FHIR API, any errors from making the FHIR API request, the MIME type, the TLS version, and the capability
// statement itself.
type Message struct {
	URL                 string      `json:"url"`
	Err                 string      `json:"err"`
	MimeType            string      `json:"mimetype"`
	TLSVersion          string      `json:"tlsVersion"`
	CapabilityStatement interface{} `json:"capabilityStatement"`
}

// GetAndSendCapabilityStatement gets a capability statement from a FHIR API endpoints and then puts the capability
// statement and accompanying data on a receiving queue.
func GetAndSendCapabilityStatement(
	ctx context.Context,
	fhirURL *url.URL,
	client *http.Client,
	mq *lanternmq.MessageQueue,
	ch *lanternmq.ChannelID,
	queueName string) error {

	var err error
	message := Message{
		URL: fhirURL.String(),
	}

	capResp, mimeType, tlsVersion, err := requestCapabilityStatement(ctx, fhirURL, client)
	if err == nil {
		message.MimeType = mimeType
		message.TLSVersion = tlsVersion
		err = json.Unmarshal(capResp, &(message.CapabilityStatement))
		if err != nil {
			message.Err = err.Error()
		}
	} else {
		message.Err = err.Error()
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return errors.Wrapf(err, "error marshalling json message for request to %s", fhirURL.String())
	}
	msgStr := string(msgBytes)

	err = sendToQueue(ctx, msgStr, mq, ch, queueName)
	if err != nil {
		return errors.Wrapf(err, "error sending capability statement for FHIR endpoint %s to queue '%s'", fhirURL.String(), queueName)
	}

	return nil
}

func requestCapabilityStatement(ctx context.Context, fhirURL *url.URL, client *http.Client) ([]byte, string, string, error) {
	var err error
	var resp *http.Response
	var is406 bool

	req, err := http.NewRequest("GET", fhirURL.String(), nil)
	if err != nil {
		return nil, "", "", errors.Wrap(err, "unable to create new GET request from URL: "+fhirURL.String())
	}
	req = req.WithContext(ctx)

	// make the request using a JSON mime type to get a JSON response.
	// track the mime type to provide to the queue as data re the request.
	tryOtherMimeType := false
	mimeType := fhir3PlusJSONMIMEType
	resp, is406, err = requestWithMimeType(req, mimeType, client)
	if err != nil {
		return nil, "", "", err
	} else if is406 {
		tryOtherMimeType = true
	} else {
		defer resp.Body.Close()

		// if the response type isn't right for FHIR 3+, it's possible it's just an older version.
		respMimeType := resp.Header.Get("Content-Type")
		if !mimeTypesMatch(mimeType, respMimeType) {
			tryOtherMimeType = true
		}
	}

	if tryOtherMimeType {
		mimeType = fhir2LessJSONMIMEType
		resp, is406, err = requestWithMimeType(req, mimeType, client)
		if err != nil {
			return nil, "", "", err
		} else if is406 {
			return nil, "", "", fmt.Errorf("GET request to %s responded with status 406 Not Acceptable", fhirURL.String())
		}
		defer resp.Body.Close()

		respMimeType := resp.Header.Get("Content-Type")
		if !mimeTypesMatch(mimeType, respMimeType) {
			return nil, "", "", fmt.Errorf("response MIME type (%s) does not match JSON request MIME types for FHIR 3+ (%s) or FHIR 2- (%s)",
				respMimeType, fhir3PlusJSONMIMEType, fhir2LessJSONMIMEType)
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", "", errors.Wrapf(err, "reading the response from %s failed", fhirURL.String())
	}

	return body, mimeType, getTLSVersion(resp), nil
}

func getTLSVersion(resp *http.Response) string {
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

func mimeTypesMatch(reqMimeType string, respMimeType string) bool {
	respMimeTypes := strings.Split(respMimeType, "; ")
	for _, rmt := range respMimeTypes {
		if rmt == reqMimeType {
			return true
		}
	}
	return false
}

// responds with the http.Response, whether or not the status code was a 406 (indicating we should try with a different
// mimetype), and any errors.
func requestWithMimeType(req *http.Request, mimeType string, client *http.Client) (*http.Response, bool, error) {
	req.Header.Set("Accept", mimeType)

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, errors.Wrapf(err, "making the GET request to %s failed", req.URL.String())
	}

	if resp.StatusCode == http.StatusNotAcceptable {
		return nil, true, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GET request to %s responded with status %s", req.URL.String(), resp.Status)
	}

	return resp, false, nil
}

func sendToQueue(
	ctx context.Context,
	message string,
	mq *lanternmq.MessageQueue,
	ch *lanternmq.ChannelID,
	queueName string) error {

	// don't send the message if the context is done
	select {
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "unable to send message to queue - context ended")
	default:
		// ok
	}

	err := (*mq).PublishToQueue(*ch, queueName, message)
	if err != nil {
		return err
	}

	return nil
}
