package capabilityhandler

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

var testQueueMsg = map[string]interface{}{
	"url":               "http://example.com/DTSU2/",
	"err":               "",
	"mimeTypes":         []string{"application/json+fhir"},
	"httpResponse":      200,
	"tlsVersion":        "TLS 1.2",
	"smarthttpResponse": 0,
	"smartResp":         nil,
}

var testValidationObj = endpointmanager.Validation{
	Errors: []endpointmanager.Rule{
		endpointmanager.Rule{
			RuleName:  endpointmanager.R4MimeTypeRule,
			Expected:  "application/fhir+json",
			Comment:   "The formal MIME-type for FHIR resources is application/fhir+json for FHIR version STU3 and above. The correct mime type SHALL be used by clients and servers.",
			Reference: "http://hl7.org/fhir/http.html",
		},
	},
	Warnings: []endpointmanager.Rule{},
}

var testFhirEndpointInfo = endpointmanager.FHIREndpointInfo{
	URL:               "http://example.com/DTSU2/",
	MIMETypes:         []string{"application/json+fhir"},
	TLSVersion:        "TLS 1.2",
	HTTPResponse:      200,
	Errors:            "",
	SMARTHTTPResponse: 0,
	SMARTResponse:     nil,
	Validation:        testValidationObj,
}

// Convert the test Queue Message into []byte format for testing purposes
func convertInterfaceToBytes(message map[string]interface{}) ([]byte, error) {
	returnMsg, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return returnMsg, nil
}

func setupCapabilityStatement(t *testing.T) {
	// capability statement
	path := filepath.Join("../testdata", "cerner_capability_dstu2.json")
	csJSON, err := ioutil.ReadFile(path)
	th.Assert(t, err == nil, err)
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	th.Assert(t, err == nil, err)
	testFhirEndpointInfo.CapabilityStatement = cs
	var capStat map[string]interface{}
	err = json.Unmarshal(csJSON, &capStat)
	th.Assert(t, err == nil, err)
	testQueueMsg["capabilityStatement"] = capStat
}

func Test_formatMessage(t *testing.T) {
	setupCapabilityStatement(t)
	expectedEndpt := testFhirEndpointInfo
	tmpMessage := testQueueMsg

	message, err := convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)

	// basic test
	endpt, returnErr := formatMessage(message)
	th.Assert(t, returnErr == nil, returnErr)
	th.Assert(t, expectedEndpt.Equal(endpt), "An error was thrown because the endpoints are not equal")

	// should not throw error if metadata is not in the URL
	tmpMessage["url"] = "http://example.com/DTSU2/"
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr == nil, "An error was thrown because metadata was not included in the url")

	// test incorrect error message
	tmpMessage["err"] = nil
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to an incorrect error message")
	tmpMessage["err"] = ""

	// test incorrect URL
	tmpMessage["url"] = nil
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to an incorrect URL")

	tmpMessage["url"] = "http://example.com/DTSU2/"
	// test incorrect TLS Version
	tmpMessage["tlsVersion"] = 1
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to an incorrect TLS Version")
	tmpMessage["tlsVersion"] = "TLS 1.2"

	// test incorrect MIME Type
	tmpMessage["mimeTypes"] = 1
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to incorrect MIME Types")
	tmpMessage["mimeTypes"] = []string{"application/json+fhir"}

	// test incorrect http response
	tmpMessage["httpResponse"] = "200"
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to an incorrect HTTP response")
	tmpMessage["httpResponse"] = 200

	// test incorrect http response
	tmpMessage["smarthttpResponse"] = "200"
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to an incorrect smart HTTP response")
	tmpMessage["smarthttpResponse"] = 200
}
