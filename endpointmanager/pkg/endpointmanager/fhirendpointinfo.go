package endpointmanager

import (
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
)

// FHIREndpointInfo represents a fielded FHIR API endpoint hosted by a
// HealthITProduct and populated by a ProviderOrganization.
// Information about the FHIR API endpoint is populated by the FHIR
// capability statement found at that endpoint.
type FHIREndpointInfo struct {
	ID                  int
	HealthITProductID   int
	URL                 string
	TLSVersion          string
	MIMETypes           []string
	HTTPResponse        int
	Errors              string
	VendorID            int
	CapabilityStatement capabilityparser.CapabilityStatement // the JSON representation of the FHIR capability statement
	Validation          map[string]interface{}
	CreatedAt           time.Time
	UpdatedAt           time.Time
	SMARTHTTPResponse   int
	SMARTResponse       SMARTResponse
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

// Equal checks each field of the two FHIREndpointInfos except for the database ID, CreatedAt and UpdatedAt fields to see if they are equal.
func (e *FHIREndpointInfo) Equal(e2 *FHIREndpointInfo) bool {
	if e == nil && e2 == nil {
		return true
	} else if e == nil {
		return false
	} else if e2 == nil {
		return false
	}

	if e.URL != e2.URL {
		return false
	}
	if e.HealthITProductID != e2.HealthITProductID {
		return false
	}

	if e.TLSVersion != e2.TLSVersion {
		return false
	}

	if !helpers.StringArraysEqual(e.MIMETypes, e2.MIMETypes) {
		return false
	}

	if e.HTTPResponse != e2.HTTPResponse {
		return false
	}
	if e.Errors != e2.Errors {
		return false
	}
	if e.VendorID != e2.VendorID {
		return false
	}
	// because CapabilityStatement is an interface, we need to confirm it's not nil before using the Equal
	// method.
	if e.CapabilityStatement != nil && !e.CapabilityStatement.Equal(e2.CapabilityStatement) {
		return false
	}
	if e.CapabilityStatement == nil && e2.CapabilityStatement != nil {
		return false
	}
	if e.SMARTHTTPResponse != e2.SMARTHTTPResponse {
		return false
	}
 	if e.SMARTResponse != nil && !e.SMARTResponse.Equal(e2.SMARTResponse) {
		return false
	}
	if e.SMARTResponse == nil && e2.SMARTResponse != nil {
		return false
	}

	return true
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