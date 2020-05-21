package capabilityparser

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// from https://www.hl7.org/fhir/codesystem-FHIR-version.html
// looking at official and release versions only
var dstu2 = []string{"1.0.1", "1.0.2"}
var stu3 = []string{"3.0.0", "3.0.1"}
var r4 = []string{"4.0.0", "4.0.1"}

// CapabilityStatement provides access to key fields of the capability statement. It wraps the capability statements
// so users don't need to worry about the capability statement version.
type CapabilityStatement interface {
	GetPublisher() (string, error)
	GetFHIRVersion() (string, error)
	GetSoftwareName() (string, error)
	GetSoftwareVersion() (string, error)
	GetCopyright() (string, error)

	Equal(CapabilityStatement) bool
	GetJSON() ([]byte, error)
}

type SMARTResponse interface {
	Equal(SMARTResponse) bool
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

func NewSMARTRespFromInterface(response map[string]interface{}) (SMARTResponse, error) {
	if response == nil {
		return nil, nil
	}
	return NewResponseBody(response), nil
}

func NewSMARTResp(respJSON []byte) (SMARTResponse, error) {
	var err error
	var respMsg map[string]interface{}

	if len(respJSON) == 0 {
		return nil, nil
	}

	err = json.Unmarshal(respJSON, &respMsg)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling JSON capability statement")
	}

	return NewSMARTRespFromInterface(respMsg)
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

	if contains(dstu2, fhirVersion) {
		return newDSTU2(capStat), nil
	} else if contains(stu3, fhirVersion) {
		return newSTU3(capStat), nil
	} else if contains(r4, fhirVersion) {
		return newR4(capStat), nil
	}

	return nil, fmt.Errorf("unknown FHIR version %s", fhirVersion)
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
