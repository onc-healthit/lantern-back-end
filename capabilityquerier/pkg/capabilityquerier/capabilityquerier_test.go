package capabilityquerier

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"net/http"
	"net/url"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/pkg/errors"

	"github.com/onc-healthit/lantern-back-end/lanternmq/mock"
)

var sampleURL = "https://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2/metadata"

func Test_GetAndSendCapabilityStatement(t *testing.T) {
	var ctx context.Context
	var fhirURL *url.URL
	var tc *th.TestClient
	var message []byte
	var ch lanternmq.ChannelID
	var err error

	mq := mock.NewBasicMockMessageQueue()
	ch = 1
	queueName := "queue name"

	// basic test

	fhirURL = &url.URL{}
	fhirURL, err = fhirURL.Parse(sampleURL)
	th.Assert(t, err == nil, err)
	ctx = context.Background()
	tc, err = testClientWithContentType(fhir2LessJSONMIMEType)
	th.Assert(t, err == nil, err)
	defer tc.Close()

	// create the expected result
	path := filepath.Join("testdata", "metadata.json")
	expectedCapStat, err := ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	expectedMimeType := fhir2LessJSONMIMEType
	expectedTLSVersion := "TLS 1.0"
	expectedMsgStruct := Message{
		URL:        fhirURL.String(),
		MimeType:   expectedMimeType,
		TLSVersion: expectedTLSVersion,
	}
	err = json.Unmarshal(expectedCapStat, &(expectedMsgStruct.CapabilityStatement))
	th.Assert(t, err == nil, err)
	expectedMsg, err := json.Marshal(expectedMsgStruct)
	th.Assert(t, err == nil, err)

	// execute tested function
	err = GetAndSendCapabilityStatement(ctx, fhirURL, &(tc.Client), &mq, &ch, queueName)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(mq.(*mock.BasicMockMessageQueue).Queue) == 1, "expect one message on the queue")
	message = <-mq.(*mock.BasicMockMessageQueue).Queue
	th.Assert(t, bytes.Equal(message, expectedMsg), "expected the capability statement on the queue to be the same as the one sent")

	// context canceled error
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = GetAndSendCapabilityStatement(ctx, fhirURL, &(tc.Client), &mq, &ch, queueName)
	th.Assert(t, errors.Cause(err) == context.Canceled, "expected GetAndSendCapabilityStatement to error out due to context ending")
	th.Assert(t, len(mq.(*mock.BasicMockMessageQueue).Queue) == 0, "expect no messages on the queue")

	// server error response
	ctx = context.Background()

	tc = th.NewTestClientWith404()
	defer tc.Close()

	err = GetAndSendCapabilityStatement(ctx, fhirURL, &(tc.Client), &mq, &ch, queueName)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(mq.(*mock.BasicMockMessageQueue).Queue) == 1, "expect one message on the queue")
	message = <-mq.(*mock.BasicMockMessageQueue).Queue
	var messageStruct Message
	err = json.Unmarshal(message, &messageStruct)
	th.Assert(t, err == nil, err)
	th.Assert(t, messageStruct.Err == fmt.Sprintf("GET request to %s responded with status 404 Not Found", sampleURL), "expected 404 error to be documented in sent error message")
}

