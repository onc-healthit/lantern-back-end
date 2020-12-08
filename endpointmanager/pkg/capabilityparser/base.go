package capabilityparser

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// base struct to handle any methods that don't change between the versions of FHIR
// capability statements
type baseParser struct {
	capStat map[string]interface{}
	version string
}

// GetPublisher returns the publisher field from the conformance/capability statement.
func (cp *baseParser) GetPublisher() (string, error) {
	publisher := cp.capStat["publisher"]
	if publisher == nil {
		return "", nil
	}
	publisherStr, ok := publisher.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement publisher value to a string", cp.version)
	}
	return publisherStr, nil
}

// GetFHIRVersion returns the FHIR version specifiedin the conformance/capability statement.
func (cp *baseParser) GetFHIRVersion() (string, error) {
	fhirVersion := cp.capStat["fhirVersion"]
	if fhirVersion == nil {
		return "", nil
	}
	fhirVersionStr, ok := fhirVersion.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement fhirVersion value to a string", cp.version)
	}
	return fhirVersionStr, nil
}

// GetSoftware returns the software field from the conformance/capability statement.
func (cp *baseParser) GetSoftware() (map[string]interface{}, error) {
	var defaultVal map[string]interface{}

	software := cp.capStat["software"]
	if software == nil {
		return defaultVal, nil
	}
	softwareMap, ok := software.(map[string]interface{})
	if !ok {
		return defaultVal, fmt.Errorf("unable to cast %s capability statement software value to a map[string]interface{}", cp.version)
	}
	return softwareMap, nil
}

// GetSoftwareName returns the software name specified in the conformance/capability statement.
func (cp *baseParser) GetSoftwareName() (string, error) {
	softwareMap, err := cp.GetSoftware()
	if err != nil || len(softwareMap) == 0 {
		return "", err
	}
	name := softwareMap["name"]
	if name == nil {
		return "", nil
	}
	nameStr, ok := name.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement software.name value to a string", cp.version)
	}
	return nameStr, nil
}

// GetSoftwareVersion returns the software version specified in the conformance/capability statement.
func (cp *baseParser) GetSoftwareVersion() (string, error) {
	softwareMap, err := cp.GetSoftware()
	if err != nil || len(softwareMap) == 0 {
		return "", err
	}
	version := softwareMap["version"]
	if version == nil {
		return "", nil
	}
	versionStr, ok := version.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement software.version value to a string", cp.version)
	}
	return versionStr, nil
}

// GetCopyright returns the copyright specified in the capability/conformance statement.
func (cp *baseParser) GetCopyright() (string, error) {
	copyright := cp.capStat["copyright"]
	if copyright == nil {
		return "", nil
	}
	copyrightStr, ok := copyright.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement copyright value to a string", cp.version)
	}
	return copyrightStr, nil
}

// GetRest returns the rest array specified in the capability/conformance statement.
func (cp *baseParser) GetRest() ([]map[string]interface{}, error) {
	var returnList []map[string]interface{}

	rest := cp.capStat["rest"]
	if rest == nil {
		return returnList, nil
	}
	restList, ok := rest.([]interface{})
	if !ok {
		return returnList, fmt.Errorf("unable to cast %s capability statement rest value to a []interface{}", cp.version)
	}
	for _, restElem := range restList {
		restMap, ok := restElem.(map[string]interface{})
		if !ok {
			return returnList, fmt.Errorf("unable to cast %s capability statement messaging value to a map[string]interface{}", cp.version)
		}
		returnList = append(returnList, restMap)
	}
	return returnList, nil
}

// GetResourceList returns the list of resources in the given rest map of the capability/conformance statement.
func (cp *baseParser) GetResourceList(rest map[string]interface{}) ([]map[string]interface{}, error) {
	var returnList []map[string]interface{}

	resource := rest["resource"]
	if resource == nil {
		return returnList, nil
	}
	resourceList, ok := resource.([]interface{})
	if !ok {
		return returnList, fmt.Errorf("unable to cast %s capability statement resource list value to an []interface{}", cp.version)
	}
	for _, resource := range resourceList {
		resourceMap, ok := resource.(map[string]interface{})
		if !ok {
			return returnList, fmt.Errorf("unable to cast %s capability statement resource value to a map[string]interface{}", cp.version)
		}
		returnList = append(returnList, resourceMap)
	}
	return returnList, nil
}

// GetKind returns the kind specified in the capability/conformance statement.
func (cp *baseParser) GetKind() (string, error) {
	kind := cp.capStat["kind"]
	if kind == nil {
		return "", nil
	}
	kindStr, ok := kind.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement kind value to a string", cp.version)
	}
	return kindStr, nil
}

