package populatefhirendpoints

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/mock"
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

func Test_saveEndpointData(t *testing.T) {
	var err error
	store := mock.NewBasicMockFhirEndpointStore()

	endpt := testEndpointEntry
	fhirEndpt := testFHIREndpoint

	// check that nothing is stored and that saveEndpointData throws an error if the context is canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = saveEndpointData(ctx, store, &endpt)
	th.Assert(t, len(store.(*mock.BasicMockStore).FhirEndpointData) == 0, "should not have stored data")
	th.Assert(t, errors.Cause(err) == context.Canceled, "should have errored out with root cause that the context was canceled")

	// reset context
	ctx = context.Background()

	// check that new item is stored
	err = saveEndpointData(ctx, store, &endpt)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(store.(*mock.BasicMockStore).FhirEndpointData) == 1, "did not store data as expected")
	th.Assert(t, fhirEndpt.Equal(store.(*mock.BasicMockStore).FhirEndpointData[0]), "stored data does not equal expected store data")

	// check that an item with the same URL does not replace item
	endpt.OrganizationName = "A Woman's Place 2"
	err = saveEndpointData(ctx, store, &endpt)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(store.(*mock.BasicMockStore).FhirEndpointData) == 1, "did not store data as expected")
	th.Assert(t, fhirEndpt.Equal(store.(*mock.BasicMockStore).FhirEndpointData[0]), "stored data does not equal expected store data")

	// check that error adding to store throws error
	endpt = testEndpointEntry
	endpt.FHIRPatientFacingURI = "http://a-new-url.com/metadata/"
	addFn := store.(*mock.BasicMockStore).AddFHIREndpointFn
	store.(*mock.BasicMockStore).AddFHIREndpointFn = func(_ context.Context, _ *endpointmanager.FHIREndpoint) error {
		return errors.New("add fhir endpoint test error")
	}
	err = saveEndpointData(ctx, store, &endpt)
	th.Assert(t, errors.Cause(err).Error() == "add fhir endpoint test error", "expected error adding product")
	store.(*mock.BasicMockStore).AddFHIREndpointFn = addFn
}

func Test_AddEndpointData(t *testing.T) {
	var err error
	store := mock.NewBasicMockFhirEndpointStore()

	endpt1 := testEndpointEntry
	endpt2 := testEndpointEntry
	endpt2.FHIRPatientFacingURI = "http://a-new-url.com/metadata/"
	listEndpoints := fetcher.ListOfEndpoints{Entries: []fetcher.EndpointEntry{endpt1, endpt2}}
	expectedEndptsStored := 2

	// check that nothing is stored and that AddEndpointData throws an error if the context is canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = AddEndpointData(ctx, store, &listEndpoints)
	th.Assert(t, len(store.(*mock.BasicMockStore).FhirEndpointData) == 0, "should not have stored data")
	th.Assert(t, errors.Cause(err) == context.Canceled, "should have errored out with root cause that the context was canceled")

	// reset context
	ctx = context.Background()

	// check basic functionality
	err = AddEndpointData(ctx, store, &listEndpoints)
	th.Assert(t, err == nil, err)
	actualEndptsStored := len(store.(*mock.BasicMockStore).FhirEndpointData)
	th.Assert(t, actualEndptsStored == expectedEndptsStored, fmt.Sprintf("Expected %d products stored. Actually had %d products stored.", expectedEndptsStored, actualEndptsStored))
	th.Assert(t, store.(*mock.BasicMockStore).FhirEndpointData[0].URL == endpt1.FHIRPatientFacingURI, "Did not store first product as expected")
	th.Assert(t, store.(*mock.BasicMockStore).FhirEndpointData[1].URL == endpt2.FHIRPatientFacingURI, "Did not store second product as expected")

	// reset values
	store = mock.NewBasicMockFhirEndpointStore()
	endpt2 = testEndpointEntry
	endpt2.OrganizationName = "New Name"
	listEndpoints = fetcher.ListOfEndpoints{Entries: []fetcher.EndpointEntry{endpt1, endpt2}}
	err = AddEndpointData(ctx, store, &listEndpoints)
	th.Assert(t, err == nil, err)
	th.Assert(t, len(store.(*mock.BasicMockStore).FhirEndpointData) == 1, "did not persist one product as expected")
	th.Assert(t, store.(*mock.BasicMockStore).FhirEndpointData[0].OrganizationName == endpt1.OrganizationName, "stored data does not equal expected store data")
}