func Test_requestCapabilityStatement(t *testing.T) {
	var ctx context.Context
	var fhirURL *url.URL
	var tc *th.TestClient
	var capStat, expectedCapStat []byte
	var path, mimeType, tlsVersion, expectedMimeType, expectedTLSVersion string
	var err error

	// basic test: fhir2LessJSONMIMEType

	path = filepath.Join("testdata", "metadata.json")
	expectedCapStat, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	expectedMimeType = fhir2LessJSONMIMEType
	expectedTLSVersion = "TLS 1.0"

	ctx = context.Background()
	fhirURL = &url.URL{}
	fhirURL, err = fhirURL.Parse(sampleURL)
	th.Assert(t, err == nil, err)
	tc, err = testClientWithContentType(fhir2LessJSONMIMEType)
	th.Assert(t, err == nil, err)
	defer tc.Close()

	capStat, mimeType, tlsVersion, err = requestCapabilityStatement(ctx, fhirURL, &(tc.Client))
	th.Assert(t, err == nil, err)
	th.Assert(t, bytes.Equal(capStat, expectedCapStat), "capability statement did not match expected capability statement")
	th.Assert(t, mimeType == expectedMimeType, fmt.Sprintf("expected mimeType %s; received mimeType %s", expectedMimeType, mimeType))
	th.Assert(t, tlsVersion == expectedTLSVersion, fmt.Sprintf("expected TLS version %s; received TLS version %s", expectedTLSVersion, tlsVersion))

	// basic test: fhir3PlusJSONMIMEType

	path = filepath.Join("testdata", "metadata.json")
	expectedCapStat, err = ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	expectedMimeType = fhir3PlusJSONMIMEType
	expectedTLSVersion = "TLS 1.0"

	ctx = context.Background()
	fhirURL = &url.URL{}
	fhirURL, err = fhirURL.Parse(sampleURL)
	th.Assert(t, err == nil, err)
	tc, err = testClientWithContentType(fhir3PlusJSONMIMEType)
	th.Assert(t, err == nil, err)
	defer tc.Close()

	capStat, mimeType, tlsVersion, err = requestCapabilityStatement(ctx, fhirURL, &(tc.Client))
	th.Assert(t, err == nil, err)
	th.Assert(t, bytes.Equal(capStat, expectedCapStat), "capability statement did not match expected capability statement")
	th.Assert(t, mimeType == expectedMimeType, fmt.Sprintf("expected mimeType %s; received mimeType %s", expectedMimeType, mimeType))
	th.Assert(t, tlsVersion == expectedTLSVersion, fmt.Sprintf("expected TLS version %s; received TLS version %s", expectedTLSVersion, tlsVersion))

	// requestWithMimeType error due to test server closing

	tc, err = basicTestClient()
	th.Assert(t, err == nil, err)
	tc.Close() // makes request fail

	_, _, _, err = requestCapabilityStatement(ctx, fhirURL, &(tc.Client))
	switch errors.Cause(err).(type) {
	case *url.Error:
		// expect url.Error because we closed the connection that we're querying.
	default:
		t.Fatal("expected connection error")
	}

	// mimeType mismatch
	expectedErrStr := fmt.Sprintf("response MIME type (nonesense mimetype; charset=utf-8) does not match JSON request MIME types for FHIR 3+ (%s) or FHIR 2- (%s)",
		fhir3PlusJSONMIMEType, fhir2LessJSONMIMEType)
	tc, err = testClientWithContentType("nonesense mimetype")
	th.Assert(t, err == nil, err)
	defer tc.Close()

	_, _, _, err = requestCapabilityStatement(ctx, fhirURL, &(tc.Client))
	th.Assert(t, err.Error() == expectedErrStr, fmt.Sprintf("expected error '%s'; received error '%s'", expectedErrStr, err.Error()))
}

func Test_getTLSVersion(t *testing.T) {
	var tc *th.TestClient
	var resp *http.Response
	var tlsVersion string
	var expectedTLSVersion string

	req, err := http.NewRequest("GET", sampleURL, nil)
	th.Assert(t, err == nil, err)

	// LDC 12/19/19
	// can't test SSL 3.0/TLS 1.3/Unknown. Go client does not appear to be able to support these
	// values. When setting up the test client with these values, the following exception
	// is thrown: "tls: no supported versions satisfy MinVersion and MaxVersion"

	// TLS 1.0

	expectedTLSVersion = "TLS 1.0"
	tc, err = testClientWithTLSVersion(tls.VersionTLS10)
	th.Assert(t, err == nil, err)
	resp, err = tc.Client.Do(req)
	th.Assert(t, err == nil, err)

	tlsVersion = getTLSVersion(resp)
	th.Assert(t, tlsVersion == expectedTLSVersion, fmt.Sprintf("expected %s; received %s", expectedTLSVersion, tlsVersion))

	// TLS 1.1

	expectedTLSVersion = "TLS 1.1"
	tc, err = testClientWithTLSVersion(tls.VersionTLS11)
	th.Assert(t, err == nil, err)
	resp, err = tc.Client.Do(req)
	th.Assert(t, err == nil, err)

	tlsVersion = getTLSVersion(resp)
	th.Assert(t, tlsVersion == expectedTLSVersion, fmt.Sprintf("expected %s; received %s", expectedTLSVersion, tlsVersion))

	// TLS 1.2

	expectedTLSVersion = "TLS 1.2"
	tc, err = testClientWithTLSVersion(tls.VersionTLS12)
	th.Assert(t, err == nil, err)
	resp, err = tc.Client.Do(req)
	th.Assert(t, err == nil, err)

	tlsVersion = getTLSVersion(resp)
	th.Assert(t, tlsVersion == expectedTLSVersion, fmt.Sprintf("expected %s; received %s", expectedTLSVersion, tlsVersion))
}

