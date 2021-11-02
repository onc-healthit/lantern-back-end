package validation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

var tls12 = "TLS 1.2"
var tls13 = "TLS 1.3"

var usCoreProfiles = []string{"AllergyIntolerance", "CarePlan", "CareTeam",
	"Condition", "DiagnosticReport", "DocumentReference", "Encounter", "Goal",
	"Immunization", "Device", "Observation", "Location", "Medication",
	"MedicationRequest", "Organization", "Practitioner", "PractitionerRole",
	"Procedure", "Provenance"}

type r4Validation struct {
	baseVal
}

func newR4Val() *r4Validation {
	return &r4Validation{
		baseVal: baseVal{},
	}
}

// RunValidation runs all of the defined validation checks
func (v *r4Validation) RunValidation(capStat capabilityparser.CapabilityStatement,
	httpResponse int,
	mimeTypes []string,
	fhirVersion string,
	tlsVersion string,
	smartHTTPRsp int,
	requestedFhirVersion string,
	defaultFhirVersion string) endpointmanager.Validation {
	var validationResults []endpointmanager.Rule
	validationWarnings := make([]endpointmanager.Rule, 0)

	returnedRule := v.CapStatExists(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.MimeTypeValid(mimeTypes, fhirVersion)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.HTTPResponseValid(httpResponse)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.baseVal.FhirVersion(fhirVersion)
	validationResults = append(validationResults, returnedRule)

	if requestedFhirVersion == "None" {
		returnedRule = v.VersionResponseValid(fhirVersion, defaultFhirVersion)
		validationResults = append(validationResults, returnedRule)
	}

	returnedRule = v.TLSVersion(tlsVersion)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.PatientResourceExists(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.OtherResourceExists(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.SmartHTTPResponseValid(smartHTTPRsp)
	validationResults = append(validationResults, returnedRule)

	returnedRules := v.KindValid(capStat)
	validationResults = append(validationResults, returnedRules[0], returnedRules[1])

	returnedRule = v.MessagingEndpointValid(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.EndpointFunctionValid(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.DescribeEndpointValid(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.DocumentSetValid(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.UniqueResources(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.SearchParamsUnique(capStat)
	validationResults = append(validationResults, returnedRule)

	validations := endpointmanager.Validation{
		Results:  validationResults,
		Warnings: validationWarnings,
	}

	return validations
}

// CapStatExists checks if the capability statement exists using the base function, and then
// adds specific R4 reference information
func (v *r4Validation) CapStatExists(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.CapStatExists(capStat)
	baseRule.Comment = "Servers SHALL provide a Capability Statement that specifies which interactions and resources are supported."
	baseRule.Reference = "http://hl7.org/fhir/http.html"
	baseRule.ImplGuide = "USCore 3.1"
	return baseRule
}

// MimeTypeValid checks if the given mime types include the correct mime type for the given version
// using the base function, and then adds specific R4 reference information
func (v *r4Validation) MimeTypeValid(mimeTypes []string, fhirVersion string) endpointmanager.Rule {
	baseRule := v.baseVal.MimeTypeValid(mimeTypes, fhirVersion)
	baseRule.Reference = "http://hl7.org/fhir/http.html"
	baseRule.ImplGuide = "USCore 3.1"
	return baseRule
}

// HTTPResponseValid checks if the given response code is 200 using the base function, and then
// adds specific R4 reference information
func (v *r4Validation) HTTPResponseValid(httpResponse int) endpointmanager.Rule {
	baseRule := v.baseVal.HTTPResponseValid(httpResponse)
	baseRule.Reference = "http://hl7.org/fhir/http.html"
	baseRule.ImplGuide = "USCore 3.1"
	baseRule.Comment = baseRule.Comment + "Applications SHALL return a resource that describes the functionality of the server end-point."
	return baseRule
}

// TLSVersion checks if the given TLS version string is version 1.2 or higher, which is a
// USCore security requirement
func (v *r4Validation) TLSVersion(tlsVersion string) endpointmanager.Rule {
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.TLSVersion,
		Valid:     true,
		Expected:  "TLS 1.2, TLS 1.3",
		Actual:    tlsVersion,
		Comment:   "Systems SHALL use TLS version 1.2 or higher for all transmissions not taking place over a secure network connection.",
		Reference: "https://www.hl7.org/fhir/us/core/security.html",
		ImplGuide: "USCore 3.1",
	}

	if (tlsVersion != tls12) && (tlsVersion != tls13) {
		ruleError.Valid = false
	}

	return ruleError
}

// PatientResourceExists checks to see if the Patient resource is included in the resource list,
// which is a USCore requirement
func (v *r4Validation) PatientResourceExists(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseComment := "The US Core Server SHALL support the US Core Patient resource profile."
	returnVal := checkResourceList(capStat, endpointmanager.PatResourceExists)
	returnVal.Comment = returnVal.Comment + baseComment
	returnVal.Reference = "https://www.hl7.org/fhir/us/core/CapabilityStatement-us-core-server.html"
	return returnVal
}

// OtherResourceExists checks to see if there is another resource besides Patient included
// in the resource list, which is a USCore requirement
func (v *r4Validation) OtherResourceExists(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseComment := "The US Core Server SHALL support at least one additional resource profile (besides Patient) from the list of US Core Profiles."
	returnVal := checkResourceList(capStat, endpointmanager.OtherResourceExists)
	returnVal.Comment = returnVal.Comment + baseComment
	returnVal.Reference = "https://www.hl7.org/fhir/us/core/CapabilityStatement-us-core-server.html"
	return returnVal
}

// checkResourceList gets the resources from the Capability statement, which is used in various validation
// checks, and then runs the given check based on the rule parameter
func checkResourceList(capStat capabilityparser.CapabilityStatement, rule endpointmanager.RuleOption) endpointmanager.Rule {
	ruleError := endpointmanager.Rule{
		RuleName:  rule,
		Valid:     false,
		Actual:    "false",
		Expected:  "true",
		ImplGuide: "USCore 3.1",
	}

	if capStat == nil {
		ruleError.Comment = "The Capability Statement does not exist; cannot check resource profiles. "
		return ruleError
	}

	rest, err := capStat.GetRest()
	if err != nil || len(rest) == 0 {
		ruleError.Comment = "Rest field does not exist. "
		return ruleError
	}

	var uniqueRecs []string
	areParamsValid := true
	for _, restElem := range rest {
		resourceList, err := capStat.GetResourceList(restElem)
		if err != nil || len(resourceList) == 0 {
			ruleError.Comment = "The Resource Profiles do not exist. "
			return ruleError
		}
		for _, resource := range resourceList {
			typeVal := resource["type"]
			if typeVal == nil {
				ruleError.Comment = "The Resource Profiles are not properly formatted. "
				return ruleError
			}
			typeStr, ok := typeVal.(string)
			if !ok {
				ruleError.Comment = "The Resource Profiles are not properly formatted. "
				return ruleError
			}
			if rule == endpointmanager.OtherResourceExists {
				if stringInList(typeStr, usCoreProfiles) {
					ruleError.Valid = true
					ruleError.Actual = "true"
					return ruleError
				}
			} else if rule == endpointmanager.PatResourceExists {
				if typeStr == "Patient" {
					ruleError.Valid = true
					ruleError.Actual = "true"
					return ruleError
				}
			} else if rule == endpointmanager.UniqueResourcesRule {
				if stringInList(typeStr, uniqueRecs) {
					ruleError.Comment = fmt.Sprintf("The resource type %s is not unique. ", typeStr)
					return ruleError
				}
				uniqueRecs = append(uniqueRecs, typeStr)
			} else if rule == endpointmanager.SearchParamsRule {
				check, err := areSearchParamsValid(resource)
				if err != nil {
					areParamsValid = false
					ruleError.Comment = ruleError.Comment + fmt.Sprintf("The resource type %s is not formatted properly. ", typeStr)
					continue
				}
				if !check {
					areParamsValid = false
					ruleError.Comment = ruleError.Comment + fmt.Sprintf("The resource type %s does not have unique searchParams. ", typeStr)
				}
			}
		}
	}

	if rule == endpointmanager.UniqueResourcesRule || (rule == endpointmanager.SearchParamsRule && areParamsValid) {
		ruleError.Valid = true
		ruleError.Actual = "true"
		return ruleError
	}
	return ruleError
}

// SmartHTTPResponseValid checks if the SMART-on-FHIR response code is 200 using the base
// HTTPResponse function, and then adds specific R4 reference information
func (v *r4Validation) SmartHTTPResponseValid(smartHTTPRsp int) endpointmanager.Rule {
	baseComment := "FHIR endpoints requiring authorization SHALL serve a JSON document at the location formed by appending /.well-known/smart-configuration to their base URL."
	baseRule := v.baseVal.HTTPResponseValid(smartHTTPRsp)
	baseRule.RuleName = endpointmanager.SmartHTTPRespRule
	baseRule.Comment = baseComment
	baseRule.Reference = "http://www.hl7.org/fhir/smart-app-launch/conformance/index.html"
	baseRule.ImplGuide = "USCore 3.1"
	if (smartHTTPRsp != 0) && (smartHTTPRsp != 200) {
		strResp := strconv.Itoa(smartHTTPRsp)
		baseRule.Comment = "The HTTP response code was " + strResp + " instead of 200. " + baseComment
	}
	return baseRule
}

// KindValid checks 2 Rules: The first, which is the baseVal rule, is that kind = instance since all of the
// endpoints we are looking at are for server instances. It then checks the rule: "If kind = instance,
// implementation should be present."
func (v *r4Validation) KindValid(capStat capabilityparser.CapabilityStatement) []endpointmanager.Rule {
	var rules []endpointmanager.Rule
	baseRule := v.baseVal.KindValid(capStat)
	baseRule[0].Reference = "http://hl7.org/fhir/capabilitystatement.html"
	baseRule[0].ImplGuide = "USCore 3.1"
	rules = append(rules, baseRule[0])

	instanceRule := endpointmanager.Rule{
		RuleName:  endpointmanager.InstanceRule,
		Valid:     true,
		Expected:  "true",
		Actual:    "true",
		Comment:   "If kind = instance, implementation must be present. This endpoint must be an instance.",
		Reference: "http://hl7.org/fhir/capabilitystatement.html",
		ImplGuide: "USCore 3.1",
	}
	impl, err := capStat.GetImplementation()
	if err != nil || len(impl) == 0 {
		instanceRule.Valid = false
		instanceRule.Actual = "false"
	}
	rules = append(rules, instanceRule)
	return rules
}

// MessagingEndpointValid checks the requirement "Messaging endpoint is required (and is only permitted) when a statement is for an implementation."
// Every endpoint we are testing should be an implementation, which means the endpoint field should be there.
func (v *r4Validation) MessagingEndpointValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseComment := "Messaging end-point is required (and is only permitted) when a statement is for an implementation. This endpoint must be an implementation."
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.MessagingEndptRule,
		Valid:     false,
		Expected:  "true",
		Actual:    "false",
		Comment:   baseComment,
		Reference: "http://hl7.org/fhir/capabilitystatement.html",
		ImplGuide: "USCore 3.1",
	}

	kindRule := v.baseVal.KindValid(capStat)
	if !kindRule[0].Valid {
		ruleError.Comment = kindRule[0].Comment + " " + baseComment
		return ruleError
	}
	messaging, err := capStat.GetMessaging()
	if err != nil || len(messaging) == 0 {
		ruleError.Comment = "Messaging does not exist. " + baseComment
		return ruleError
	}
	for _, message := range messaging {
		endpoints, err := capStat.GetMessagingEndpoint(message)
		if err != nil || len(endpoints) == 0 {
			ruleError.Comment = "Endpoint field in Messaging does not exist. " + baseComment
			return ruleError
		}
	}

	ruleError.Valid = true
	ruleError.Actual = "true"
	return ruleError
}

// EndpointFunctionValid checks the requirement "A Capability Statement SHALL have at least one of REST,
// messaging or document element."
func (v *r4Validation) EndpointFunctionValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	var actualVal []string
	baseComment := "A Capability Statement SHALL have at least one of REST, messaging or document element."
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.EndptFunctionRule,
		Valid:     true,
		Expected:  "rest,messaging,document",
		Comment:   baseComment,
		Reference: "http://hl7.org/fhir/capabilitystatement.html",
		ImplGuide: "USCore 3.1",
	}
	// If rest is not nil, add to actual list
	rest, err := capStat.GetRest()
	if err == nil && len(rest) > 0 {
		actualVal = append(actualVal, "rest")
	}
	// If messaging is not nil, add to actual list
	messaging, err := capStat.GetMessaging()
	if err == nil && len(messaging) > 0 {
		actualVal = append(actualVal, "messaging")
	}
	// if document is not nil, add to actual list
	document, err := capStat.GetDocument()
	if err == nil && len(document) > 0 {
		actualVal = append(actualVal, "document")
	}
	// If none of the above exist, the capability statement is not valid
	if len(actualVal) == 0 {
		ruleError.Actual = ""
		ruleError.Valid = false
		return ruleError
	}
	ruleError.Actual = strings.Join(actualVal, ",")
	return ruleError
}

// DescribeEndpointValid checks the requirement: "A Capability Statement SHALL have at least one of description,
// software, or implementation element."
func (v *r4Validation) DescribeEndpointValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	var actualVal []string
	baseComment := "A Capability Statement SHALL have at least one of description, software, or implementation element."
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.DescribeEndptRule,
		Valid:     true,
		Expected:  "description,software,implementation",
		Comment:   baseComment,
		Reference: "http://hl7.org/fhir/capabilitystatement.html",
		ImplGuide: "USCore 3.1",
	}
	// If description is not an empty string, add to actual list
	description, err := capStat.GetDescription()
	if err == nil && len(description) > 0 {
		actualVal = append(actualVal, "description")
	}
	// If software is not nil, add to actual list
	software, err := capStat.GetSoftware()
	if err == nil && len(software) > 0 {
		actualVal = append(actualVal, "software")
	}
	// if implementation is not nil, add to actual list
	implementation, err := capStat.GetImplementation()
	if err == nil && len(implementation) > 0 {
		actualVal = append(actualVal, "implementation")
	}
	// If none of the above exist, the capability statement is not valid
	if len(actualVal) == 0 {
		ruleError.Actual = ""
		ruleError.Valid = false
		return ruleError
	}
	ruleError.Actual = strings.Join(actualVal, ",")
	return ruleError
}

// DocumentSetValid checks the requirement: "The set of documents must be unique by the combination of profile and mode."
func (v *r4Validation) DocumentSetValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseComment := "The set of documents must be unique by the combination of profile and mode."
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.DocumentValidRule,
		Valid:     false,
		Expected:  "true",
		Actual:    "false",
		Comment:   baseComment,
		Reference: "http://hl7.org/fhir/capabilitystatement.html",
		ImplGuide: "USCore 3.1",
	}
	document, err := capStat.GetDocument()
	if err != nil {
		ruleError.Comment = "Document field is not formatted correctly. Cannot check if the set of documents are unique. " + baseComment
		return ruleError
	}
	if err == nil && len(document) == 0 {
		ruleError.Valid = true
		ruleError.Actual = "true"
		ruleError.Comment = "Document field does not exist, but is not required. " + baseComment
		return ruleError
	}
	var uniqueIDs []string
	invalid := false
	for _, doc := range document {
		mode := doc["mode"]
		if mode == nil {
			invalid = true
			break
		}
		modeStr, ok := mode.(string)
		if !ok {
			invalid = true
			break
		}
		profile := doc["profile"]
		if profile == nil {
			invalid = true
			break
		}
		profileStr, ok := profile.(string)
		if !ok {
			invalid = true
			break
		}
		// Combine profile & mode to compare against other defined documents
		id := profileStr + "." + modeStr
		if stringInList(id, uniqueIDs) {
			ruleError.Comment = "The set of documents are not unique. " + baseComment
			return ruleError
		}
		uniqueIDs = append(uniqueIDs, id)
	}
	if invalid {
		ruleError.Comment = "Document field is not formatted correctly. Cannot check if the set of documents are unique. " + baseComment
		return ruleError
	}
	ruleError.Valid = true
	ruleError.Actual = "true"
	return ruleError
}

