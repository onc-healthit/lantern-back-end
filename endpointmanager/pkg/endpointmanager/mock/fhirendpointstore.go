package mock

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// GetFHIREndpoint mocks endpointmanager.FHIREndpointStore.GetFHIREndpoint and sets s.GetFHIREndpointInvoked to true and calls s.GetFHIREndpointFn with the given arguments.
func (s *Store) GetFHIREndpoint(id int) (*endpointmanager.FHIREndpoint, error) {
	s.GetFHIREndpointInvoked = true
	return s.GetFHIREndpointFn(id)
}

// GetFHIREndpointUsingURL mocks endpointmanager.FHIREndpointStore.GetFHIREndpointUsingURL and sets s.GetFHIREndpointUsingURLInvoked to true and calls s.GetFHIREndpointUsingURLFn with the given arguments.
func (s *Store) GetFHIREndpointUsingURL(url string) (*endpointmanager.FHIREndpoint, error) {
	s.GetFHIREndpointUsingURLInvoked = true
	return s.GetFHIREndpointUsingURLFn(url)
}

// AddFHIREndpoint mocks endpointmanager.FHIREndpointStore.AddFHIREndpoint and sets s.AddFHIREndpointInvoked to true and calls s.AddFHIREndpointFn with the given arguments.
func (s *Store) AddFHIREndpoint(e *endpointmanager.FHIREndpoint) error {
	s.AddFHIREndpointInvoked = true
	return s.AddFHIREndpointFn(e)
}

// UpdateFHIREndpoint mocks endpointmanager.FHIREndpointStore.UpdateFHIREndpoint and sets s.UpdateFHIREndpointInvoked to true and calls s.UpdateFHIREndpointFn with the given arguments.
func (s *Store) UpdateFHIREndpoint(e *endpointmanager.FHIREndpoint) error {
	s.UpdateFHIREndpointInvoked = true
	return s.UpdateFHIREndpointFn(e)
}

// DeleteFHIREndpoint mocks endpointmanager.FHIREndpointStore.DeleteFHIREndpoint and sets s.DeleteFHIREndpointInvoked to true and calls s.DeleteFHIREndpointFn with the given arguments.
func (s *Store) DeleteFHIREndpoint(e *endpointmanager.FHIREndpoint) error {
	s.DeleteFHIREndpointInvoked = true
	return s.DeleteFHIREndpointFn(e)
}