func Test_mimeTypesMatch(t *testing.T) {
	var reqMimeType, respMimeType string
	var match bool

	// test success

	reqMimeType = fhir2LessJSONMIMEType
	respMimeType = fmt.Sprintf("%s; charset=utf-8", fhir2LessJSONMIMEType)

	match = mimeTypesMatch(reqMimeType, respMimeType)
	th.Assert(t, match, fmt.Sprintf("expected mime type '%s' to match '%s'", reqMimeType, respMimeType))

	// test fail

	reqMimeType = fhir2LessJSONMIMEType
	respMimeType = fmt.Sprintf("%s; charset=utf-8", fhir3PlusJSONMIMEType)

	match = mimeTypesMatch(reqMimeType, respMimeType)
	th.Assert(t, !match, fmt.Sprintf("did not expect mime type '%s' to match '%s'", reqMimeType, respMimeType))

	// test empty resp

	reqMimeType = fhir2LessJSONMIMEType
	respMimeType = ""

	match = mimeTypesMatch(reqMimeType, respMimeType)
	th.Assert(t, !match, fmt.Sprintf("did not expect mime type '%s' to match '%s'", reqMimeType, respMimeType))

	// test empty req

	reqMimeType = ""
	respMimeType = fmt.Sprintf("%s; charset=utf-8", fhir3PlusJSONMIMEType)

	match = mimeTypesMatch(reqMimeType, respMimeType)
	th.Assert(t, !match, fmt.Sprintf("did not expect mime type '%s' to match '%s'", reqMimeType, respMimeType))
}

func Test_requestWithMimeType(t *testing.T) {
	req, err := http.NewRequest("GET", sampleURL, nil)
	th.Assert(t, err == nil, err)

	// basic test

	tc, err := basicTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()

	resp, is406, err := requestWithMimeType(req, fhir2LessJSONMIMEType, &(tc.Client))
	th.Assert(t, err == nil, err)
	defer resp.Body.Close()

	th.Assert(t, !is406, "did not expect a 406 response")
	th.Assert(t, req.Header.Get("Accept") == fhir2LessJSONMIMEType, "request accept header not set to mime type as expected")

	// test http request error

	tc, err = basicTestClient()
	th.Assert(t, err == nil, err)
	tc.Close() // makes request fail

	_, is406, err = requestWithMimeType(req, fhir2LessJSONMIMEType, &(tc.Client))
	switch errors.Cause(err).(type) {
	case *url.Error:
		// expect url.Error because we closed the connection that we're querying.
	default:
		t.Fatal("expected connection error")
	}
	th.Assert(t, !is406, "did not expect a 406 response")

	// test http response code error
	tc = th.NewTestClientWith404()
	defer tc.Close()

	_, is406, err = requestWithMimeType(req, fhir2LessJSONMIMEType, &(tc.Client))
	th.Assert(t, err.Error() == fmt.Sprintf("GET request to %s responded with status 404 Not Found", sampleURL), "expected to see an error for 404 response code status")
	th.Assert(t, !is406, "did not expect a 406 response")

	// test 406 response
	tc = th.NewTestClientWith406()
	defer tc.Close()

	_, is406, err = requestWithMimeType(req, fhir2LessJSONMIMEType, &(tc.Client))
	th.Assert(t, err == nil, err)
	th.Assert(t, is406, "expected a 406 response")
}

func Test_sendToQueue(t *testing.T) {
	var ch lanternmq.ChannelID
	var ctx context.Context
	var err error

	message := "this is a message"
	mq := mock.NewBasicMockMessageQueue()
	ch = 1
	queueName := "queue name"

	// basic test

	ctx = context.Background()

	err = sendToQueue(ctx, message, &mq, &ch, queueName)
	th.Assert(t, err == nil, err)

	th.Assert(t, len(mq.(*mock.BasicMockMessageQueue).Queue) == 1, "expected a message to be in the queue")

	bRcvMsg := <-mq.(*mock.BasicMockMessageQueue).Queue
	rcvMsg := string(bRcvMsg)
	th.Assert(t, rcvMsg == message, "expected the recieved message to be the same as the sent message.")

	// test context ends
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = sendToQueue(ctx, message, &mq, &ch, queueName)
	th.Assert(t, errors.Cause(err) == context.Canceled, "expected persistProducts to error out due to context ending")
}

func basicTestClient() (*th.TestClient, error) {
	return testClientWithContentType(fhir2LessJSONMIMEType)
}

func testClientWithContentType(contentType string) (*th.TestClient, error) {
	path := filepath.Join("testdata", "metadata.json")
	okResponse, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType+"; charset=utf-8")
		_, _ = w.Write(okResponse)
	})

	tc := th.NewTestClient(h)

	return tc, nil
}

func testClientWithTLSVersion(tlsVersion uint16) (*th.TestClient, error) {
	tc, err := basicTestClient()
	if err != nil {
		return nil, err
	}

	transport := tc.Client.Transport.(*http.Transport)
	transport.TLSClientConfig.MaxVersion = tlsVersion
	transport.TLSClientConfig.MinVersion = tlsVersion

	return tc, nil
}
