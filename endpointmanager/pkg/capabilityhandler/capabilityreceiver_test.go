package capabilityhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/pkg/errors"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/mock"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)

type testCapStatement struct {
	Test1 string
	Test2 string
}

var testQueueMsg = map[string]interface{}{
	"url":        "http://example.com/DTSU2/metadata",
	"err":        "",
	"mimetype":   "application/json+fhir",
	"tlsVersion": "TLS 1.2",
	"capabilityStatement": testCapStatement{
		Test1: "TestValue1",
		Test2: "TestValue2",
	},
}

var testFhirEndpoint = endpointmanager.FHIREndpoint{
	URL:        "http://example.com/DTSU2/",
	MimeType:   "application/json+fhir",
	TLSVersion: "TLS 1.2",
	Errors:     "",
	CapabilityStatement: testCapStatement{
		Test1: "TestValue1",
		Test2: "TestValue2",
	},
}

// Convert the test Queue Message into []byte format for testing purposes
func convertInterfaceToBytes(message map[string]interface{}) ([]byte, error) {
	returnMsg, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return returnMsg, nil
}

func Test_formatMessage(t *testing.T) {
	expectedEndpt := testFhirEndpoint
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
	endpt, returnErr = formatMessage(message)
	th.Assert(t, returnErr == nil, "An error was thrown because metadata was not included in the url")
	th.Assert(t, expectedEndpt.URL == endpt.URL, fmt.Sprintf("%s and %s are not equal", expectedEndpt.URL, endpt.URL))

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
	tmpMessage["url"] = "http://example.com/DTSU2/metadata"

	// test incorrect TLS Version
	tmpMessage["tlsVersion"] = 1
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to an incorrect TLS Version")
	tmpMessage["tlsVersion"] = "TLS 1.2"

	// test incorrect MIME Type
	tmpMessage["mimetype"] = 1
	message, err = convertInterfaceToBytes(tmpMessage)
	th.Assert(t, err == nil, err)
	_, returnErr = formatMessage(message)
	th.Assert(t, returnErr != nil, "Expected an error to be thrown due to an incorrect MIME Type")
	tmpMessage["mimetype"] = "application/json+fhir"
}

func Test_saveMsgInDB(t *testing.T) {
	store := mock.NewBasicMockFhirEndpointStore()

	args := make(map[string]interface{})
	args["store"] = store

	expectedEndpt := testFhirEndpoint
	queueTmp := testQueueMsg

	queueMsg, err := convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)

	// check that nothing is stored and that saveMsgInDB throws an error if the context is canceled
	testCtx, cancel := context.WithCancel(context.Background())
	args["ctx"] = testCtx
	cancel()
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, len(store.(*mock.BasicMockStore).FhirEndpointData) == 0, "should not have stored data")
	th.Assert(t, errors.Cause(err) == context.Canceled, "should have errored out with root cause that the context was canceled")

	// reset context
	args["ctx"] = context.Background()

	// check that new item is stored
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(store.(*mock.BasicMockStore).FhirEndpointData) == 1, "did not store data as expected")
	th.Assert(t, expectedEndpt.Equal(store.(*mock.BasicMockStore).FhirEndpointData[0]), "stored data does not equal expected store data")

	// check that a second new item is stored
	queueTmp["url"] = "https://test-two.com"
	expectedEndpt.URL = "https://test-two.com"
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(store.(*mock.BasicMockStore).FhirEndpointData) == 2, "there should be two endpoints in the database")
	th.Assert(t, expectedEndpt.Equal(store.(*mock.BasicMockStore).FhirEndpointData[1]), "the second endpoint data does not equal expected store data")
	expectedEndpt = testFhirEndpoint
	queueTmp["url"] = "http://example.com/DTSU2/metadata"

	// check that an item with the same URL updates the endpoint in the database
	queueTmp["tlsVersion"] = "TLS 1.3"
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(store.(*mock.BasicMockStore).FhirEndpointData) == 2, "did not store data as expected")
	th.Assert(t, store.(*mock.BasicMockStore).FhirEndpointData[0].TLSVersion == "TLS 1.3", "The TLS Version was not updated")

	// check that error adding to store throws error
	queueTmp["url"] = "https://a-new-url.com"
	queueMsg, err = convertInterfaceToBytes(queueTmp)
	th.Assert(t, err == nil, err)
	addFn := store.(*mock.BasicMockStore).AddFHIREndpointFn
	store.(*mock.BasicMockStore).AddFHIREndpointFn = func(_ context.Context, _ *endpointmanager.FHIREndpoint) error {
		return errors.New("add fhir endpoint test error")
	}
	err = saveMsgInDB(queueMsg, &args)
	th.Assert(t, errors.Cause(err).Error() == "add fhir endpoint test error", "expected error adding product")
	store.(*mock.BasicMockStore).AddFHIREndpointFn = addFn

}
