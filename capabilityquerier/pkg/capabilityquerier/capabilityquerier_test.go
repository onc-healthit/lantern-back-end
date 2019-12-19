package capabilityquerier

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"net/http"
	"net/url"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/pkg/errors"
)

var sampleURL = "https://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2/metadata"

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

	resp, err := requestWithMimeType(req, fhir2LessJSONMIMEType, &(tc.Client))
	th.Assert(t, err == nil, err)
	defer resp.Body.Close()

	th.Assert(t, req.Header.Get("Accept") == fhir2LessJSONMIMEType, "request accept header not set to mime type as expected")

	// test http request error

	tc, err = basicTestClient()
	th.Assert(t, err == nil, err)
	tc.Close() // makes request fail

	_, err = requestWithMimeType(req, fhir2LessJSONMIMEType, &(tc.Client))
	switch errors.Cause(err).(type) {
	case *url.Error:
		// expect url.Error because we closed the connection that we're querying.
	default:
		t.Fatal("expected connection error")
	}

	// test http response code error
	tc = th.NewTestClientWith404()
	defer tc.Close()

	_, err = requestWithMimeType(req, fhir2LessJSONMIMEType, &(tc.Client))
	println(err.Error())
	th.Assert(t, err.Error() == fmt.Sprintf("GET request to %s responded with status 404 Not Found", sampleURL), "expected to see an error for 404 response code status")
}

func basicTestClient() (*th.TestClient, error) {
	path := filepath.Join("testdata", "metadata.json")
	okResponse, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tc := th.NewTestClientWithResponse(okResponse)

	return tc, nil
}
