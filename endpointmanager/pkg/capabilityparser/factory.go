package capabilityparser

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/pkg/errors"
)

// from https://www.hl7.org/fhir/codesystem-FHIR-version.html
// looking at official and release versions only
var dstu2 = []string{"0.4.0", "0.5.0", "1.0.0", "1.0.1", "1.0.2"}
var stu3 = []string{"1.1.0", "1.2.0", "1.4.0", "1.6.0", "1.8.0", "3.0.0", "3.0.1", "3.0.2"}
var r4 = []string{"3.2.0", "3.3.0", "3.5.0", "3.5a.0", "4.0.0", "4.0.1"}

// CapabilityStatement provides access to key fields of the capability statement. It wraps the capability statements
// so users don't need to worry about the capability statement version.
type CapabilityStatement interface {
	GetPublisher() (string, error)
	GetFHIRVersion() (string, error)
	GetSoftware() (map[string]interface{}, error)
	GetSoftwareName() (string, error)
	GetSoftwareVersion() (string, error)
	GetCopyright() (string, error)
	GetRest() ([]map[string]interface{}, error)
	GetResourceList(map[string]interface{}) ([]map[string]interface{}, error)
	GetKind() (string, error)
	GetImplementation() (map[string]interface{}, error)
	GetMessaging() ([]map[string]interface{}, error)
	GetMessagingEndpoint(map[string]interface{}) ([]map[string]interface{}, error)
	GetDocument() ([]map[string]interface{}, error)
	GetDescription() (string, error)

	Equal(CapabilityStatement) bool
	EqualIgnore(CapabilityStatement) bool
	GetJSON() ([]byte, error)
}

type SMARTResponse interface {
	Equal(SMARTResponse) bool
	EqualIgnore(SMARTResponse) bool
	GetJSON() ([]byte, error)
}
type Response struct {
	resp map[string]interface{}
}

type ResponseBody struct {
	Response
}

func NewResponseBody(response map[string]interface{}) *ResponseBody {
	return &ResponseBody{
		Response: Response{
			resp: response,
		},
	}
}

func NewSMARTRespFromInterface(response map[string]interface{}) SMARTResponse {
	if response == nil {
		return nil
	}
	return NewResponseBody(response)
}

func NewSMARTResp(respJSON []byte) (SMARTResponse, error) {
	var err error
	var respMsg map[string]interface{}

	if len(respJSON) == 0 {
		return nil, nil
	}

	err = json.Unmarshal(respJSON, &respMsg)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling JSON response from well known endpoint")
	}

	return NewSMARTRespFromInterface(respMsg), nil
}

// Equal checks if the smart response body is equal to the given smart response body.
func (resp *Response) Equal(resp2 SMARTResponse) bool {
	if resp2 == nil {
		return false
	}

	j1, err := resp.GetJSON()
	if err != nil {
		return false
	}
	j2, err := resp2.GetJSON()
	if err != nil {
		return false
	}
	if !bytes.Equal(j1, j2) {
		return false
	}

	return true
}

// EqualIgnore checks if the smart response body is equal to the given smart response body while ignoring certain fields that may differ.
func (resp *Response) EqualIgnore(resp2 SMARTResponse) bool {
	ignoredFields := []string{}

	if resp2 == nil {
		return false
	}

	var respCopy SMARTResponse
	var resp2Copy SMARTResponse

	respCopy = resp
	resp2Copy = resp2

	var err error

	for _, field := range ignoredFields {
		respCopy, err = deleteFieldFromSmartResponse(respCopy, field)
		if err != nil {
			return false
		}
		resp2Copy, err = deleteFieldFromSmartResponse(resp2Copy, field)
		if err != nil {
			return false
		}
	}

	j1, err := respCopy.GetJSON()
	if err != nil {
		return false
	}
	j2, err := resp2Copy.GetJSON()
	if err != nil {
		return false
	}
	if !bytes.Equal(j1, j2) {
		return false
	}

	return true
}

// GetJSON returns the JSON representation of the capability statement.
func (resp *Response) GetJSON() ([]byte, error) {
	return json.Marshal(resp.resp)
}

// NewCapabilityStatement is a factory method for creating a CapabilityStatement. It determines what version
// the capability statement JSON is and creates the relevant implementation of the CapabilityStatement interface.
func NewCapabilityStatement(capJSON []byte) (CapabilityStatement, error) {
	var err error
	var capStat map[string]interface{}

	if len(capJSON) == 0 {
		return nil, nil
	}

	err = json.Unmarshal(capJSON, &capStat)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling JSON capability statement")
	}

	return NewCapabilityStatementFromInterface(capStat)
}

// NewCapabilityStatementFromInterface is a factory method for creating a CapabilityStatement. It determines what version
// the capability statement JSON map[string]interface{} object is and creates the relevant implementation of the
// CapabilityStatement interface.
func NewCapabilityStatementFromInterface(capStat map[string]interface{}) (CapabilityStatement, error) {
	// return nil if an empty capability statement was passed in
	if capStat == nil {
		return nil, nil
	}

	// DSTU2, STU3, R4 all have fhirVersion in same location
	fhirVersion, ok := capStat["fhirVersion"].(string)
	if !ok {
		return nil, errors.New("unable to parse fhir version from capability/conformance statement")
	}

	if helpers.StringArrayContains(dstu2, fhirVersion) {
		return newDSTU2(capStat), nil
	} else if helpers.StringArrayContains(stu3, fhirVersion) {
		return newSTU3(capStat), nil
	} else if helpers.StringArrayContains(r4, fhirVersion) {
		return newR4(capStat), nil
	}

	return nil, fmt.Errorf("unknown FHIR version %s", fhirVersion)
}

func getRespFormats(resp SMARTResponse) (map[string]interface{}, []byte, error) {
	var respInt map[string]interface{}

	respJSON, err := resp.GetJSON()
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal(respJSON, &respInt)
	if err != nil {
		return nil, nil, err
	}

	return respInt, respJSON, nil
}

func deleteFieldFromSmartResponse(resp SMARTResponse, field string) (SMARTResponse, error) {
	respInt, _, err := getRespFormats(resp)
	if err != nil {
		return nil, err
	}

	delete(respInt, field)

	respJSON, err := json.Marshal(respInt)
	if err != nil {
		return nil, err
	}

	return NewSMARTResp(respJSON)
}
