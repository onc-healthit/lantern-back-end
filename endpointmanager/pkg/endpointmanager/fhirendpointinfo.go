package endpointmanager

import (
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/smartparser"
)

// FHIREndpointInfo represents a fielded FHIR API endpoint hosted by a
// HealthITProduct and populated by a ProviderOrganization.
// Information about the FHIR API endpoint is populated by the FHIR
// capability statement found at that endpoint.
type FHIREndpointInfo struct {
	ID                    int
	HealthITProductID     int
	URL                   string
	TLSVersion            string
	MIMETypes             []string
	VendorID              int
	CapabilityStatement   capabilityparser.CapabilityStatement // the JSON representation of the FHIR capability statement
	Validation            Validation
	CreatedAt             time.Time
	UpdatedAt             time.Time
	SMARTResponse         smartparser.SMARTResponse
	IncludedFields        []IncludedField
	OperationResource     map[string][]string
	Metadata              *FHIREndpointMetadata
	RequestedFhirVersion  string
	CapabilityFhirVersion string
}

// EqualExcludeMetadata checks each field of the two FHIREndpointInfos except for metadata fields to see if they are equal.
func (e *FHIREndpointInfo) EqualExcludeMetadata(e2 *FHIREndpointInfo) bool {
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

	if e.VendorID != e2.VendorID {
		return false
	}

	if e.RequestedFhirVersion != e2.RequestedFhirVersion {
		return false
	}

	if e.CapabilityFhirVersion != e2.CapabilityFhirVersion {
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
	if e.SMARTResponse != nil && !e.SMARTResponse.Equal(e2.SMARTResponse) {
		return false
	}
	if e.SMARTResponse == nil && e2.SMARTResponse != nil {
		return false
	}

	if !cmp.Equal(e.Validation, e2.Validation) {
		return false
	}

	if !cmp.Equal(e.IncludedFields, e2.IncludedFields) {
		return false
	}

	// If the two endpoints have the same values in a different order, the Equal
	// function will return false, so the resources need to be sorted for the Equal
	// function to work as expected
	return compareOperations(e.OperationResource, e2.OperationResource)
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

	if !e.EqualExcludeMetadata(e2) {
		return false
	}
	if e.Metadata != nil && !e.Metadata.Equal(e2.Metadata) {
		return false
	}
	if e.Metadata == nil && e2.Metadata != nil {
		return false
	}

	return true
}

// IncludedField is a struct used to keep track of all of the fields in the capability statement
type IncludedField struct {
	Field     string
	Exists    bool
	Extension bool
}

// Validation holds all of the errors and warnings from running the validation checks
// it is saved in JSON format to the fhir_endpoints_info database table
type Validation struct {
	Results  []Rule `json:"results"`
	Warnings []Rule `json:"warnings"`
}

// Rule is the structure for both validation errors and warnings that are saved in
// the Validations struct
type Rule struct {
	RuleName  RuleOption `json:"ruleName"`
	Valid     bool       `json:"valid"`
	Expected  string     `json:"expected"`
	Actual    string     `json:"actual"`
	Comment   string     `json:"comment"`
	Reference string     `json:"reference"`
	ImplGuide string     `json:"implGuide"`
}

// RuleOption is an enum of the names given to the rule validation checks
type RuleOption string

const (
	GeneralMimeTypeRule  RuleOption = "generalMimeType"
	HTTPResponseRule     RuleOption = "httpResponse"
	CapStatExistRule     RuleOption = "capStatExist"
	FHIRVersion          RuleOption = "fhirVersion"
	TLSVersion           RuleOption = "tlsVersion"
	PatResourceExists    RuleOption = "patResourceExists"
	OtherResourceExists  RuleOption = "otherResourceExists"
	SmartHTTPRespRule    RuleOption = "smartHttpResponse"
	KindRule             RuleOption = "kindRule"
	InstanceRule         RuleOption = "instanceRule"
	MessagingEndptRule   RuleOption = "messagingEndptRule"
	EndptFunctionRule    RuleOption = "endpointFunctionRule"
	DescribeEndptRule    RuleOption = "describeEndpointRule"
	DocumentValidRule    RuleOption = "documentValidRule"
	UniqueResourcesRule  RuleOption = "uniqueResourcesRule"
	SearchParamsRule     RuleOption = "searchParamsRule"
	VersionsResponseRule RuleOption = "versionsResponseRule"
)

// compareOperations compares the operation resource fields for an endpoint
// and returns whether or not they are equivalent
func compareOperations(e1 map[string][]string, e2 map[string][]string) bool {
	// If they don't have the same number of keys then they're not equal
	if len(e1) != len(e2) {
		return false
	}
	for key, e1val := range e1 {
		// If they both have the given key, check to see if their values are equal
		// If e1 has a key that e2 doesn't have, then they're not equal
		if e2val, ok := e2[key]; ok {
			if !helpers.StringArraysEqual(e1val, e2val) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}
