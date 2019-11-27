package mock

import (
	"context"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// GetFHIREndpoint mocks endpointmanager.FHIREndpointStore.GetFHIREndpoint and sets s.GetFHIREndpointInvoked to true and calls s.GetFHIREndpointFn with the given arguments.
func (s *Store) GetFHIREndpoint(ctx context.Context, id int) (*endpointmanager.FHIREndpoint, error) {
	s.GetFHIREndpointInvoked = true
	return s.GetFHIREndpointFn(ctx, id)
}

// GetFHIREndpointUsingURL mocks endpointmanager.FHIREndpointStore.GetFHIREndpointUsingURL and sets s.GetFHIREndpointUsingURLInvoked to true and calls s.GetFHIREndpointUsingURLFn with the given arguments.
func (s *Store) GetFHIREndpointUsingURL(ctx context.Context, url string) (*endpointmanager.FHIREndpoint, error) {
	s.GetFHIREndpointUsingURLInvoked = true
	return s.GetFHIREndpointUsingURLFn(ctx, url)
}

// AddFHIREndpoint mocks endpointmanager.FHIREndpointStore.AddFHIREndpoint and sets s.AddFHIREndpointInvoked to true and calls s.AddFHIREndpointFn with the given arguments.
func (s *Store) AddFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	s.AddFHIREndpointInvoked = true
	return s.AddFHIREndpointFn(ctx, e)
}

// UpdateFHIREndpoint mocks endpointmanager.FHIREndpointStore.UpdateFHIREndpoint and sets s.UpdateFHIREndpointInvoked to true and calls s.UpdateFHIREndpointFn with the given arguments.
func (s *Store) UpdateFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	s.UpdateFHIREndpointInvoked = true
	return s.UpdateFHIREndpointFn(ctx, e)
}

// DeleteFHIREndpoint mocks endpointmanager.FHIREndpointStore.DeleteFHIREndpoint and sets s.DeleteFHIREndpointInvoked to true and calls s.DeleteFHIREndpointFn with the given arguments.
func (s *Store) DeleteFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	s.DeleteFHIREndpointInvoked = true
	return s.DeleteFHIREndpointFn(ctx, e)
}
