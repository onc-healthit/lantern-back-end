package populatefhirendpoints

import (
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fetcher"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
)


var testEndpointEntry fetcher.EndpointEntry = fetcher.EndpointEntry{
	OrganizationName:     "AdvantageCare Physicians",
	FHIRPatientFacingURI: "https://epwebapps.acpny.com/FHIRproxy/api/FHIR/DSTU2/",
	ListSource:           "epicList",
}

var epOrg = &endpointmanager.FHIREndpointOrganization{
	OrganizationName: "AdvantageCare Physicians"}

var testFHIREndpoint endpointmanager.FHIREndpoint = endpointmanager.FHIREndpoint{
	OrganizationList:  []*endpointmanager.FHIREndpointOrganization{epOrg},
	URL:               "https://epwebapps.acpny.com/FHIRproxy/api/FHIR/DSTU2/",
	ListSource:        "epicList",
}

func Test_formatToFHIREndpt(t *testing.T) {
	endpt := testEndpointEntry
	expectedFHIREndpt := testFHIREndpoint

	// basic test

	fhirEndpt, err := formatToFHIREndpt(&endpt)
	th.Assert(t, err == nil, err)
	th.Assert(t, fhirEndpt.Equal(&expectedFHIREndpt), "EndpointEntry did not get parsed into a FHIREndpoint as expected")

	// test that a trailing '/' is added to the URL
	endpt.FHIRPatientFacingURI = "https://epwebapps.acpny.com/FHIRproxy/api/FHIR/DSTU2/"
	fhirEndpt, err = formatToFHIREndpt(&endpt)
	th.Assert(t, err == nil, err)
	th.Assert(t, fhirEndpt.Equal(&expectedFHIREndpt), "EndpointEntry did not get parsed into a FHIREndpoint as expected")
}
