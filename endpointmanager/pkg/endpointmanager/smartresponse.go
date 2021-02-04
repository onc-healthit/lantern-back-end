package endpointmanager

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
)

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
