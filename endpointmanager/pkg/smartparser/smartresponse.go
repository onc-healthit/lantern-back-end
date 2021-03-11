package smartparser

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
)

// SMARTResponse interface wraps the smart response so users don't need to worry about the smart response version.
type SMARTResponse interface {
	Equal(SMARTResponse) bool
	EqualIgnore(SMARTResponse, []string) bool
	GetJSON() ([]byte, error)
}

// Response is a structure containing the Smart Response map interface
type Response struct {
	resp map[string]interface{}
}

// ResponseBody is a structure containing a Response struct
type ResponseBody struct {
	Response
}

// NewResponseBody returns a ResponseBody struct contanining the smart response map interface
func NewResponseBody(response map[string]interface{}) *ResponseBody {
	return &ResponseBody{
		Response: Response{
			resp: response,
		},
	}
}

// NewSMARTRespFromInterface is a method for creating a SMARTResponse from a Smart Response map interface. It creates the implementation of the
// SMARTResponse interface.
func NewSMARTRespFromInterface(response map[string]interface{}) SMARTResponse {
	if response == nil {
		return nil
	}
	return NewResponseBody(response)
}

// NewSMARTResp is a method for creating a SMARTResp from a Smart Response JSON byte array. It creates the implementation of the
// SMARTResponse interface.
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
func (resp *Response) EqualIgnore(resp2 SMARTResponse, ignoredFields []string) bool {
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

	return respCopy.Equal(resp2Copy)
}

// GetJSON returns the JSON representation of a smart response
func (resp *Response) GetJSON() ([]byte, error) {
	return json.Marshal(resp.resp)
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
