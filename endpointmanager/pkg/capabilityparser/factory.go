package capabilityparser

import (
	"encoding/json"
	"fmt"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
	
	log.Warn(fmt.Errorf("unknown FHIR version, %s, defaulting to DSTU2", fhirVersion))
	return newDSTU2(capStat), nil
}
