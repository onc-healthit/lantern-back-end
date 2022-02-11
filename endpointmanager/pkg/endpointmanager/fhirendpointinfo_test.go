package endpointmanager

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
)

var testSupportedProfiles = []SupportedProfile{
	{
		Resource:    "",
		ProfileURL:  "http://hl7.org/fhir/StructureDefinition/daf-allergyintolerance",
		ProfileName: "U.S. Data Access Framework (DAF) AllergyIntolerance Profile",
	},
	{
		Resource:    "",
		ProfileURL:  "http://hl7.org/fhir/StructureDefinition/daf-condition",
		ProfileName: "U.S. Data Access Framework (DAF) Condition Profile",
	},
	{
		Resource:    "",
		ProfileURL:  "http://hl7.org/fhir/StructureDefinition/daf-diagnosticorder",
		ProfileName: "U.S. Data Access Framework (DAF) DiagnosticOrder Profile",
	},
	{
		Resource:    "",
		ProfileURL:  "http://hl7.org/fhir/StructureDefinition/daf-diagnosticreport",
		ProfileName: "U.S. Data Access Framework (DAF) DiagnosticReport Profile",
	},
	{
		Resource:    "",
		ProfileURL:  "http://hl7.org/fhir/StructureDefinition/daf-immunization",
		ProfileName: "U.S. Data Access Framework (DAF) Immunization Profile",
	},
	{
		Resource:    "",
		ProfileURL:  "http://hl7.org/fhir/StructureDefinition/daf-medicationorder",
		ProfileName: "U.S. Data Access Framework (DAF) MedicationOrder Profile",
	},
	{
		Resource:    "",
		ProfileURL:  "http://hl7.org/fhir/StructureDefinition/daf-medicationstatement",
		ProfileName: "U.S. Data Access Framework (DAF) MedicationStatement Profile",
	},
	{
		Resource:    "",
		ProfileURL:  "http://hl7.org/fhir/StructureDefinition/daf-patient",
		ProfileName: "U.S. Data Access Framework (DAF) Patient Profile",
	},
	{
		Resource:    "",
		ProfileURL:  "http://hl7.org/fhir/StuctureDefinition/daf-procedure",
		ProfileName: "U.S. Data Access Framework (DAF) Procedure Profile",
	},
	{
		Resource:    "",
		ProfileURL:  "http://hl7.org/fhir/StructureDefinition/daf-resultobs",
		ProfileName: "U.S. Data Access Framework (DAF) Results Profile",
	},
	{
		Resource:    "",
		ProfileURL:  "http://hl7.org/fhir/StructureDefinition/daf-smokingstatus",
		ProfileName: "U.S. Data Access Framework (DAF) SmokingStatus Profile",
	},
	{
		Resource:    "",
		ProfileURL:  "http://hl7.org/fhir/StructureDefinition/daf-vitalsigns",
		ProfileName: "U.S. Data Access Framework (DAF) VitalSigns Profile",
	},
}

