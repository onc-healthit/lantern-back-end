package endpointmanager

import (
	"testing"

	_ "github.com/lib/pq"
)

func Test_FHIREndpointMetadataEqual(t *testing.T) {

	var endpointMetadata1 = &FHIREndpointMetadata{
		ID:                   1,
		URL:                  "http://www.example.com",
		HTTPResponse:         200,
		Availability:         1.0,
		Errors:               "Example Error",
		ResponseTime:         0.123456,
		SMARTHTTPResponse:    200,
		RequestedFhirVersion: "None",
	}

	var endpointMetadata2 = &FHIREndpointMetadata{
		ID:                   1,
		URL:                  "http://www.example.com",
		HTTPResponse:         200,
		Availability:         1.0,
		Errors:               "Example Error",
		ResponseTime:         0.123456,
		SMARTHTTPResponse:    200,
		RequestedFhirVersion: "None",
	}

	if !endpointMetadata1.Equal(endpointMetadata2) {
		t.Errorf("Expected endpointMetadata1 to equal endpointMetadata2. They are not equal.")
	}

	endpointMetadata2.ID = 2
	if !endpointMetadata1.Equal(endpointMetadata2) {
		t.Errorf("Expect endpointMetadata1 to equal endpointMetadata2. ids should be ignored. %d vs %d", endpointMetadata1.ID, endpointMetadata2.ID)
	}
	endpointMetadata2.ID = endpointMetadata1.ID

	endpointMetadata2.URL = "other"
	if endpointMetadata1.Equal(endpointMetadata2) {
		t.Errorf("Expect endpointMetadata1 to not equal endpointMetadata2. URL should be different. %s vs %s", endpointMetadata1.URL, endpointMetadata2.URL)
	}
	endpointMetadata2.URL = endpointMetadata1.URL

	endpointMetadata2.HTTPResponse = 404
	if endpointMetadata1.Equal(endpointMetadata2) {
		t.Errorf("Expect endpointMetadata1 to not equal endpointMetadata2. HTTP responses should be different. %d vs %d", endpointMetadata1.HTTPResponse, endpointMetadata2.HTTPResponse)
	}
	endpointMetadata2.HTTPResponse = endpointMetadata1.HTTPResponse

	endpointMetadata2.SMARTHTTPResponse = 0
	if endpointMetadata1.Equal(endpointMetadata2) {
		t.Errorf("Expect endpointMetadata1 to not equal endpointMetadata2. Smart HTTP responses should be different. %d vs %d", endpointMetadata1.SMARTHTTPResponse, endpointMetadata2.SMARTHTTPResponse)
	}
	endpointMetadata2.SMARTHTTPResponse = endpointMetadata1.SMARTHTTPResponse

	endpointMetadata2.Availability = 0
	if endpointMetadata1.Equal(endpointMetadata2) {
		t.Errorf("Did not expect endpointMetadata1 to equal endpointMetadata2. Availability should be different. %f vs %f", endpointMetadata1.Availability, endpointMetadata2.Availability)
	}
	endpointMetadata2.Availability = endpointMetadata1.Availability

	endpointMetadata2.Errors = "other"
	if endpointMetadata1.Equal(endpointMetadata2) {
		t.Errorf("Did not expect endpointMetadata1 to equal endpointMetadata2. Errors should be different. %s vs %s", endpointMetadata1.Errors, endpointMetadata2.Errors)
	}
	endpointMetadata2.Errors = endpointMetadata1.Errors

	endpointMetadata2.ResponseTime = 0.234567
	if endpointMetadata1.Equal(endpointMetadata2) {
		t.Errorf("Did not expect endpointMetadata1 to equal endpointMetadata2. ResponseTime should be different. %f vs %f", endpointMetadata1.ResponseTime, endpointMetadata2.ResponseTime)
	}
	endpointMetadata2.ResponseTime = endpointMetadata1.ResponseTime

	endpointMetadata2 = nil
	if endpointMetadata1.Equal(endpointMetadata2) {
		t.Errorf("Did not expect endpointMetadata1 to equal nil endpointMetadata2.")
	}
	endpointMetadata2 = endpointMetadata1

	endpointMetadata1 = nil
	if endpointMetadata1.Equal(endpointMetadata2) {
		t.Errorf("Did not expect nil endpointMetadata1 to equal endpointMetadata2")
	}

	endpointMetadata2 = nil
	if !endpointMetadata1.Equal(endpointMetadata2) {
		t.Errorf("Nil endpointMetadata1 should equal nil endpointMetadata2.")
	}
}
