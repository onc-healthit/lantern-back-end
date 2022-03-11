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
	"time"

	"net/http"
	"net/url"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/pkg/errors"
)

var sampleURL = "https://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2/"
var sampleURLNoTLS = "http://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2/"

func Test_requestCapabilityStatementAndSmartOnFhir(t *testing.T) {
	var ctx context.Context
	var tc *th.TestClient
	var capStat, expectedCapStat []byte
	var expectedMimeType, expectedTLSVersion string
	var err error
	var message Message
	var smartResp []byte
	var expectedSmartResp = []byte("null")

	// basic test: fhir2LessJSONMIMEType

	message = Message{}
	message.RequestedFhirVersion = "None"

	expectedCapStat, err = capabilityStatement()
	th.Assert(t, err == nil, err)
	expectedMimeType = fhir2LessJSONMIMEType
	expectedTLSVersion = "TLS 1.0"

	ctx = context.Background()
	metadataURL := "https://webhook.site/6f884d36-eaf7-446f-a3d7-c306389351c6"
	th.Assert(t, err == nil, err)
	tc, err = testClientWithContentType(fhir2LessJSONMIMEType)
	th.Assert(t, err == nil, err)
	defer tc.Close()

	err = requestCapabilityStatementAndSmartOnFhir(ctx, metadataURL, "metadata", &(tc.Client), "", &message)
	th.Assert(t, err == nil, err)
	capStat, err = json.Marshal(message.CapabilityStatement)
	th.Assert(t, err == nil, err)
	th.Assert(t, bytes.Equal(capStat, expectedCapStat), "capability statement did not match expected capability statement")
	th.Assert(t, len(message.MIMETypes) == 1, fmt.Sprintf("expected one matched mime type. Got %d, %+v", len(message.MIMETypes), message.MIMETypes))
	th.Assert(t, message.MIMETypes[0] == expectedMimeType, fmt.Sprintf("expected mimeType %s; received mimeTypes %s", expectedMimeType, message.MIMETypes[0]))
	th.Assert(t, message.TLSVersion == expectedTLSVersion, fmt.Sprintf("expected TLS version %s; received TLS version %s", expectedTLSVersion, message.TLSVersion))

	client := &http.Client{
		Timeout: time.Second * 35,
	}

	// check that response from well known endpt is null
	wellKnownURL := endpointmanager.NormalizeWellKnownURL(sampleURL)
	err = requestCapabilityStatementAndSmartOnFhir(ctx, wellKnownURL, "well-known", client, "", &message)
	th.Assert(t, err == nil, err)
	smartResp, err = json.Marshal(message.SMARTResp)
	th.Assert(t, err == nil, err)
	th.Assert(t, bytes.Equal(smartResp, expectedSmartResp), "response from well known endpt did not match expected response")
	th.Assert(t, len(message.MIMETypes) == 0, fmt.Sprintf("expected no matched mime types. Got %d.", len(message.MIMETypes)))

	// basic test: fhir3PlusJSONMIMEType

	message = Message{}
	message.RequestedFhirVersion = "None"
	// Add fhir3PlusJSONMIMEType MIME type to message so that it tries this saved MIME type first
	message.MIMETypes = []string{fhir3PlusJSONMIMEType}

	expectedCapStat, err = capabilityStatement()
	th.Assert(t, err == nil, err)
	expectedMimeType = fhir3PlusJSONMIMEType
	expectedTLSVersion = "TLS 1.0"

	ctx = context.Background()
	th.Assert(t, err == nil, err)
	tc, err = testClientWithContentType(fhir3PlusJSONMIMEType)
	th.Assert(t, err == nil, err)
	defer tc.Close()

	err = requestCapabilityStatementAndSmartOnFhir(ctx, metadataURL, "metadata", &(tc.Client), "", &message)
	th.Assert(t, err == nil, err)
	capStat, err = json.Marshal(message.CapabilityStatement)
	th.Assert(t, err == nil, err)
	th.Assert(t, bytes.Equal(capStat, expectedCapStat), "capability statement did not match expected capability statement")
	th.Assert(t, len(message.MIMETypes) == 1, fmt.Sprintf("expected one matched mime type. Got %d.", len(message.MIMETypes)))
	th.Assert(t, message.MIMETypes[0] == expectedMimeType, fmt.Sprintf("expected mimeType %s; received mimeTypes %s", expectedMimeType, message.MIMETypes[0]))
	th.Assert(t, message.TLSVersion == expectedTLSVersion, fmt.Sprintf("expected TLS version %s; received TLS version %s", expectedTLSVersion, message.TLSVersion))

	// requestWithMimeType error due to test server closing

	message = Message{}

	tc, err = basicTestClient()
	th.Assert(t, err == nil, err)
	tc.Close() // makes request fail

	err = requestCapabilityStatementAndSmartOnFhir(ctx, metadataURL, "metadata", &(tc.Client), "", &message)
	switch errors.Cause(err).(type) {
	case *url.Error:
		// expect url.Error because we closed the connection that we're querying.
	default:
		t.Fatal("expected connection error")
	}

	// mimeType mismatch

	message = Message{}

	tc, err = testClientWithContentType("nonesense mimetype")
	th.Assert(t, err == nil, err)
	defer tc.Close()

	err = requestCapabilityStatementAndSmartOnFhir(ctx, metadataURL, "metadata", &(tc.Client), "", &message)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(message.MIMETypes) == 0, "expected no matched mime types")

	// test with fhir3PlusJSONMIMEType already saved
	message = Message{}
	message.MIMETypes = []string{fhir3PlusJSONMIMEType}
	expectedCapStat, err = capabilityStatement()
	th.Assert(t, err == nil, err)
	expectedMimeType = fhir3PlusJSONMIMEType
	expectedTLSVersion = "TLS 1.0"

	ctx = context.Background()
	tc, err = testClientWithContentType(fhir3PlusJSONMIMEType)
	th.Assert(t, err == nil, err)
	defer tc.Close()

	err = requestCapabilityStatementAndSmartOnFhir(ctx, metadataURL, "metadata", &(tc.Client), "", &message)
	th.Assert(t, err == nil, err)
	capStat, err = json.Marshal(message.CapabilityStatement)
	th.Assert(t, err == nil, err)
	th.Assert(t, bytes.Equal(capStat, expectedCapStat), "capability statement did not match expected capability statement")
	th.Assert(t, len(message.MIMETypes) == 1, fmt.Sprintf("expected one matched mime type. Got %d.", len(message.MIMETypes)))
	th.Assert(t, message.MIMETypes[0] == expectedMimeType, fmt.Sprintf("expected mimeType %s; received mimeType %s", expectedMimeType, message.MIMETypes[0]))
	th.Assert(t, message.TLSVersion == expectedTLSVersion, fmt.Sprintf("expected TLS version %s; received TLS version %s", expectedTLSVersion, message.TLSVersion))

	// test with fhir2LessJSONMIMEType already saved
	message = Message{}
	message.MIMETypes = []string{fhir2LessJSONMIMEType}
	expectedCapStat, err = capabilityStatement()
	th.Assert(t, err == nil, err)
	expectedMimeType = fhir2LessJSONMIMEType
	expectedTLSVersion = "TLS 1.0"

	ctx = context.Background()
	tc, err = testClientWithContentType(fhir2LessJSONMIMEType)
	th.Assert(t, err == nil, err)
	defer tc.Close()

	err = requestCapabilityStatementAndSmartOnFhir(ctx, metadataURL, "metadata", &(tc.Client), "", &message)
	th.Assert(t, err == nil, err)
	capStat, err = json.Marshal(message.CapabilityStatement)
	th.Assert(t, err == nil, err)
	th.Assert(t, bytes.Equal(capStat, expectedCapStat), "capability statement did not match expected capability statement")
	th.Assert(t, len(message.MIMETypes) == 1, fmt.Sprintf("expected one matched mime type. Got %d.", len(message.MIMETypes)))
	th.Assert(t, message.MIMETypes[0] == expectedMimeType, fmt.Sprintf("expected mimeType %s; received mimeType %s", expectedMimeType, message.MIMETypes[0]))
	th.Assert(t, message.TLSVersion == expectedTLSVersion, fmt.Sprintf("expected TLS version %s; received TLS version %s", expectedTLSVersion, message.TLSVersion))

	// test situation where fhir2LessJSONMIMEType is saved but only fhir3PlusJSONMIMEType works
	message = Message{}
	message.RequestedFhirVersion = "None"
	message.MIMETypes = []string{fhir2LessJSONMIMEType}
	expectedCapStat, err = capabilityStatement()
	th.Assert(t, err == nil, err)
	expectedMimeType = fhir3PlusJSONMIMEType
	expectedTLSVersion = "TLS 1.0"

	ctx = context.Background()
	tc, err = testClientOnlyAcceptGivenType(fhir3PlusJSONMIMEType)
	th.Assert(t, err == nil, err)
	defer tc.Close()

	err = requestCapabilityStatementAndSmartOnFhir(ctx, metadataURL, "metadata", &(tc.Client), "", &message)
	th.Assert(t, err == nil, err)
	capStat, err = json.Marshal(message.CapabilityStatement)
	th.Assert(t, err == nil, err)
	th.Assert(t, bytes.Equal(capStat, expectedCapStat), "capability statement did not match expected capability statement")
	th.Assert(t, len(message.MIMETypes) == 1, fmt.Sprintf("expected one matched mime type. Got %d.", len(message.MIMETypes)))
	th.Assert(t, message.MIMETypes[0] == expectedMimeType, fmt.Sprintf("mismatched: expected mimeType %s; received mimeType %s", expectedMimeType, message.MIMETypes[0]))
	th.Assert(t, message.TLSVersion == expectedTLSVersion, fmt.Sprintf("expected TLS version %s; received TLS version %s", expectedTLSVersion, message.TLSVersion))

	// test with two mime types and both saved ones don't work
	message = Message{}
	expectedMimeType = fhir2LessJSONMIMEType
	message.RequestedFhirVersion = "None"
	message.MIMETypes = []string{"nonsense mimetype", "nonsense mimetype2"}
	tc, err = basicTestClient()
	th.Assert(t, err == nil, err)
	defer tc.Close()
	ctx = context.Background()

	err = requestCapabilityStatementAndSmartOnFhir(ctx, metadataURL, "metadata", &(tc.Client), "", &message)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(message.MIMETypes) == 1, fmt.Sprintf("expected one matched mime types, got %d", len(message.MIMETypes)))
	th.Assert(t, message.MIMETypes[0] == expectedMimeType, fmt.Sprintf("mismatched: expected mimeType %s; received mimeType %s", expectedMimeType, message.MIMETypes[0]))

	// Can't test with two mime types and only one works because the first one tested is chosen randomly
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

	// No TLS

	expectedTLSVersion = "No TLS"

	req, err = http.NewRequest("GET", sampleURLNoTLS, nil)
	th.Assert(t, err == nil, err)

	tc, err = testClientWithNoTLS()
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

	httpCode, tlsVersion, mimeMatch, capStat, _, err := requestWithMimeType(req, fhir2LessJSONMIMEType, &(tc.Client))
	th.Assert(t, err == nil, err)
	th.Assert(t, httpCode == 200, "expected 200 response")
	th.Assert(t, tlsVersion == "TLS 1.0", fmt.Sprintf("expected TLS 1.0. got %s", tlsVersion))
	th.Assert(t, mimeMatch, "expected the mime types to match")
	th.Assert(t, capStat != nil, "expected to receive a capability statement")

	// test http request error

	tc, err = basicTestClient()
	th.Assert(t, err == nil, err)
	tc.Close() // makes request fail

	_, _, _, _, _, err = requestWithMimeType(req, fhir2LessJSONMIMEType, &(tc.Client))
	switch errors.Cause(err).(type) {
	case *url.Error:
		// expect url.Error because we closed the connection that we're querying.
	default:
		t.Fatal("expected connection error")
	}

	// test http response code error
	tc = th.NewTestClientWith404()
	defer tc.Close()

	httpCode, _, _, _, _, err = requestWithMimeType(req, fhir2LessJSONMIMEType, &(tc.Client))
	th.Assert(t, err == nil, err)
	th.Assert(t, httpCode == 404, fmt.Sprintf("expected 404 response code. Got %d", httpCode))
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

		if r.Header.Get("Accept") != fhir2LessJSONMIMEType && r.Header.Get("Accept") != fhir3PlusJSONMIMEType {
			http.Error(w, "sample 406 error", http.StatusNotAcceptable)
		} else {
			_, _ = w.Write(okResponse)
		}
	})

	tc := th.NewTestClient(h)

	return tc, nil
}

func testClientOnlyAcceptGivenType(contentType string) (*th.TestClient, error) {
	path := filepath.Join("testdata", "metadata.json")
	okResponse, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType+"; charset=utf-8")

		if r.Header.Get("Accept") != contentType {
			http.Error(w, "sample 406 error", http.StatusNotAcceptable)
		} else {
			_, _ = w.Write(okResponse)
		}
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

func testClientWithNoTLS() (*th.TestClient, error) {

	path := filepath.Join("testdata", "metadata.json")
	okResponse, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(okResponse)
	})

	tc := th.NewTestClientNoTLS(h)

	return tc, nil
}

func capabilityStatement() ([]byte, error) {
	path := filepath.Join("testdata", "metadata.json")
	expectedCapStat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var capStatInt interface{}

	err = json.Unmarshal(expectedCapStat, &capStatInt)
	if err != nil {
		return nil, err
	}

	expectedCapStat, err = json.Marshal(capStatInt)
	if err != nil {
		return nil, err
	}

	return expectedCapStat, nil
}
