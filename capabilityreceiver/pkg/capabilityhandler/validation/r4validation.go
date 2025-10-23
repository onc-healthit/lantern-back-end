package validation

import (
	"fmt"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/smartparser"
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
	fhirVersion string,
	tlsVersion string,
	smartRsp smartparser.SMARTResponse,
	requestedFhirVersion string,
	defaultFhirVersion string) endpointmanager.Validation {
	var validationResults []endpointmanager.Rule

	returnedRule := v.CapStatExists(capStat)
	validationResults = append(validationResults, returnedRule)

	if requestedFhirVersion == "None" && defaultFhirVersion != "" {
		returnedRule = v.VersionResponseValid(fhirVersion, defaultFhirVersion)
		validationResults = append(validationResults, returnedRule)
	}

	returnedRule = v.TLSVersion(tlsVersion)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.PatientResourceExists(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.OtherResourceExists(capStat)
	validationResults = append(validationResults, returnedRule)

	returnedRule = v.SmartResponseExists(smartRsp)
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
		Results: validationResults,
	}

	return validations
}

// CapStatExists checks if the capability statement exists using the base function, and then
// adds specific R4 reference information
func (v *r4Validation) CapStatExists(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseComment := "Servers SHALL provide a Capability Statement that specifies which interactions and resources are supported."

	baseRule := v.baseVal.CapStatExists(capStat)
	baseRule.Reference = "http://hl7.org/fhir/http.html"
	baseRule.ImplGuide = "USCore 3.1"

	if baseRule.Valid {
		baseRule.Comment = "The Capability Statement exists. " + baseComment
	} else {
		baseRule.Comment = "The Capability Statement does not exist. " + baseComment
	}

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

// SmartResponseExists checks if the SMART-on-FHIR response exists
func (v *r4Validation) SmartResponseExists(smartRsp smartparser.SMARTResponse) endpointmanager.Rule {
	ruleError := endpointmanager.Rule{
		RuleName:  endpointmanager.SmartRespExistsRule,
		Valid:     true,
		Expected:  "true",
		Actual:    "true",
		Comment:   "FHIR endpoints requiring authorization SHALL serve a JSON document at the location formed by appending /.well-known/smart-configuration to their base URL.",
		Reference: "http://www.hl7.org/fhir/smart-app-launch/conformance/index.html",
		ImplGuide: "USCore 3.1",
	}

	if smartRsp != nil {
		return ruleError
	}

	ruleError.Valid = false
	ruleError.Actual = "false"
	ruleError.Comment = `The SMART Response does not exist. FHIR endpoints requiring authorization SHALL serve a JSON document at the location formed by appending /.well-known/smart-configuration to their base URL.`
	return ruleError
}

// KindValid checks 2 Rules: The first, which is the baseVal rule, is that kind = instance since all of the
// endpoints we are looking at are for server instances. It then checks the rule: "If kind = instance,
// implementation should be present."
func (v *r4Validation) KindValid(capStat capabilityparser.CapabilityStatement) []endpointmanager.Rule {
	baseComment := "Kind value should be set to 'instance' because this is a specific system instance."

	var rules []endpointmanager.Rule
	baseRule := v.baseVal.KindValid(capStat)
	baseRule[0].Reference = "http://hl7.org/fhir/capabilitystatement.html"
	baseRule[0].ImplGuide = "USCore 3.1"
	rules = append(rules, baseRule[0])

	if capStat == nil {
		rules[0].Comment = "Capability Statement does not exist; cannot check kind value. " + baseComment
		return baseRule
	}

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
	baseKindComment := "Kind value should be set to 'instance' because this is a specific system instance."
	baseMessagingComment := "Messaging end-point is required (and is only permitted) when a statement is for an implementation. This endpoint must be an implementation."

	baseRule := v.baseVal.MessagingEndpointValid(capStat)
	baseRule.Reference = "http://hl7.org/fhir/capabilitystatement.html"

	if capStat == nil {
		baseRule.Comment = "Capability Statement does not exist; cannot check kind value. " + baseKindComment + " " + baseMessagingComment
	}

	return baseRule
}

// EndpointFunctionValid checks the requirement "A Capability Statement SHALL have at least one of REST,
// messaging or document element."
func (v *r4Validation) EndpointFunctionValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.EndpointFunctionValid(capStat)
	baseRule.Reference = "http://hl7.org/fhir/capabilitystatement.html"
	baseRule.Comment = "A Capability Statement SHALL have at least one of REST, messaging or document element."

	if capStat == nil {
		baseRule.Comment = "The Capability Statement does not exist; cannot check REST, messaging or document elements."
	}

	return baseRule
}

// DescribeEndpointValid checks the requirement: "A Capability Statement SHALL have at least one of description,
// software, or implementation element."
func (v *r4Validation) DescribeEndpointValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.DescribeEndpointValid(capStat)
	baseRule.Reference = "http://hl7.org/fhir/capabilitystatement.html"
	baseRule.Comment = "A Capability Statement SHALL have at least one of description, software, or implementation element."

	if capStat == nil {
		baseRule.Comment = "The Capability Statement does not exist; cannot check description, software, or implementation elements."
	}

	return baseRule
}

// DocumentSetValid checks the requirement: "The set of documents must be unique by the combination of profile and mode."
func (v *r4Validation) DocumentSetValid(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.DocumentSetValid(capStat)
	baseRule.Reference = "http://hl7.org/fhir/capabilitystatement.html"

	if capStat == nil {
		baseRule.Comment = "The Capability Statement does not exist; cannot check documents."
	}

	return baseRule
}

// UniqueResources checks the requirement: "A given resource can only be described once per RESTful mode."
func (v *r4Validation) UniqueResources(capStat capabilityparser.CapabilityStatement) endpointmanager.Rule {
	baseRule := v.baseVal.UniqueResources(capStat)
	baseRule.Reference = "http://hl7.org/fhir/capabilitystatement.html"
	return baseRule
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
		RuleName:  endpointmanager.VersionsResponseRule,
		Valid:     true,
		Expected:  defaultFhirVersion,
		Reference: "https://www.hl7.org/fhir/capabilitystatement-operation-versions.html",
		Comment:   "The default fhir version as specified by the $versions operation should be returned from server when no version specified.",
	}

	// As long as the major and minor versions of the returned FHIR version and default FHIR version are the same we don't need to worry about the 3rd number
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