// GetImplementation returns the implementation specified in the capability/conformance statement.
func (cp *baseParser) GetImplementation() (map[string]interface{}, error) {
	var defaultVal map[string]interface{}
	impl := cp.capStat["implementation"]
	if impl == nil {
		return defaultVal, nil
	}
	implMap, ok := impl.(map[string]interface{})
	if !ok {
		return defaultVal, fmt.Errorf("unable to cast %s capability statement implementation value to a map[string]interface{}", cp.version)
	}
	return implMap, nil
}

// GetMessaging returns the messaging field specified in the capability/conformance statement.
func (cp *baseParser) GetMessaging() ([]map[string]interface{}, error) {
	var returnList []map[string]interface{}

	messaging := cp.capStat["messaging"]
	if messaging == nil {
		return returnList, nil
	}
	messagingList, ok := messaging.([]interface{})
	if !ok {
		return returnList, fmt.Errorf("unable to cast %s capability statement messaging value to a []interface{}", cp.version)
	}
	for _, message := range messagingList {
		messageMap, ok := message.(map[string]interface{})
		if !ok {
			return returnList, fmt.Errorf("unable to cast %s capability statement messaging value to a map[string]interface{}", cp.version)
		}
		returnList = append(returnList, messageMap)
	}
	return returnList, nil
}

// GetMessagingEndpoint gets a list of the given messaging element's endpoints from the capability/conformance statement.
func (cp *baseParser) GetMessagingEndpoint(messaging map[string]interface{}) ([]map[string]interface{}, error) {
	var returnList []map[string]interface{}

	endpoint := messaging["endpoint"]
	if endpoint == nil {
		return returnList, nil
	}
	endpointList, ok := endpoint.([]interface{})
	if !ok {
		return returnList, fmt.Errorf("unable to cast %s capability statement endpoint list value to an []interface{}", cp.version)
	}
	for _, e := range endpointList {
		endpointMap, ok := e.(map[string]interface{})
		if !ok {
			return returnList, fmt.Errorf("unable to cast %s capability statement endpoint value to a map[string]interface{}", cp.version)
		}
		returnList = append(returnList, endpointMap)
	}
	return returnList, nil
}

// GetDocument returns the document specified in the capability/conformance statement.
func (cp *baseParser) GetDocument() ([]map[string]interface{}, error) {
	var returnList []map[string]interface{}

	document := cp.capStat["document"]
	if document == nil {
		return returnList, nil
	}
	documentList, ok := document.([]interface{})
	if !ok {
		return returnList, fmt.Errorf("unable to cast %s capability statement document value to a []interface{}", cp.version)
	}
	for _, doc := range documentList {
		docMap, ok := doc.(map[string]interface{})
		if !ok {
			return returnList, fmt.Errorf("unable to cast %s capability statement document array value to a map[string]interface{}", cp.version)
		}
		returnList = append(returnList, docMap)
	}
	return returnList, nil
}

// GetDescription returns the description specified in the capability/conformance statement.
func (cp *baseParser) GetDescription() (string, error) {
	description := cp.capStat["description"]
	if description == nil {
		return "", nil
	}
	descriptionStr, ok := description.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s capability statement description value to a string", cp.version)
	}
	return descriptionStr, nil
}

// EqualIgnore checks if the conformance/capability statement is equal to the given conformance/capability statement while ignoring certain fields that may differ.
func (cp *baseParser) EqualIgnore(cs2 CapabilityStatement) bool {
	ignoredFields := []string{"date"}

	if cs2 == nil {
		return false
	}

	var cpCopy *baseParser
	var cs2Copy CapabilityStatement
	err := DeepCopy(cp, cpCopy)
	if err != nil {
		return false
	}
	err = DeepCopy(cs2, cs2Copy)
	if err != nil {
		return false
	}

	for _, field := range ignoredFields {
		DeleteFieldFromCapStat(cpCopy, field)
		DeleteFieldFromCapStat(cs2Copy, field)
	}

	if err != nil {
		return false
	}

	j1, err := cpCopy.GetJSON()
	if err != nil {
		return false
	}
	j2, err := cs2Copy.GetJSON()
	if err != nil {
		return false
	}
	if !bytes.Equal(j1, j2) {
		return false
	}

	return true
}

// Equal checks if the conformance/capability statement is equal to the given conformance/capability statement.
func (cp *baseParser) Equal(cs2 CapabilityStatement) bool {
	if cs2 == nil {
		return false
	}

	j1, err := cp.GetJSON()
	if err != nil {
		return false
	}
	j2, err := cs2.GetJSON()
	if err != nil {
		return false
	}
	if !bytes.Equal(j1, j2) {
		return false
	}

	return true
}

// GetJSON returns the JSON representation of the capability statement.
func (cp *baseParser) GetJSON() ([]byte, error) {
	return json.Marshal(cp.capStat)
}

// DeepCopy deepcopies a to b using json marshaling
func DeepCopy(a, b interface{}) error {
	byt, err := json.Marshal(a)
	if err != nil {
		return err
	}
	json.Unmarshal(byt, b)
	return nil
}