// UniqueResources checks the requirement: "A given resource can only be described once per RESTful mode."
func (v *r4Validation) UniqueResources(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseComment := "A given resource can only be described once per RESTful mode."
	returnVal := checkResourceList(capStat, endpointmanager.UniqueResourcesRule)
	returnVal.Comment = returnVal.Comment + baseComment
	returnVal.Reference = "http://hl7.org/fhir/capabilitystatement.html"
	return returnVal
}

// SearchParamsUnique checks the requirement: "Search parameter names must be unique in the context of a resource."
func (v *r4Validation) SearchParamsUnique(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseComment := "Search parameter names must be unique in the context of a resource."
	returnVal := checkResourceList(capStat, endpointmanager.SearchParamsRule)
	returnVal.Comment = returnVal.Comment + baseComment
	returnVal.Reference = "http://hl7.org/fhir/capabilitystatement.html"
	return returnVal
}

// areSearchParamsValid checks each resource's searchParam field and makes sure all of the values
// are unique. The searchParam field is not required.
func areSearchParamsValid(resource map[string]interface{}) (bool, error) {
	var searchParams []string
	search := resource["searchParam"]
	if search == nil {
		return true, nil
	}
	searchList, ok := search.([]interface{})
	if !ok {
		return false, fmt.Errorf("Unable to cast searchParam value in a resource to a list")
	}
	for _, elem := range searchList {
		obj, ok := elem.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("Unable to cast element of searchParam list to a map[string]interface{}")
		}
		name := obj["name"]
		if name == nil {
			return false, fmt.Errorf("Name does not exist but is required in searchParam values")
		}
		nameStr, ok := obj["name"].(string)
		if !ok {
			return false, fmt.Errorf("Unable to cast the name of a searchParam to a string")
		}
		if stringInList(nameStr, searchParams) {
			return false, nil
		}
		searchParams = append(searchParams, nameStr)
	}
	return true, nil
}

// VersionResponseValid checks if $versions operation is supported and that the default version is returned when no version requested
func (v *r4Validation) VersionResponseValid(fhirVersion string, defaultFhirVersion string) endpointmanager.Rule {
	ruleError := endpointmanager.Rule{
		RuleName: endpointmanager.VersionsResponseRule,
		Valid:    true,
		Expected: defaultFhirVersion,
		Comment:  "Expected $versions operation to be supported, and expected default fhir version to be returned from server when no version specified.",
	}

	if defaultFhirVersion == "None" {
		ruleError.Valid = false
		ruleError.Actual = fhirVersion
		ruleError.Expected = ""
		ruleError.Comment = "Expected $versions operation to be supported, but no response was received"
		return ruleError
	}
	fhirVersionSplit := strings.Split(fhirVersion, ".")
	defaultVersionSplit := strings.Split(defaultFhirVersion, ".")
	if fhirVersionSplit[0] != defaultVersionSplit[0] || fhirVersionSplit[1] != defaultVersionSplit[1] {
		ruleError.Valid = false
	}

	ruleError.Actual = fhirVersion

	return ruleError
}

func stringInList(str string, list []string) bool {
	for _, b := range list {
		if b == str {
			return true
		}
	}
	return false
}
