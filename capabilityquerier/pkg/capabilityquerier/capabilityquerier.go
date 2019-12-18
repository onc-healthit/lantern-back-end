package capabilityquerier

import (
	"context"
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

type Message struct {
	URL                 string      `json:"url"`
	Err                 string      `json:"err"`
	MimeType            string      `json:"mimetype"`
	CapabilityStatement interface{} `json:"capabilityStatement"`
}

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

	capResp, mimeType, err := requestCapabilityStatement(ctx, fhirURL, client)
	if err == nil {
		message.MimeType = mimeType
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
		return errors.Wrapf(err, "error sending capability statement to queue '%s' over channel '%v'", queueName, *ch)
	}

	return nil
}

func requestCapabilityStatement(ctx context.Context, fhirURL *url.URL, client *http.Client) ([]byte, string, error) {
	var err error
	var resp *http.Response

	req, err := http.NewRequest("GET", fhirURL.String(), nil)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to create new GET request from URL: "+fhirURL.String())
	}
	req = req.WithContext(ctx)

	// make the request using a JSON mime type to get a JSON response.
	// track the mime type to provide to the queue as data re the request.
	tryOtherMimeType := false
	mimeType := fhir3PlusJSONMIMEType
	resp, err = requestWithMimeType(req, mimeType, client)
	if err != nil {
		// it's possible that the error is due to a 406 Not Acceptable error for the wrong meme type, so we
		// don't want to return here
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
		resp, err = requestWithMimeType(req, mimeType, client)
		if err != nil {
			return nil, "", err
		}
		defer resp.Body.Close()

		respMimeType := resp.Header.Get("Content-Type")
		if !mimeTypesMatch(mimeType, respMimeType) {
			return nil, "", fmt.Errorf("response MIME type (%s) does not match JSON request MIME types for FHIR 3+ (%s) or FHIR 2- (%s)",
				respMimeType, fhir3PlusJSONMIMEType, fhir2LessJSONMIMEType)
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", errors.Wrapf(err, "reading the response from %s failed", fhirURL.String())
	}

	return body, mimeType, nil
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

func requestWithMimeType(req *http.Request, mimeType string, client *http.Client) (*http.Response, error) {
	req.Header.Set("Accept", mimeType)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "making the GET request to %s failed", req.URL.String())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET request to %s responded with status %s", req.URL.String(), resp.Status)
	}

	return resp, nil
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
		return errors.Wrap(ctx.Err(), "unable to message to queue - context ended")
	default:
		// ok
	}

	err := (*mq).PublishToQueue(*ch, queueName, message)
	if err != nil {
		return err
	}

	return nil
}
