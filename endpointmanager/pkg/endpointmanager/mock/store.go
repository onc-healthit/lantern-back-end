package mock

// Store implements the endpointmanager FHIREndpointStore, HealthITProductStore, and ProviderOrganizationStore
// interfaces and allows mock implementations of the associated methods.
// Each Store method calls the corresponding method <methodName>Fn as assigned in the mock Store structure.
// It also assigns <methodName>Invoked to true when <methodName> is called.
type Store struct {
}

// NewStore creates a mock store.
func NewStore() (*Store, error) {
	var store Store

	return &store, nil
}

// Close calls the mocked close function.
func (s *Store) Close() {
}
