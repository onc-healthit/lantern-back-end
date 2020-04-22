package populatefhirendpoints

import (
	"testing"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"
)

var testEndpointEntry fetcher.EndpointEntry = fetcher.EndpointEntry{
	OrganizationName:     "A Woman's Place",
	FHIRPatientFacingURI: "https://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2/",
	ListSource:           "CareEvolution",
}

var testFHIREndpoint endpointmanager.FHIREndpoint = endpointmanager.FHIREndpoint{
	OrganizationName: "A Woman's Place",
	URL:              "https://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2/",
	ListSource:       "CareEvolution",
}

func Test_formatToFHIREndpt(t *testing.T) {
	endpt := testEndpointEntry
	expectedFHIREndpt := testFHIREndpoint

	// basic test

	fhirEndpt, err := formatToFHIREndpt(&endpt)
	th.Assert(t, err == nil, err)
	th.Assert(t, fhirEndpt.Equal(&expectedFHIREndpt), "EndpointEntry did not get parsed into a FHIREndpoint as expected")

	// test that a trailing '/' is added to the URL
	endpt.FHIRPatientFacingURI = "https://fhir-myrecord.cerner.com/dstu2/sqiH60CNKO9o0PByEO9XAxX0dZX5s5b2"
	fhirEndpt, err = formatToFHIREndpt(&endpt)
	th.Assert(t, err == nil, err)
	th.Assert(t, fhirEndpt.Equal(&expectedFHIREndpt), "EndpointEntry did not get parsed into a FHIREndpoint as expected")
}
