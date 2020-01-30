package mock

import (
	"context"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

var _ endpointmanager.FHIREndpointStore = &Store{}
var _ endpointmanager.HealthITProductStore = &Store{}

// Store implements the endpointmanager FHIREndpointStore, HealthITProductStore, and ProviderOrganizationStore
// interfaces and allows mock implementations of the associated methods.
// Each Store method calls the corresponding method <methodName>Fn as assigned in the mock Store structure.
// It also assigns <methodName>Invoked to true when <methodName> is called.
type Store struct {
	GetFHIREndpointFn      func(context.Context, int) (*endpointmanager.FHIREndpoint, error)
	GetFHIREndpointInvoked bool

	GetFHIREndpointUsingURLFn      func(context.Context, string) (*endpointmanager.FHIREndpoint, error)
	GetFHIREndpointUsingURLInvoked bool

	AddFHIREndpointFn      func(context.Context, *endpointmanager.FHIREndpoint) error
	AddFHIREndpointInvoked bool

	UpdateFHIREndpointFn      func(context.Context, *endpointmanager.FHIREndpoint) error
	UpdateFHIREndpointInvoked bool

	DeleteFHIREndpointFn      func(context.Context, *endpointmanager.FHIREndpoint) error
	DeleteFHIREndpointInvoked bool

	GetHealthITProductFn      func(context.Context, int) (*endpointmanager.HealthITProduct, error)
	GetHealthITProductInvoked bool

	GetHealthITProductUsingNameAndVersionFn      func(context.Context, string, string) (*endpointmanager.HealthITProduct, error)
	GetHealthITProductUsingNameAndVersionInvoked bool

	GetHealthITProductsUsingVendorFn      func(context.Context, string) ([]*endpointmanager.HealthITProduct, error)
	GetHealthITProductsUsingVendorInvoked bool

	GetHealthITProductDevelopersFn      func(context.Context) ([]string, error)
	GetHealthITProductDevelopersInvoked bool

	AddHealthITProductFn      func(context.Context, *endpointmanager.HealthITProduct) error
	AddHealthITProductInvoked bool

	UpdateHealthITProductFn      func(context.Context, *endpointmanager.HealthITProduct) error
	UpdateHealthITProductInvoked bool

	DeleteHealthITProductFn      func(context.Context, *endpointmanager.HealthITProduct) error
	DeleteHealthITProductInvoked bool

	CloseFn      func()
	CloseInvoked bool
}

// NewStore creates a mock store.
func NewStore() (*Store, error) {
	var store Store

	return &store, nil
}

// Close calls the mocked close function.
func (s *Store) Close() {
	s.CloseInvoked = true
	s.CloseFn()
}
