package capabilityquerier

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/pkg/errors"
)

func GetAndSendCapabilityStatement(
	ctx context.Context,
	fhirURL *url.URL,
	client *http.Client,
	mq lanternmq.MessageQueue,
	ch lanternmq.ChannelID,
	queueName string) error {
	var err error

	capResp, err := requestCapabilityStatement(ctx, fhirURL, client)
	if err != nil {
		return errors.Wrap(err, "error requesting capability statement from "+fhirURL.String())
	}

	err = sendToQueue(ctx, capResp, mq, ch, queueName)
	if err != nil {
		return errors.Wrapf(err, "error sending capability statement to queue '%s' over channel '%v'", queueName, ch)
	}

	return err
}

func requestCapabilityStatement(ctx context.Context, fhirURL *url.URL, client *http.Client) (string, error) {
	var err error

	req, err := http.NewRequest("GET", fhirURL.String(), nil)
	if err != nil {
		return "", errors.Wrap(err, "unable to create new GET request from URL: "+fhirURL.String())
	}
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrapf(err, "making the GET request to %s failed", fhirURL.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET request to %s responded with status %s", fhirURL.String(), resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "reading the response from %s failed", fhirURL.String())
	}

	return string(body), nil
}

func sendToQueue(
	ctx context.Context,
	message string,
	mq lanternmq.MessageQueue,
	ch lanternmq.ChannelID,
	queueName string) error {

	// don't send the message if the context is done
	select {
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "unable to message to queue - context ended")
	default:
		// ok
	}

	err := mq.PublishToQueue(ch, queueName, message)
	if err != nil {
		return err
	}

	return nil
}
