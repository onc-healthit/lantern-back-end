package mock

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

var _ endpointmanager.FHIREndpointStore = &Store{}
var _ endpointmanager.HealthITProductStore = &Store{}
var _ endpointmanager.ProviderOrganizationStore = &Store{}

// Store implements the endpointmanager FHIREndpointStore, HealthITProductStore, and ProviderOrganizationStore
// interfaces and allows mock implementations of the associated methods.
// Each Store method calls the corresponding method <methodName>Fn as assigned in the mock Store structure.
// It also assigns <methodName>Invoked to true when <methodName> is called.
type Store struct {
	GetFHIREndpointFn      func(int) (*endpointmanager.FHIREndpoint, error)
	GetFHIREndpointInvoked bool

	GetFHIREndpointUsingURLFn      func(string) (*endpointmanager.FHIREndpoint, error)
	GetFHIREndpointUsingURLInvoked bool

	AddFHIREndpointFn      func(*endpointmanager.FHIREndpoint) error
	AddFHIREndpointInvoked bool

	UpdateFHIREndpointFn      func(*endpointmanager.FHIREndpoint) error
	UpdateFHIREndpointInvoked bool

	DeleteFHIREndpointFn      func(*endpointmanager.FHIREndpoint) error
	DeleteFHIREndpointInvoked bool

	GetHealthITProductFn      func(int) (*endpointmanager.HealthITProduct, error)
	GetHealthITProductInvoked bool

	GetHealthITProductUsingNameAndVersionFn      func(string, string) (*endpointmanager.HealthITProduct, error)
	GetHealthITProductUsingNameAndVersionInvoked bool

	AddHealthITProductFn      func(*endpointmanager.HealthITProduct) error
	AddHealthITProductInvoked bool

	UpdateHealthITProductFn      func(*endpointmanager.HealthITProduct) error
	UpdateHealthITProductInvoked bool

	DeleteHealthITProductFn      func(*endpointmanager.HealthITProduct) error
	DeleteHealthITProductInvoked bool

	GetProviderOrganizationFn      func(int) (*endpointmanager.ProviderOrganization, error)
	GetProviderOrganizationInvoked bool

	AddProviderOrganizationFn      func(*endpointmanager.ProviderOrganization) error
	AddProviderOrganizationInvoked bool

	UpdateProviderOrganizationFn      func(*endpointmanager.ProviderOrganization) error
	UpdateProviderOrganizationInvoked bool

	DeleteProviderOrganizationFn      func(*endpointmanager.ProviderOrganization) error
	DeleteProviderOrganizationInvoked bool

	CloseFn      func()
	CloseInvoked bool
}

// NewStore creates a connection to the postgresql database and adds a reference to the database
// in store.DB.
func NewStore() (*Store, error) {
	var store Store

	return &store, nil
}

// Close closes the postgresql database connection.
func (s *Store) Close() {
	s.CloseInvoked = true
	s.CloseFn()
}
