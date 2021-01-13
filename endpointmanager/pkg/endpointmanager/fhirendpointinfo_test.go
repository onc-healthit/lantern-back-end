package endpointmanager

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
)

var testIncludedFields = []IncludedField{
	{
		Field:     "url",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "date",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "kind",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "name",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "title",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "format",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "status",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "contact",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "imports",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "profile",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "purpose",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "version",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "copyright",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "publisher",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "useContext",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "description",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "fhirVersion",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "patchFormat",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "experimental",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "instantiates",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "jurisdiction",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "requirements",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "acceptUnknown",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "software.name",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "software.version",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "implementation.url",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "implementationGuide",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "software.releaseDate",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "implementation.custodian",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "implementation.description",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "messaging",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "document",
		Exists:    false,
		Extension: false,
	},
}

func Test_FHIREndpointInfoEqual(t *testing.T) {

	// capability statement
	path := filepath.Join("../testdata", "cerner_capability_dstu2.json")
	csJSON, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
	}
	cs, err := capabilityparser.NewCapabilityStatement(csJSON)
	if err != nil {
		t.Error(err)
	}

	var endpointMetadata1 = &FHIREndpointMetadata{
		URL:          "http://www.example.com",
		HTTPResponse: 200,
		Availability: 1.0,
		Errors:       "Example Error",
		ResponseTime: 0.123456,
	}

	var endpointMetadata2 = &FHIREndpointMetadata{
		URL:          "http://www.example.com",
		HTTPResponse: 200,
		Availability: 1.0,
		Errors:       "Example Error",
		ResponseTime: 0.123456,
	}

	// endpointInfos
	var endpointInfo1 = &FHIREndpointInfo{
		ID:                1,
		URL:               "http://www.example.com",
		HealthITProductID: 3,
		TLSVersion:        "TLS 1.1",
		MIMETypes:         []string{"application/json+fhir", "application/fhir+json"},
		VendorID:          2,
		Validation: Validation{
			Results: []Rule{
				{
					RuleName:  "httpResponse",
					Valid:     false,
					Expected:  "200",
					Actual:    "404",
					Comment:   "Not 200",
					Reference: "reference.com",
					ImplGuide: "Guide",
				},
			},
		},
		IncludedFields:      testIncludedFields,
		SupportedResources:  []string{"AllergyIntolerance", "Binary", "CarePlan"},
		CapabilityStatement: cs,
		Metadata:            endpointMetadata1}
	includedFieldsCopy := make([]IncludedField, len(testIncludedFields))
	copy(includedFieldsCopy, testIncludedFields)
	var endpointInfo2 = &FHIREndpointInfo{
		ID:                1,
		URL:               "http://www.example.com",
		HealthITProductID: 3,
		TLSVersion:        "TLS 1.1",
		MIMETypes:         []string{"application/json+fhir", "application/fhir+json"},
		VendorID:          2,
		Validation: Validation{
			Results: []Rule{
				{
					RuleName:  "httpResponse",
					Valid:     false,
					Expected:  "200",
					Actual:    "404",
					Comment:   "Not 200",
					Reference: "reference.com",
					ImplGuide: "Guide",
				},
			},
		},
		IncludedFields:      includedFieldsCopy,
		SupportedResources:  []string{"AllergyIntolerance", "Binary", "CarePlan"},
		CapabilityStatement: cs,
		Metadata:            endpointMetadata2}

	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expected endpointInfo1 to equal endpointInfo2. They are not equal.")
	}

	endpointInfo2.ID = 2
	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expect endpointInfo 1 to equal endpointInfo 2. ids should be ignored. %d vs %d", endpointInfo1.ID, endpointInfo2.ID)
	}
	endpointInfo2.ID = endpointInfo1.ID

	endpointInfo2.URL = "other"
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expect endpointInfo 1 to not equal endpointInfo 2. URL should be different. %s vs %s", endpointInfo1.URL, endpointInfo2.URL)
	}
	endpointInfo2.URL = endpointInfo1.URL

	endpointInfo2.HealthITProductID = 4
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expect endpointInfo 1 to not equal endpointInfo 2. HealthITProductID should be different. %d vs %d", endpointInfo1.HealthITProductID, endpointInfo2.HealthITProductID)
	}
	endpointInfo2.HealthITProductID = endpointInfo1.HealthITProductID

	endpointInfo2.VendorID = 3
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expect endpointInfo 1 to not equal endpointInfo 2. Vendor should be different. %d vs %d", endpointInfo1.VendorID, endpointInfo2.VendorID)
	}
	endpointInfo2.VendorID = endpointInfo1.VendorID

	endpointInfo2.TLSVersion = "other"
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. TLSVersion should be different. %s vs %s", endpointInfo1.TLSVersion, endpointInfo2.TLSVersion)
	}
	endpointInfo2.TLSVersion = endpointInfo1.TLSVersion

	endpointInfo2.MIMETypes = []string{"other"}
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. MIMETypes should be different. %s vs %s", endpointInfo1.MIMETypes, endpointInfo2.MIMETypes)
	}
	endpointInfo2.MIMETypes = []string{"application/fhir+json", "other"}
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. MIMETypes should be different. %s vs %s", endpointInfo1.MIMETypes, endpointInfo2.MIMETypes)
	}
	endpointInfo2.MIMETypes = []string{"application/fhir+json", "application/json+fhir"}
	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expected endpointInfo1 to equal endpointInfo 2. MIMETypes are same but in different order. %s vs %s", endpointInfo1.MIMETypes, endpointInfo2.MIMETypes)
	}
	endpointInfo2.MIMETypes = endpointInfo1.MIMETypes

	endpointInfo2.Metadata.HTTPResponse = 404
	if endpointInfo2.Equal(endpointInfo1) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. HTTPResponse should be different. %d vs %d", endpointInfo1.Metadata.HTTPResponse, endpointInfo2.Metadata.HTTPResponse)
	}
	endpointInfo2.Metadata.HTTPResponse = endpointInfo1.Metadata.HTTPResponse

	endpointInfo2.Metadata.Availability = 0
	if endpointInfo2.Equal(endpointInfo1) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. Availability should be different. %f vs %f", endpointInfo1.Metadata.Availability, endpointInfo2.Metadata.Availability)
	}
	endpointInfo2.Metadata.Availability = endpointInfo1.Metadata.Availability

	endpointInfo2.Metadata.Errors = "other"
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. Errors should be different. %s vs %s", endpointInfo1.Metadata.Errors, endpointInfo2.Metadata.Errors)
	}
	endpointInfo2.Metadata.Errors = endpointInfo1.Metadata.Errors

	endpointInfo2.CapabilityStatement = nil
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. CapabilityStatement should be different. %s vs %s", endpointInfo1.CapabilityStatement, endpointInfo2.CapabilityStatement)
	}
	endpointInfo2.CapabilityStatement = endpointInfo1.CapabilityStatement

	endpointInfo1.CapabilityStatement = nil
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. CapabilityStatement should be different. %s vs %s", endpointInfo1.CapabilityStatement, endpointInfo2.CapabilityStatement)
	}
	endpointInfo1.CapabilityStatement = endpointInfo2.CapabilityStatement

	endpointInfo2.Validation = Validation{}
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. Validation should be different. %+v vs %+v", endpointInfo1.Validation, endpointInfo2.Validation)
	}
	endpointInfo2.Validation = endpointInfo1.Validation

	endpointInfo1.Validation = Validation{}
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. Validation should be different. %+v vs %+v", endpointInfo1.Validation, endpointInfo2.Validation)
	}
	endpointInfo1.Validation = endpointInfo2.Validation

	endpointInfo1.IncludedFields[0] = IncludedField{
		Field:  "url",
		Exists: false,
	}
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. IncludedFields should be different. %+v vs %+v", endpointInfo1.IncludedFields[0], endpointInfo2.IncludedFields[0])
	}
	endpointInfo1.IncludedFields = endpointInfo2.IncludedFields

	endpointInfo2.IncludedFields = make([]IncludedField, 0)
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. IncludedFields should be different. %+v vs %+v", endpointInfo1.IncludedFields, endpointInfo2.IncludedFields)
	}
	endpointInfo2.IncludedFields = endpointInfo1.IncludedFields

	endpointInfo1.IncludedFields = make([]IncludedField, 0)
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. IncludedFields should be different. %+v vs %+v", endpointInfo1.IncludedFields, endpointInfo2.IncludedFields)
	}
	endpointInfo1.IncludedFields = endpointInfo2.IncludedFields

	endpointInfo2.SupportedResources = []string{"other"}
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. SupportedResources should be different. %s vs %s", endpointInfo1.SupportedResources, endpointInfo2.SupportedResources)
	}
	endpointInfo2.SupportedResources = []string{"AllergyIntolerance", "Binary", "other"}
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. SupportedResources should be different. %s vs %s", endpointInfo1.SupportedResources, endpointInfo2.SupportedResources)
	}
	endpointInfo2.SupportedResources = []string{"Binary", "CarePlan", "AllergyIntolerance"}
	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expected endpointInfo1 to equal endpointInfo 2. SupportedResources are same but in different order. %s vs %s", endpointInfo1.SupportedResources, endpointInfo2.SupportedResources)
	}
	endpointInfo2.SupportedResources = endpointInfo1.SupportedResources

	endpointInfo2.Metadata.ResponseTime = 0.234567
	if endpointInfo2.Equal(endpointInfo1) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. ResponseTime should be different. %f vs %f", endpointInfo1.Metadata.ResponseTime, endpointInfo2.Metadata.ResponseTime)
	}
	endpointInfo2.Metadata.ResponseTime = endpointInfo1.Metadata.ResponseTime

	endpointInfo2 = nil
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal nil endpointInfo 2.")
	}
	endpointInfo2 = endpointInfo1

	endpointInfo1 = nil
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect nil endpointInfo1 to equal endpointInfo 2.")
	}

	endpointInfo2 = nil
	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Nil endpointInfo 1 should equal nil endpointInfo 2.")
	}
}