var testIncludedFields = []IncludedField{
	{
		Field:     "url",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "version",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "name",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "status",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "experimental",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "publisher",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "contact",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "date",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "description",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "copyright",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "kind",
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
		Field:     "software.releaseDate",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "implementation.description",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "implementation.url",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "fhirVersion",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "format",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "rest.mode",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "rest.resource.type",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "rest.resource.interaction.code",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "rest.resource.versioning",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "rest.resource.conditionalDelete",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "rest.resource.searchParam.type",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "rest.interaction.code",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "document",
		Exists:    false,
		Extension: false,
	},
	{
		Field:     "document.mode",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "messaging",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "requirements",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "profile",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "acceptUnknown",
		Exists:    true,
		Extension: false,
	},
	{
		Field:     "conformance-supported-system",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "conformance-search-parameter-combination",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "conformance-expectation",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "conformance-prohibited",
		Exists:    false,
		Extension: true,
	},
	{
		Field:     "DSTU2-oauth-uris",
		Exists:    false,
		Extension: true,
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
		URL:                  "http://www.example.com",
		HTTPResponse:         200,
		Availability:         1.0,
		Errors:               "Example Error",
		ResponseTime:         0.123456,
		SMARTHTTPResponse:    200,
		RequestedFhirVersion: "None",
	}

	var endpointMetadata2 = &FHIREndpointMetadata{
		URL:                  "http://www.example.com",
		HTTPResponse:         200,
		Availability:         1.0,
		Errors:               "Example Error",
		ResponseTime:         0.123456,
		SMARTHTTPResponse:    200,
		RequestedFhirVersion: "None",
	}

	// endpointInfos
	var endpointInfo1 = &FHIREndpointInfo{
		ID:                1,
		URL:               "http://www.example.com",
		HealthITProductID: 3,
		TLSVersion:        "TLS 1.1",
		MIMETypes:         []string{"application/json+fhir", "application/fhir+json"},
		VendorID:          2,
		ValidationID:      1,
		IncludedFields:    testIncludedFields,
		OperationResource: map[string][]string{
			"read": {"AllergyIntolerance", "Binary", "CarePlan"}},
		CapabilityStatement:   cs,
		RequestedFhirVersion:  "None",
		CapabilityFhirVersion: "1.0.2",
		Metadata:              endpointMetadata1,
		SupportedProfiles:     testSupportedProfiles}

	includedFieldsCopy := make([]IncludedField, len(testIncludedFields))
	copy(includedFieldsCopy, testIncludedFields)

	supportedProfilesCopy := make([]SupportedProfile, len(testSupportedProfiles))
	copy(supportedProfilesCopy, testSupportedProfiles)

	var endpointInfo2 = &FHIREndpointInfo{
		ID:                1,
		URL:               "http://www.example.com",
		HealthITProductID: 3,
		TLSVersion:        "TLS 1.1",
		MIMETypes:         []string{"application/json+fhir", "application/fhir+json"},
		VendorID:          2,
		ValidationID:      1,
		IncludedFields:    includedFieldsCopy,
		OperationResource: map[string][]string{
			"read": {"AllergyIntolerance", "Binary", "CarePlan"}},
		CapabilityStatement:   cs,
		RequestedFhirVersion:  "None",
		CapabilityFhirVersion: "1.0.2",
		Metadata:              endpointMetadata2,
		SupportedProfiles:     supportedProfilesCopy}

	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expected endpointInfo1 to equal endpointInfo2. They are not equal.")
	}

	endpointInfo2.ID = 2
	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expect endpointInfo 1 to equal endpointInfo 2. ids should be ignored. %d vs %d", endpointInfo1.ID, endpointInfo2.ID)
	}
	endpointInfo2.ID = endpointInfo1.ID

	endpointInfo2.CapabilityFhirVersion = "3.0.2"
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expect endpointInfo 1 to not equal endpointInfo 2. capability fhir versions should be different. %s vs %s", endpointInfo1.CapabilityFhirVersion, endpointInfo2.CapabilityFhirVersion)
	}
	endpointInfo2.CapabilityFhirVersion = endpointInfo1.CapabilityFhirVersion

	endpointInfo2.RequestedFhirVersion = "3.0.2"
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expect endpointInfo 1 to not equal endpointInfo 2. requested fhir versions should be different. %s vs %s", endpointInfo1.RequestedFhirVersion, endpointInfo2.RequestedFhirVersion)
	}
	endpointInfo2.RequestedFhirVersion = endpointInfo1.RequestedFhirVersion

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

	endpointInfo2.ValidationID = 4
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expect endpointInfo 1 to not equal endpointInfo 2. ValidationID should be different. %d vs %d", endpointInfo1.ValidationID, endpointInfo2.ValidationID)
	}
	endpointInfo2.ValidationID = endpointInfo1.ValidationID

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
	if !endpointInfo1.EqualExcludeMetadata(endpointInfo2) {
		t.Errorf("Expect endpointInfo1 to equal endpointInfo2 when excluding Metadata. Metadata HTTP responses should be ignored. %d vs %d", endpointInfo1.Metadata.HTTPResponse, endpointInfo1.Metadata.HTTPResponse)
	}
	endpointInfo2.Metadata.HTTPResponse = endpointInfo1.Metadata.HTTPResponse

	endpointInfo2.Metadata.SMARTHTTPResponse = 0
	if endpointInfo2.Equal(endpointInfo1) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. SMARTHTTPResponse should be different. %d vs %d", endpointInfo1.Metadata.SMARTHTTPResponse, endpointInfo2.Metadata.SMARTHTTPResponse)
	}
	if !endpointInfo1.EqualExcludeMetadata(endpointInfo2) {
		t.Errorf("Expect endpointInfo1 to equal endpointInfo2 when excluding Metadata. Metadata Smart HTTP responses should be ignored. %d vs %d", endpointInfo1.Metadata.SMARTHTTPResponse, endpointInfo1.Metadata.SMARTHTTPResponse)
	}
	endpointInfo2.Metadata.SMARTHTTPResponse = endpointMetadata1.SMARTHTTPResponse

	endpointInfo2.Metadata.Availability = 0
	if endpointInfo2.Equal(endpointInfo1) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. Availability should be different. %f vs %f", endpointInfo1.Metadata.Availability, endpointInfo2.Metadata.Availability)
	}
	if !endpointInfo1.EqualExcludeMetadata(endpointInfo2) {
		t.Errorf("Expect endpointInfo1 to equal endpointInfo2 when excluding Metadata. Metadata Availability should be ignored. %f vs %f", endpointInfo1.Metadata.Availability, endpointInfo2.Metadata.Availability)
	}
	endpointInfo2.Metadata.Availability = endpointInfo1.Metadata.Availability

	endpointInfo2.Metadata.Errors = "other"
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. Errors should be different. %s vs %s", endpointInfo1.Metadata.Errors, endpointInfo2.Metadata.Errors)
	}
	if !endpointInfo1.EqualExcludeMetadata(endpointInfo2) {
		t.Errorf("Expect endpointInfo1 to equal endpointInfo2 when excluding Metadata. Metadata Errors should be ignored. %s vs %s", endpointInfo1.Metadata.Errors, endpointInfo2.Metadata.Errors)
	}
	endpointInfo2.Metadata.Errors = endpointInfo1.Metadata.Errors

	endpointInfo2.Metadata.ResponseTime = 0.234567
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. Response times should be different. %f vs %f", endpointInfo1.Metadata.ResponseTime, endpointInfo2.Metadata.ResponseTime)
	}
	if !endpointInfo1.EqualExcludeMetadata(endpointInfo2) {
		t.Errorf("Expect endpointInfo1 to equal endpointInfo2 when excluding Metadata. Metadata Response time should be ignored. %f vs %f", endpointInfo1.Metadata.ResponseTime, endpointInfo2.Metadata.ResponseTime)
	}
	endpointInfo2.Metadata.ResponseTime = endpointInfo1.Metadata.ResponseTime

	endpointInfo2.Metadata.URL = "other"
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. Metadata URL should be different. %s vs %s", endpointInfo1.Metadata.URL, endpointInfo2.Metadata.URL)
	}
	if !endpointInfo1.EqualExcludeMetadata(endpointInfo2) {
		t.Errorf("Expect endpointInfo1 to equal endpointInfo2 when excluding Metadata. Metadata URL should be ignored. %s vs %s", endpointInfo1.Metadata.URL, endpointInfo2.Metadata.URL)
	}
	endpointInfo2.Metadata.URL = endpointMetadata1.URL

	endpointInfo2.Metadata.RequestedFhirVersion = "other"
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. Metadata RequestedFhirVersion should be different. %s vs %s", endpointInfo1.Metadata.RequestedFhirVersion, endpointInfo2.Metadata.RequestedFhirVersion)
	}
	if !endpointInfo1.EqualExcludeMetadata(endpointInfo2) {
		t.Errorf("Expect endpointInfo1 to equal endpointInfo2 when excluding Metadata. Metadata RequestedFhirVersion should be ignored. %s vs %s", endpointInfo1.Metadata.RequestedFhirVersion, endpointInfo2.Metadata.RequestedFhirVersion)
	}
	endpointInfo2.Metadata.RequestedFhirVersion = endpointMetadata1.RequestedFhirVersion

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

	endpointInfo2.OperationResource = map[string][]string{"write": {"AllergyIntolerance", "Binary", "CarePlan"}}
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. OperationResource should be different. %s vs %s", endpointInfo1.OperationResource, endpointInfo2.OperationResource)
	}
	endpointInfo2.OperationResource = map[string][]string{"read": {"AllergyIntolerance", "Binary", "other"}}
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. OperationResource should be different. %s vs %s", endpointInfo1.OperationResource, endpointInfo2.OperationResource)
	}
	endpointInfo2.OperationResource = map[string][]string{"read": {"Binary", "AllergyIntolerance", "CarePlan"}}
	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expected endpointInfo1 to equal endpointInfo 2. OperationResource are same but in different order. %s vs %s", endpointInfo1.OperationResource, endpointInfo2.OperationResource)
	}
	endpointInfo2.OperationResource = endpointInfo1.OperationResource

	endpointInfo1.SupportedProfiles[0].ProfileName = "Wrong Profile Name"

	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. SupportedProfiles should be different. %+v vs %+v", endpointInfo1.SupportedProfiles[0], endpointInfo2.SupportedProfiles[0])
	}

	endpointInfo1.SupportedProfiles[0].ProfileName = "U.S. Data Access Framework (DAF) AllergyIntolerance Profile"

	endpointInfo2.SupportedProfiles = make([]SupportedProfile, 0)
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. SupportedProfiles should be different. %+v vs %+v", endpointInfo1.SupportedProfiles, endpointInfo2.SupportedProfiles)
	}
	endpointInfo2.SupportedProfiles = endpointInfo1.SupportedProfiles

	endpointInfo1.SupportedProfiles = make([]SupportedProfile, 0)
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2. SupportedProfiles should be different. %+v vs %+v", endpointInfo1.SupportedProfiles, endpointInfo2.SupportedProfiles)
	}

	endpointInfo1.SupportedProfiles = testSupportedProfiles
	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Expected endpointInfo1 to equal endpointInfo 2. SupportedProfiles should be the same. %+v vs %+v", endpointInfo1.SupportedProfiles, endpointInfo2.SupportedProfiles)
	}
	endpointInfo1.SupportedProfiles = endpointInfo2.SupportedProfiles

	endpointInfo2.Metadata = nil
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 to equal endpointInfo 2 with nil Metadata.")
	}
	if !endpointInfo1.EqualExcludeMetadata(endpointInfo2) {
		t.Errorf("Expect endpointInfo1 to equal endpointInfo2 when excluding Metadata.")
	}
	endpointInfo2.Metadata = endpointMetadata1

	endpointInfo1.Metadata = nil
	if endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did not expect endpointInfo1 with nil Metadata to equal endpointInfo 2.")
	}
	if !endpointInfo1.EqualExcludeMetadata(endpointInfo2) {
		t.Errorf("Expect endpointInfo1 to equal endpointInfo2 when excluding Metadata.")
	}

	endpointInfo2.Metadata = nil
	if !endpointInfo1.Equal(endpointInfo2) {
		t.Errorf("Did expect endpointInfo1 with nil Metadata to equal endpointInfo 2 with nil Metadata .")
	}
	if !endpointInfo1.EqualExcludeMetadata(endpointInfo2) {
		t.Errorf("Expect endpointInfo1 to equal endpointInfo2 when excluding Metadata.")
	}

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
