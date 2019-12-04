package mock

import (
	"context"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// GetProviderOrganization mocks endpointmanager.ProviderOrganizationStore.GetProviderOrganization and sets s.GetProviderOrganizationInvoked to true and calls s.GetProviderOrganizationFn with the given arguments.
func (s *Store) GetProviderOrganization(ctx context.Context, id int) (*endpointmanager.ProviderOrganization, error) {
	s.GetProviderOrganizationInvoked = true
	return s.GetProviderOrganizationFn(ctx, id)
}

// AddProviderOrganization mocks endpointmanager.ProviderOrganizationStore.AddProviderOrganization and sets s.AddProviderOrganizationInvoked to true and calls s.AddProviderOrganizationFn with the given arguments.
func (s *Store) AddProviderOrganization(ctx context.Context, po *endpointmanager.ProviderOrganization) error {
	s.AddProviderOrganizationInvoked = true
	return s.AddProviderOrganizationFn(ctx, po)
}

// UpdateProviderOrganization mocks endpointmanager.ProviderOrganizationStore.UpdateProviderOrganization and sets s.UpdateProviderOrganizationInvoked to true and calls s.UpdateProviderOrganizationFn with the given arguments.
func (s *Store) UpdateProviderOrganization(ctx context.Context, po *endpointmanager.ProviderOrganization) error {
	s.UpdateProviderOrganizationInvoked = true
	return s.UpdateProviderOrganizationFn(ctx, po)
}

// DeleteProviderOrganization mocks endpointmanager.ProviderOrganizationStore.DeleteProviderOrganization and sets s.DeleteProviderOrganizationInvoked to true and calls s.DeleteProviderOrganizationFn with the given arguments.
func (s *Store) DeleteProviderOrganization(ctx context.Context, po *endpointmanager.ProviderOrganization) error {
	s.DeleteProviderOrganizationInvoked = true
	return s.DeleteProviderOrganizationFn(ctx, po)
}
