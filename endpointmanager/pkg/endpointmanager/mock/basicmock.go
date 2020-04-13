package mock

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// BasicMockStore is a mocked store that includes both the HealthItProduct and FHIREndpoint stores
type BasicMockStore struct {
	HealthITProductData []*endpointmanager.HealthITProduct
	FhirEndpointData    []*endpointmanager.FHIREndpoint
	Store
}

func newBasicMockStore() *BasicMockStore {
	store := BasicMockStore{}

	return &store
}
